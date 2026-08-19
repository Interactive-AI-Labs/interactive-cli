package clients

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// A provider that refuses still answers 200, so the status — not the
// transport — is what tells run-tool whether it has a result to print.
func TestDecodeMcpToolCallResult(t *testing.T) {
	errorClass := "unauthorized"
	tests := []struct {
		name           string
		body           string
		wantStatus     string
		wantErrorClass *string
		wantResult     bool
	}{
		{
			name:           "error status carries a class and no result",
			body:           `{"data":{"name":"socket","tool":"depscore","status":"error","error_class":"unauthorized"}}`,
			wantStatus:     "error",
			wantErrorClass: &errorClass,
			wantResult:     false,
		},
		{
			name:       "ok status carries a result and no class",
			body:       `{"data":{"name":"socket","tool":"depscore","status":"ok","result":{"ok":true}}}`,
			wantStatus: "ok",
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res McpToolCallResult
			if err := json.Unmarshal([]byte(tt.body), &res); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if diff := cmp.Diff(tt.wantStatus, res.Data.Status); diff != "" {
				t.Errorf("Status mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantErrorClass, res.Data.ErrorClass); diff != "" {
				t.Errorf("ErrorClass mismatch (-want +got):\n%s", diff)
			}
			if hasResult := len(res.Data.Result) > 0; hasResult != tt.wantResult {
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
