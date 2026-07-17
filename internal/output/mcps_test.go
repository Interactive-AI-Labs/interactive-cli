package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintMcpCatalog(t *testing.T) {
	var buf bytes.Buffer
	entries := []clients.McpCatalogEntry{
		{
			ID:          "e1",
			Name:        "GitHub",
			Category:    "dev",
			Type:        "platform",
			AuthMethods: []string{"api_key"},
		},
	}
	if err := PrintMcpCatalog(&buf, entries); err != nil {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(buf.String(), "GitHub") || !strings.Contains(buf.String(), "e1") {
		t.Fatalf("unexpected output:\n%s", buf.String())
	}
}

func TestPrintMcpTools(t *testing.T) {
	tools := []map[string]any{
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
	}
	var buf bytes.Buffer
	if err := PrintMcpTools(&buf, tools, nil, nil, ""); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "search(limit, query) — Search the knowledge base") {
		t.Fatalf("expected sorted arg names in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ping — No-arg health check") {
		t.Fatalf("expected no parens for a tool with no inputSchema, got:\n%s", out)
	}
}
