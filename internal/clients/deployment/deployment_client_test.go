package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/buildinfo"
	"github.com/google/go-cmp/cmp"
)

func TestNewDeploymentClient(t *testing.T) {
	t.Run("creates client with API key", func(t *testing.T) {
		client, err := NewDeploymentClient(
			"https://deploy.example.com",
			30*time.Second,
			"",
			"test-key",
			nil,
		)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.apiKey != "test-key" {
			t.Errorf("apiKey = %q, want %q", client.apiKey, "test-key")
		}
	})

	t.Run("creates client with token", func(t *testing.T) {
		client, err := NewDeploymentClient(
			"https://deploy.example.com",
			30*time.Second,
			"fake-token",
			"",
			nil,
		)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.token != "fake-token" {
			t.Errorf("token = %q, want %q", client.token, "fake-token")
		}
	})

	t.Run("creates client with cookies", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewDeploymentClient(
			"https://deploy.example.com",
			30*time.Second,
			"",
			"",
			cookies,
		)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if len(client.cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(client.cookies))
		}
	})

	t.Run("returns error with no auth", func(t *testing.T) {
		_, err := NewDeploymentClient("https://deploy.example.com", 30*time.Second, "", "", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("stores hostname correctly", func(t *testing.T) {
		hostname := "https://deploy.example.com"
		client, err := NewDeploymentClient(hostname, 30*time.Second, "", "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}
		if client.hostname != hostname {
			t.Errorf("hostname = %q, want %q", client.hostname, hostname)
		}
	})
}

func TestListRevisionsWithAttribution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != buildinfo.UserAgent {
			t.Errorf("User-Agent = %q, want %q", got, buildinfo.UserAgent)
		}
		fmt.Fprint(w, `{"revisions":[{"revision":48,"updated":"2026-07-28T14:56:00Z",`+
			`"status":"deployed","actor":{"type":"api_key","id":"key_123",`+
			`"displayName":"silverspin-release"},"source":{"type":"cli",`+
			`"version":"0.39.0"},"requestId":"req_abc123"}]}`)
	}))
	t.Cleanup(server.Close)

	client, err := NewDeploymentClient(server.URL, 5*time.Second, "test-token", "", nil)
	if err != nil {
		t.Fatalf("NewDeploymentClient() error = %v", err)
	}
	want := RevisionMeta{
		Revision: 48,
		Updated:  "2026-07-28T14:56:00Z",
		Status:   "deployed",
		Actor: &RevisionActor{
			Type:        "api_key",
			ID:          "key_123",
			DisplayName: "silverspin-release",
		},
		Source:    &RevisionSource{Type: "cli", Version: "0.39.0"},
		RequestID: "req_abc123",
	}

	tests := []struct {
		name string
		list func(context.Context, string, string, string) ([]RevisionMeta, error)
	}{
		{name: "agents", list: client.ListAgentRevisions},
		{name: "services", list: client.ListServiceRevisions},
		{name: "mcps", list: client.ListMcpRevisions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			revisions, err := tt.list(t.Context(), "org-1", "project-1", "resource-1")
			if err != nil {
				t.Fatalf("list revisions: %v", err)
			}
			if len(revisions) != 1 {
				t.Fatalf("len(revisions) = %d, want 1", len(revisions))
			}

			if diff := cmp.Diff(want, revisions[0]); diff != "" {
				t.Errorf("revision mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRevisionResponsesWithoutAttribution(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{
			name:    "fields omitted",
			payload: `{"revision":47,"updated":"2026-07-24T10:20:00Z","status":"deployed"}`,
		},
		{
			name: "null historical fields",
			payload: `{"revision":47,"updated":"2026-07-24T10:20:00Z",` +
				`"status":"deployed","actor":null,"source":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var revision AgentRevisionResponse
			if err := json.Unmarshal([]byte(tt.payload), &revision); err != nil {
				t.Fatalf("decode old backend response: %v", err)
			}
			if revision.Revision != 47 {
				t.Errorf("Revision = %d, want 47", revision.Revision)
			}
			if revision.Actor != nil || revision.Source != nil || revision.RequestID != "" {
				t.Errorf(
					"attribution = (%#v, %#v, %q), want empty",
					revision.Actor,
					revision.Source,
					revision.RequestID,
				)
			}
		})
	}
}

func TestFormatAgentValidationError(t *testing.T) {
	tests := []struct {
		name string
		body []byte
		want string
	}{
		{
			name: "reference error not_found",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":{"detail":"Agent manifest references could not be resolved","errors":[{"path":"agent_config.context.description","id":"does-not-exist","version":1,"expected_type":null,"reason":"not_found"}]}}`,
			),
			want: "Agent manifest references could not be resolved\n  - agent_config.context.description: \"does-not-exist\" version 1 not found",
		},
		{
			name: "reference error version_not_found",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":{"detail":"Agent manifest references could not be resolved","errors":[{"path":"agent_config.context.routines[0]","id":"bonus","version":99,"expected_type":"routine","reason":"version_not_found","available_versions":[1,2]}]}}`,
			),
			want: "Agent manifest references could not be resolved\n  - agent_config.context.routines[0]: \"bonus\" version 99 not found (available: 1, 2)",
		},
		{
			name: "reference error wrong_type",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":{"detail":"Agent manifest references could not be resolved","errors":[{"path":"agent_config.context.routines[0]","id":"my-policy","version":1,"expected_type":"routine","reason":"wrong_type","actual_type":"policy"}]}}`,
			),
			want: "Agent manifest references could not be resolved\n  - agent_config.context.routines[0]: \"my-policy\" is type \"policy\", expected \"routine\"",
		},
		{
			name: "reference error priority_unresolved",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":{"detail":"Agent manifest references could not be resolved","errors":[{"path":"agent_config.context.relationships.priorities[0]","id":"unknown-routine","version":1,"expected_type":"routine_or_policy","reason":"priority_unresolved"}]}}`,
			),
			want: "Agent manifest references could not be resolved\n  - agent_config.context.relationships.priorities[0]: \"unknown-routine\" version 1: priority reference not in manifest",
		},
		{
			name: "multiple reference errors",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":{"detail":"Agent manifest references could not be resolved","errors":[{"path":"agent_config.context.routines[0]","id":"bonus","version":99,"expected_type":"routine","reason":"version_not_found","available_versions":[1]},{"path":"agent_config.context.routines[1]","id":"missing","version":1,"expected_type":"routine","reason":"not_found"}]}}`,
			),
			want: "Agent manifest references could not be resolved\n  - agent_config.context.routines[0]: \"bonus\" version 99 not found (available: 1)\n  - agent_config.context.routines[1]: \"missing\" version 1 not found",
		},
		{
			name: "structural error single field",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":[{"type":"missing","loc":["agent_config","context","description"],"msg":"Field required","input":{"language":"en"}}]}`,
			),
			want: "Agent configuration validation failed:\n  - agent_config.context.description: Field required",
		},
		{
			name: "structural error multiple fields",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":[{"type":"missing","loc":["agent_config","context","description"],"msg":"Field required"},{"type":"string_type","loc":["agent_config","context","language"],"msg":"Input should be a valid string"}]}`,
			),
			want: "Agent configuration validation failed:\n  - agent_config.context.description: Field required\n  - agent_config.context.language: Input should be a valid string",
		},
		{
			name: "schema validation error single",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":{"detail":"Agent config does not match schema","errors":[{"path":"context.preamble","message":"'examples' is a required property"}]}}`,
			),
			want: "Agent config does not match schema\n  - context.preamble: 'examples' is a required property",
		},
		{
			name: "schema validation error multiple",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":{"detail":"Agent config does not match schema","errors":[{"path":"llms","message":"'llms' is a required property"},{"path":"context","message":"Additional properties are not allowed ('description' was unexpected)"},{"path":"context.system_prompt","message":"'system_prompt' is a required property"}]}}`,
			),
			want: "Agent config does not match schema\n  - llms: 'llms' is a required property\n  - context: Additional properties are not allowed ('description' was unexpected)\n  - context.system_prompt: 'system_prompt' is a required property",
		},
		{
			name: "no detail field falls back to message",
			body: []byte(`{"code":422,"message":"Agent configuration validation failed"}`),
			want: "Agent configuration validation failed",
		},
		{
			name: "unrecognized detail falls back to message",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":"unexpected string"}`,
			),
			want: "Agent configuration validation failed",
		},
		{
			name: "reference error with unknown reason",
			body: []byte(
				`{"code":422,"message":"Agent configuration validation failed","detail":{"detail":"Agent manifest references could not be resolved","errors":[{"path":"agent_config.context.routines[0]","id":"something","version":1,"expected_type":"routine","reason":"new_unknown_reason"}]}}`,
			),
			want: "Agent manifest references could not be resolved\n  - agent_config.context.routines[0]: \"something\": new_unknown_reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fmtAgentValErr(tt.body)
			if got != tt.want {
				t.Errorf("fmtAgentValErr() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNamedHost covers the annotation that tells a wrong-host mistake apart from
// a permissions one: --hostname and --deployment-hostname default independently,
// so overriding only the first leaves this client on production.
func TestNamedHost(t *testing.T) {
	c := &DeploymentClient{hostname: "https://deployment.interactive.ai"}
	tests := []struct {
		name   string
		status int
		msg    string
		want   string
	}{
		{
			name:   "unauthorized names the host that answered",
			status: http.StatusUnauthorized,
			msg:    "Unauthorized",
			want:   "Unauthorized (deployment API at https://deployment.interactive.ai)",
		},
		{
			name:   "forbidden names it too",
			status: http.StatusForbidden,
			msg:    "Forbidden",
			want:   "Forbidden (deployment API at https://deployment.interactive.ai)",
		},
		{
			name:   "a not-found is about the resource, not the host",
			status: http.StatusNotFound,
			msg:    "Revision not found",
			want:   "Revision not found",
		},
		{
			name:   "a server error is left alone",
			status: http.StatusInternalServerError,
			msg:    "boom",
			want:   "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.namedHost(tt.status, tt.msg); got != tt.want {
				t.Errorf("namedHost() = %q, want %q", got, tt.want)
			}
		})
	}
}
