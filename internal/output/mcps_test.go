package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintMcpCatalog(t *testing.T) {
	tests := []struct {
		name    string
		entries []clients.McpCatalogEntry
		want    string
	}{
		{
			name: "single entry",
			entries: []clients.McpCatalogEntry{
				{
					ID:          "e1",
					Name:        "GitHub",
					Category:    "dev",
					Type:        "platform",
					AuthMethods: []string{"api_key"},
				},
			},
			want: "ID   NAME     CATEGORY   TYPE       AUTH\n" +
				"e1   GitHub   dev        platform   api_key\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintMcpCatalog(&buf, tt.entries); err != nil {
				t.Fatalf("PrintMcpCatalog() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintMcpTools(t *testing.T) {
	tests := []struct {
		name        string
		tools       []map[string]any
		added       []string
		removed     []string
		changedFrom string
		want        string
	}{
		{
			name: "tool with args and one without",
			tools: []map[string]any{
				{
					"name":        "search",
					"description": "Search the knowledge base",
					"inputSchema": map[string]any{
						"properties": map[string]any{
							"query": map[string]any{"type": "string"},
							"limit": map[string]any{"type": "integer"},
						},
					},
				},
				{"name": "ping", "description": "No-arg health check"},
			},
			want: "Tools (2):\n" +
				"  search(limit, query) — Search the knowledge base\n" +
				"  ping — No-arg health check\n",
		},
		{
			name:  "no tools cached",
			tools: nil,
			want:  "No tools cached — run 'iai mcps verify' first.\n",
		},
		{
			name: "changed since revision",
			tools: []map[string]any{
				{"name": "search", "description": "Search"},
			},
			added:       []string{"search"},
			removed:     []string{"fetch"},
			changedFrom: "3",
			want: "Tools (1):\n" +
				"  search — Search\n" +
				"\nChanged since revision 3: +search -fetch\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintMcpTools(&buf, tt.tools, tt.added, tt.removed, tt.changedFrom)
			if err != nil {
				t.Fatalf("PrintMcpTools() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
