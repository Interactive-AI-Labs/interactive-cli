package clients

import (
	"net/http"
	"testing"
	"time"
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
			got := formatAgentValidationError(tt.body)
			if got != tt.want {
				t.Errorf("formatAgentValidationError() = %q, want %q", got, tt.want)
			}
		})
	}
}
