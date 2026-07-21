package clients

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

func TestDecodeMcpCatalog(t *testing.T) {
	data, err := decodeSuccess[McpCatalogListData](
		[]byte(
			`{"success":true,"data":{"entries":[{"id":"e1","name":"GitHub","category":"dev","type":"platform","auth_methods":["api_key"]}]}}`,
		),
		"list mcp catalog",
	)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(data.Entries) != 1 || data.Entries[0].ID != "e1" {
		t.Fatalf("unexpected catalog: %#v", data.Entries)
	}
}
