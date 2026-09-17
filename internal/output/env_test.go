package output

import (
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestFormatEnvValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
		wantJSON string
	}{
		{
			name:     "literal value",
			input:    `{"name":"LOG_LEVEL","value":"debug"}`,
			want:     "debug",
			wantJSON: `{"name":"LOG_LEVEL","value":"debug"}`,
		},
		{
			name:     "empty literal is not presented as a secret",
			input:    `{"name":"EMPTY","value":""}`,
			want:     "",
			wantJSON: `{"name":"EMPTY","value":""}`,
		},
		{
			name:     "secret reference is displayed and survives JSON output",
			input:    `{"name":"MCP_KEY_TOOLS_DEV","value":"","valueFrom":{"secretKeyRef":{"name":"tools-dev","key":"MCP_API_KEY"}}}`,
			want:     "<secret: tools-dev/MCP_API_KEY>",
			wantJSON: `{"name":"MCP_KEY_TOOLS_DEV","value":"","valueFrom":{"secretKeyRef":{"name":"tools-dev","key":"MCP_API_KEY"}}}`,
		},
		{
			name:     "old API missing reference is not guessed from the name",
			input:    `{"name":"MCP_KEY_TOOLS_DEV","value":""}`,
			want:     "",
			wantJSON: `{"name":"MCP_KEY_TOOLS_DEV","value":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env deployment.EnvVar
			if err := json.Unmarshal([]byte(tt.input), &env); err != nil {
				t.Fatal(err)
			}
			if got := formatEnvValue(env); got != tt.want {
				t.Fatalf("value = %q, want %q", got, tt.want)
			}
			got, err := json.Marshal(env)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.wantJSON {
				t.Fatalf("JSON = %s, want %s", got, tt.wantJSON)
			}
		})
	}
}
