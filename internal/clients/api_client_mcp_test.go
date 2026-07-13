package clients

import (
	"testing"
)

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
