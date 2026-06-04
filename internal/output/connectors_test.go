package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintMcpConnectionListEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintMcpConnectionList(&buf, nil); err != nil {
		t.Fatalf("error = %v", err)
	}
	if buf.String() != "No connectors found.\n" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPrintMcpConnectionListRows(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")
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
			EndpointURL: "https://x", Transport: "streamable_http",
			AuthType: "bearer", HasCredential: true,
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
	if !strings.Contains(out, "Credential Set:") || !strings.Contains(out, "true") {
		t.Fatalf("expected credential-set line:\n%s", out)
	}
	if strings.Index(out, "ID:") > strings.Index(out, "NAME") {
		t.Fatalf("describe block should render before tools table:\n%s", out)
	}
}

func TestPrintMcpConnectionDetailNoAuthHidesCredential(t *testing.T) {
	var buf bytes.Buffer
	conn := &clients.McpConnectionDetail{
		McpConnection: clients.McpConnection{
			ID: "c2", Name: "open", Type: "custom", Status: "ok",
			EndpointURL: "https://x", Transport: "streamable_http", AuthType: "none",
		},
	}
	if err := PrintMcpConnectionDetail(&buf, conn); err != nil {
		t.Fatalf("error = %v", err)
	}
	if strings.Contains(buf.String(), "Credential Set:") {
		t.Fatalf("credential-set line should be hidden for auth_type=none:\n%s", buf.String())
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

func TestPrintMcpToolResult(t *testing.T) {
	tests := []struct {
		name       string
		res        *clients.McpToolCallData
		wantOutput []string
		wantErr    bool
	}{
		{
			name: "ok with object result",
			res: &clients.McpToolCallData{
				Status: "ok",
				Result: json.RawMessage(`{"content":"hi"}`),
			},
			wantOutput: []string{"ok", "content"},
			wantErr:    false,
		},
		{
			name:       "ok with array result is preserved",
			res:        &clients.McpToolCallData{Status: "ok", Result: json.RawMessage(`[1,2,3]`)},
			wantOutput: []string{"Result:", "1"},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintMcpToolResult(&buf, tt.res)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			out := buf.String()
			for _, want := range tt.wantOutput {
				if !strings.Contains(out, want) {
					t.Fatalf("output missing %q:\n%s", want, out)
				}
			}
		})
	}
}

func TestMcpToolCallError(t *testing.T) {
	tests := []struct {
		name string
		res  *clients.McpToolCallData
		want []string
	}{
		{
			name: "class and message",
			res: &clients.McpToolCallData{
				Status:       "error",
				ErrorClass:   "tool_error",
				ErrorMessage: "boom",
			},
			want: []string{`status "error"`, "tool_error", "boom"},
		},
		{
			name: "message only",
			res:  &clients.McpToolCallData{Status: "error", ErrorMessage: "boom"},
			want: []string{`status "error"`, "boom"},
		},
		{
			name: "class only",
			res:  &clients.McpToolCallData{Status: "error", ErrorClass: "tool_error"},
			want: []string{`status "error"`, "tool_error"},
		},
		{
			name: "status only",
			res:  &clients.McpToolCallData{Status: "timeout"},
			want: []string{`status "timeout"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := McpToolCallError(tt.res)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			for _, want := range tt.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q missing %q", err.Error(), want)
				}
			}
		})
	}
}
