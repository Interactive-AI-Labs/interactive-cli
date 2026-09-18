package platform

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newScopeTestClient serves body for every request and records the raw
// request-target, which is what the CLI actually put on the wire.
func newScopeTestClient(t *testing.T, body string, status int) (*APIClient, *string) {
	t.Helper()
	var gotURI string
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotURI = r.RequestURI
			w.WriteHeader(status)
			_, _ = io.WriteString(w, body)
		}),
	)
	t.Cleanup(server.Close)

	client, err := NewAPIClient(server.URL, 5*time.Second, "fake-token", "", nil)
	if err != nil {
		t.Fatalf("NewAPIClient() error = %v", err)
	}
	return client, &gotURI
}

// The global scope rides on the project's own route, so the query param is the
// whole contract: drop it and the project's rows come back instead, with no error.
func TestListPromptsSendsScope(t *testing.T) {
	const body = `{"success":true,"data":{"prompts":[],"totalCount":0}}`
	base := "/api/platform/v1/projects/proj-1/prompts/skills"

	tests := []struct {
		name string
		opts PromptListOptions
		want string
	}{
		{name: "project scope sends nothing", want: base},
		{
			name: "global scope",
			opts: PromptListOptions{Scope: ScopeGlobal},
			want: base + "?scope=global",
		},
		{
			// The same Limit has to reach both reads, or a server that ignores
			// scope=global answers with a page the merge cannot de-dupe.
			name: "global scope keeps the limit",
			opts: PromptListOptions{Limit: 3, Scope: ScopeGlobal},
			want: base + "?limit=3&scope=global",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, gotURI := newScopeTestClient(t, body, http.StatusOK)
			if _, err := client.ListPrompts(
				context.Background(),
				"proj-1",
				"skills",
				tt.opts,
			); err != nil {
				t.Fatalf("ListPrompts() error = %v", err)
			}
			if *gotURI != tt.want {
				t.Errorf("request = %q, want %q", *gotURI, tt.want)
			}
		})
	}
}

// A nested name goes out percent-encoded; the server decodes it before matching
// {name:path}. version and label are never sent: a global skill has one version.
func TestGetPromptSendsScope(t *testing.T) {
	const body = `{"success":true,"data":{"name":"team/deploy","prompt":"# Deploy","labels":["active"]}}`
	client, gotURI := newScopeTestClient(t, body, http.StatusOK)

	result, err := client.GetPrompt(
		context.Background(), "proj-1", "skills", "team/deploy", 0, "", ScopeGlobal,
	)
	if err != nil {
		t.Fatalf("GetPrompt() error = %v", err)
	}

	want := "/api/platform/v1/projects/proj-1/prompts/skills/team%2Fdeploy?scope=global"
	if *gotURI != want {
		t.Errorf("request = %q, want %q", *gotURI, want)
	}
	if result.Name != "team/deploy" || result.Version != 0 {
		t.Errorf(
			"detail = %q v%d, want %q with no version",
			result.Name,
			result.Version,
			"team/deploy",
		)
	}
}

// get tells "no such skill, try the global scope" from "the server is broken" by
// the error type alone, so a 404 has to arrive as NotFoundError and nothing else may.
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
			name:         "404 with an unparsable body",
			status:       http.StatusNotFound,
			body:         "not json",
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
			client, _ := newScopeTestClient(t, tt.body, tt.status)
			_, err := client.GetPrompt(context.Background(), "proj-1", "skills", "nope", 0, "", "")
			if err == nil {
				t.Fatalf("GetPrompt() error = nil, want %q", tt.wantMessage)
			}
			var notFound *NotFoundError
			if errors.As(err, &notFound) != tt.wantNotFound {
				t.Errorf(
					"errors.As(NotFoundError) = %v, want %v",
					!tt.wantNotFound,
					tt.wantNotFound,
				)
			}
			if err.Error() != tt.wantMessage {
				t.Errorf("error = %q, want %q", err, tt.wantMessage)
			}
		})
	}
}
