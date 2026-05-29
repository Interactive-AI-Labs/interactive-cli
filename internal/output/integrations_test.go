package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintMcpConnectionListEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintMcpConnectionList(&buf, nil); err != nil {
		t.Fatalf("error = %v", err)
	}
	if buf.String() != "No integration connections found.\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPrintMcpConnectionListRows(t *testing.T) {
	var buf bytes.Buffer
	conns := []clients.McpConnection{
		{
			Name:        "github",
			Type:        "custom",
			Status:      "ok",
			ToolCount:   3,
			EndpointURL: "https://api.githubcopilot.com/mcp",
			UpdatedAt:   "2026-05-01T10:00:00Z",
		},
	}
	if err := PrintMcpConnectionList(&buf, conns); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "github") ||
		!strings.Contains(out, "custom") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestPrintMcpConnectionDetailWithTools(t *testing.T) {
	var buf bytes.Buffer
	conn := &clients.McpConnectionDetail{
		McpConnection: clients.McpConnection{
			ID: "c1", Name: "github", Type: "custom", Status: "ok",
			EndpointURL: "https://x", Transport: "streamable_http", AuthType: "bearer",
		},
		Tools: []clients.McpTool{{Name: "search", Enabled: true, Description: "Search repos"}},
	}
	if err := PrintMcpConnectionDetail(&buf, conn); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "github") || !strings.Contains(out, "search") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	// Describe block (flushed before the tools table) must appear first.
	if strings.Index(out, "ID:") > strings.Index(out, "NAME") {
		t.Fatalf("describe block should render before tools table:\n%s", out)
	}
}

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

func TestPrintMcpVerifyResultError(t *testing.T) {
	var buf bytes.Buffer
	res := &clients.McpVerifyData{
		Status:       "error",
		ErrorClass:   "unauthorized",
		ErrorMessage: "bad token",
	}
	if err := PrintMcpVerifyResult(&buf, res); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "error") || !strings.Contains(out, "unauthorized") ||
		!strings.Contains(out, "bad token") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestPrintMcpToolResultOk(t *testing.T) {
	var buf bytes.Buffer
	res := &clients.McpToolCallData{Status: "ok", Result: map[string]any{"content": "hi"}}
	if err := PrintMcpToolResult(&buf, res); err != nil {
		t.Fatalf("error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ok") || !strings.Contains(out, "content") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}
