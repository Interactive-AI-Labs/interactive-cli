package clients

import (
	"encoding/json"
	"strings"
	"testing"
)

// Request bodies are built in cmd and serialized here; these assert the wire
// shape (omitempty rules, catalog vs custom) without a fake HTTP server.
func TestMcpConnectionCreateBodyMarshal(t *testing.T) {
	tests := []struct {
		name   string
		body   McpConnectionCreateBody
		want   []string
		absent []string
	}{
		{
			name: "custom connector carries endpoint and transport",
			body: McpConnectionCreateBody{
				Type:        "custom",
				Name:        "db",
				EndpointURL: "https://x",
				Transport:   "streamable_http",
				AuthType:    "none",
			},
			want: []string{
				`"type":"custom"`,
				`"endpoint_url":"https://x"`,
				`"transport":"streamable_http"`,
				`"auth_type":"none"`,
			},
			absent: []string{
				`"catalog_id"`,
				`"slug"`,
				`"description"`,
				`"credential"`,
				`"custom_headers"`,
			},
		},
		{
			name: "catalog connector forwards endpoint_url and omits transport",
			// The backend verifies endpoint_url against the catalog entry, so the
			// command forwards it for catalog connections too; transport is omitempty.
			body: McpConnectionCreateBody{
				Type:        "platform",
				CatalogID:   "github",
				EndpointURL: "https://mcp.github.com/",
				Name:        "gh",
				AuthType:    "bearer",
				Credential:  "tok",
			},
			want: []string{
				`"type":"platform"`,
				`"catalog_id":"github"`,
				`"endpoint_url":"https://mcp.github.com/"`,
				`"credential":"tok"`,
			},
			absent: []string{`"transport"`},
		},
		{
			name: "custom headers serialized when present",
			body: McpConnectionCreateBody{
				Type:          "custom",
				Name:          "h",
				EndpointURL:   "https://x",
				AuthType:      "api_key",
				Credential:    "k",
				CustomHeaders: map[string]string{"X-Team": "platform"},
			},
			want: []string{`"custom_headers":{"X-Team":"platform"}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}
			got := string(raw)
			for _, s := range tt.want {
				if !strings.Contains(got, s) {
					t.Fatalf("body %s missing %s", got, s)
				}
			}
			for _, s := range tt.absent {
				if strings.Contains(got, s) {
					t.Fatalf("body %s should omit %s", got, s)
				}
			}
		})
	}
}

func TestDecodeMcpConnectionList(t *testing.T) {
	data, err := decodeSuccess[McpConnectionListData](
		[]byte(
			`{"success":true,"data":{"connections":[{"id":"c1","name":"github","type":"custom","status":"ok","tool_count":3}]}}`,
		),
		"list mcp connections",
	)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(data.Connections) != 1 || data.Connections[0].ID != "c1" ||
		data.Connections[0].ToolCount != 3 {
		t.Fatalf("unexpected connections: %#v", data.Connections)
	}
}

func TestDecodeMcpConnectionDetail(t *testing.T) {
	data, err := decodeSuccess[McpConnectionDetailData](
		[]byte(
			`{"success":true,"data":{"connection":{"id":"c1","name":"github","type":"custom","status":"ok","tools":[{"name":"search","enabled":true}]}}}`,
		),
		"get mcp connection",
	)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	conn := data.Connection
	if conn.ID != "c1" || len(conn.Tools) != 1 || conn.Tools[0].Name != "search" {
		t.Fatalf("unexpected detail: %#v", conn)
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

func TestDecodeMcpVerify(t *testing.T) {
	data, err := decodeSuccess[McpVerifyData](
		[]byte(
			`{"success":true,"data":{"status":"ok","protocol_version":"2025-03-26","tools":[{"name":"t","enabled":true}]}}`,
		),
		"verify mcp connection",
	)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if data.Status != "ok" || data.ProtocolVersion != "2025-03-26" || len(data.Tools) != 1 {
		t.Fatalf("unexpected verify: %#v", data)
	}
}

// Result is json.RawMessage so a tool may return an object, array, or scalar at
// the top level without the non-object shapes being dropped.
func TestDecodeMcpToolCallResultShapes(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus string
		wantResult string
	}{
		{
			"object result",
			`{"success":true,"data":{"status":"ok","result":{"content":"hi"}}}`,
			"ok",
			`{"content":"hi"}`,
		},
		{
			"array result",
			`{"success":true,"data":{"status":"ok","result":[1,2,3]}}`,
			"ok",
			`[1,2,3]`,
		},
		{"scalar result", `{"success":true,"data":{"status":"ok","result":42}}`, "ok", `42`},
		{
			"error has no result",
			`{"success":true,"data":{"status":"error","error_class":"tool_error","error_message":"boom"}}`,
			"error",
			``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := decodeSuccess[McpToolCallData]([]byte(tt.body), "run mcp tool")
			if err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if data.Status != tt.wantStatus {
				t.Fatalf("status = %q, want %q", data.Status, tt.wantStatus)
			}
			if got := strings.TrimSpace(string(data.Result)); got != tt.wantResult {
				t.Fatalf("result = %q, want %q", got, tt.wantResult)
			}
		})
	}
}
