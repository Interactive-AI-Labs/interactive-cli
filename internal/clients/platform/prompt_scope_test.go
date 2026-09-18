package platform

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func newScopeTestClient(t *testing.T, handler http.HandlerFunc) (*APIClient, func()) {
	t.Helper()
	server := httptest.NewServer(handler)
	client, err := NewAPIClient(server.URL, 5*time.Second, "fake-token", "", nil)
	if err != nil {
		server.Close()
		t.Fatalf("NewAPIClient() error = %v", err)
	}
	return client, server.Close
}

// The global scope is served by the project's own routes under ?scope=global, so
// the query param reaching the wire is the whole contract. Getting it wrong returns
// the project's rows instead of the shared ones, with no error to notice.
func TestListPromptsSendsScope(t *testing.T) {
	tests := []struct {
		name      string
		opts      PromptListOptions
		wantScope string
	}{
		{
			name:      "global scope is sent",
			opts:      PromptListOptions{Scope: ScopeGlobal},
			wantScope: "global",
		},
		{name: "no scope means the project's own", opts: PromptListOptions{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath, gotScope string
			client, closeServer := newScopeTestClient(
				t,
				func(w http.ResponseWriter, r *http.Request) {
					gotPath = r.URL.Path
					gotScope = r.URL.Query().Get("scope")
					_, _ = io.WriteString(
						w,
						`{"success":true,"data":{"prompts":[{"name":"routines"}],"totalCount":1}}`,
					)
				},
			)
			defer closeServer()

			result, err := client.ListPrompts(context.Background(), "proj-1", "skills", tt.opts)
			if err != nil {
				t.Fatalf("ListPrompts() error = %v", err)
			}
			if want := "/api/platform/v1/projects/proj-1/prompts/skills"; gotPath != want {
				t.Errorf("path = %q, want %q", gotPath, want)
			}
			if gotScope != tt.wantScope {
				t.Errorf("scope = %q, want %q", gotScope, tt.wantScope)
			}
			want := []PromptInfo{{Name: "routines"}}
			if diff := cmp.Diff(want, result.Prompts); diff != "" {
				t.Errorf("prompts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetPromptSendsScope(t *testing.T) {
	var gotPath, gotRawPath, gotScope, gotLabel string
	client, closeServer := newScopeTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRawPath = r.URL.EscapedPath()
		gotScope = r.URL.Query().Get("scope")
		gotLabel = r.URL.Query().Get("label")
		_, _ = io.WriteString(
			w,
			`{"success":true,"data":{"name":"team/deploy","prompt":"# Deploy",`+
				`"config":{"skill":{"description":"Deploy"}},"labels":["active"],"tags":[]}}`,
		)
	})
	defer closeServer()

	// A nested name must reach the server intact. The separator goes out
	// percent-encoded, and the server decodes it before matching {name:path} —
	// so assert the encoded form on the wire as well as the decoded one.
	result, err := client.GetPrompt(
		context.Background(), "proj-1", "skills", "team/deploy", 0, "", ScopeGlobal,
	)
	if err != nil {
		t.Fatalf("GetPrompt() error = %v", err)
	}
	if want := "/api/platform/v1/projects/proj-1/prompts/skills/team/deploy"; gotPath != want {
		t.Errorf("decoded path = %q, want %q", gotPath, want)
	}
	if want := "/api/platform/v1/projects/proj-1/prompts/skills/team%2Fdeploy"; gotRawPath != want {
		t.Errorf("wire path = %q, want %q", gotRawPath, want)
	}
	if gotScope != "global" {
		t.Errorf("scope = %q, want %q", gotScope, "global")
	}
	if gotLabel != "" {
		t.Errorf("label = %q, want it unset: a global record has one version", gotLabel)
	}
	if result.Name != "team/deploy" {
		t.Errorf("name = %q, want %q", result.Name, "team/deploy")
	}
	if result.Version != 0 {
		t.Errorf(
			"version = %d, want 0: the shared project's counter is not exposed",
			result.Version,
		)
	}
}

// A missing name has to arrive as NotFoundError, or `get` cannot tell "try the
// shared scope" apart from "the server is down" and would mask real failures.
func TestGetPromptNotFoundIsTyped(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		body         string
		wantNotFound bool
		wantMessage  string
	}{
		{
			name:         "404 with a server message",
			status:       http.StatusNotFound,
			body:         `{"success":false,"error":{"message":"No prompt named 'nope'"}}`,
			wantNotFound: true,
			wantMessage:  "No prompt named 'nope'",
		},
		{
			// ExtractServerMessage falls through to the raw body, so an unparsable
			// error still reaches the user rather than being swallowed.
			name:         "404 with no parsable body",
			status:       http.StatusNotFound,
			body:         `not json`,
			wantNotFound: true,
			wantMessage:  "not json",
		},
		{
			name:         "404 with an empty body still says what happened",
			status:       http.StatusNotFound,
			wantNotFound: true,
			wantMessage:  "failed to get prompt: server returned 404 Not Found",
		},
		{
			name:        "500 is not a not-found",
			status:      http.StatusInternalServerError,
			body:        `{"success":false,"error":{"message":"boom"}}`,
			wantMessage: "boom",
		},
		{
			name:        "403 is not a not-found",
			status:      http.StatusForbidden,
			body:        `{"success":false,"error":{"message":"forbidden"}}`,
			wantMessage: "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, closeServer := newScopeTestClient(
				t,
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
					_, _ = io.WriteString(w, tt.body)
				},
			)
			defer closeServer()

			_, err := client.GetPrompt(context.Background(), "proj-1", "skills", "nope", 0, "", "")
			if err == nil {
				t.Fatal("GetPrompt() error = nil, want an error")
			}
			var notFound *NotFoundError
			if gotNotFound := errors.As(err, &notFound); gotNotFound != tt.wantNotFound {
				t.Errorf("errors.As(NotFoundError) = %v, want %v", gotNotFound, tt.wantNotFound)
			}
			if err.Error() != tt.wantMessage {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantMessage)
			}
		})
	}
}
