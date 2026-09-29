package deployment

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeRunMcpToolResult(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantError  *McpToolError
		wantResult bool
	}{
		{
			name:       "error body populates Error, not Result",
			body:       `{"mcp":"socket","tool":"depscore","error":{"code":-32603,"message":"boom"}}`,
			wantError:  &McpToolError{Code: -32603, Message: "boom"},
			wantResult: false,
		},
		{
			name:       "success body populates Result, not Error",
			body:       `{"mcp":"socket","tool":"depscore","result":{"ok":true}}`,
			wantError:  nil,
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res RunMcpToolResult
			if err := json.Unmarshal([]byte(tt.body), &res); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if diff := cmp.Diff(tt.wantError, res.Error); diff != "" {
				t.Errorf("Error mismatch (-want +got):\n%s", diff)
			}
			if hasResult := len(res.Result) > 0; hasResult != tt.wantResult {
				t.Errorf("has Result = %v, want %v", hasResult, tt.wantResult)
			}
		})
	}
}

func TestCreateMcpBodyEndpointJSON(t *testing.T) {
	tests := []struct {
		name     string
		endpoint bool
		want     string
	}{
		{name: "enabled is sent", endpoint: true, want: `{"type":"internal","image":{"type":"","name":"","tag":""},"resources":{"memory":"","cpu":""},"endpoint":true,"auth":{}}`},
		{name: "disabled is omitted", endpoint: false, want: `{"type":"internal","image":{"type":"","name":"","tag":""},"resources":{"memory":"","cpu":""},"auth":{}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(CreateMcpBody{Type: "internal", Endpoint: tt.endpoint})
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}
			if diff := cmp.Diff(tt.want, string(got)); diff != "" {
				t.Errorf("body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
