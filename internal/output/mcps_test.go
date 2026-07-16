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
