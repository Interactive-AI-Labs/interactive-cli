package output

import (
	"bytes"
	"strings"
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
			name: "static key entry has no sign-in details to show",
			entries: []clients.McpCatalogEntry{
				{
					ID:          "e1",
					Name:        "GitHub",
					Category:    "dev",
					Type:        "platform",
					AuthMethods: []string{"api_key"},
				},
			},
			want: "ID   NAME     CATEGORY   SIGN-IN   PERMISSIONS\n" +
				"e1   GitHub   dev        api_key   —\n",
		},
		{
			name: "self-registering provider whose permissions are all-or-nothing",
			entries: []clients.McpCatalogEntry{
				{
					ID:               "dropbox",
					Name:             "Dropbox",
					Category:         "Productivity",
					Type:             "platform",
					GrantsAllowed:    []string{"pkce"},
					ScopesSupported:  []string{"files.metadata.read", "files.content.read"},
					ScopesSelectable: boolPtr(false),
					SelfRegisters:    map[string]bool{"pkce": true},
				},
			},
			want: "ID        NAME      CATEGORY       SIGN-IN   PERMISSIONS\n" +
				"dropbox   Dropbox   Productivity   oauth     files.metadata.read, files.content.read (all required)\n",
		},
		{
			name: "provider needing our app, that also takes a plain token",
			entries: []clients.McpCatalogEntry{
				{
					ID:              "github",
					Name:            "GitHub",
					Category:        "Development",
					Type:            "platform",
					AuthMethods:     []string{"bearer"},
					GrantsAllowed:   []string{"pkce"},
					ScopesSupported: []string{"repo", "read:org"},
					SelfRegisters:   map[string]bool{"pkce": false},
				},
			},
			want: "ID       NAME     CATEGORY      SIGN-IN                             PERMISSIONS\n" +
				"github   GitHub   Development   oauth, bearer · needs admin setup   repo, read:org\n",
		},
		{
			name: "provider that publishes a grant but not whether it registers us",
			entries: []clients.McpCatalogEntry{
				{
					ID:            "gorgias",
					Name:          "Gorgias",
					Category:      "Customer Support",
					Type:          "platform",
					GrantsAllowed: []string{"pkce"},
				},
			},
			want: "ID        NAME      CATEGORY           SIGN-IN   PERMISSIONS\n" +
				"gorgias   Gorgias   Customer Support   oauth     —\n",
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
		name  string
		tools []clients.McpToolSchema
		want  string
	}{
		{
			name: "one tool with description",
			tools: []clients.McpToolSchema{
				{Name: "search", Description: strPtr("Search the knowledge base")},
				{Name: "ping", Description: strPtr("No-arg health check")},
			},
			want: "Tools (2):\n  search - Search the knowledge base\n  ping - No-arg health check\n",
		},
		{
			name:  "no tools cached",
			tools: nil,
			want:  "No tools cached - run 'iai mcps verify' first.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintMcpTools(&buf, tt.tools); err != nil {
				t.Fatalf("PrintMcpTools() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string { return &s }

// boolPtr is for catalog fields where nil means "not established yet", which
// must stay distinct from false.
func boolPtr(b bool) *bool { return &b }

func TestPrintMcpsNeedsSignIn(t *testing.T) {
	tests := []struct {
		name string
		mcp  clients.McpSchema
		want string
	}{
		{
			name: "oauth mcp that was never connected",
			mcp:  clients.McpSchema{Name: "notion", Backend: "external", AuthType: strPtr("oauth")},
			want: "needs sign-in",
		},
		{
			name: "oauth mcp already connected",
			mcp:  clients.McpSchema{Name: "notion", Backend: "external", AuthType: strPtr("oauth"), HasCredential: true},
			want: "set",
		},
		{
			name: "static mcp with credential",
			mcp:  clients.McpSchema{Name: "acme", Backend: "external", AuthType: strPtr("bearer"), HasCredential: true},
			want: "set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintMcpList(&buf, []clients.McpSchema{tt.mcp}); err != nil {
				t.Fatalf("PrintMcpList() error = %v", err)
			}
			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("output %q does not contain %q", buf.String(), tt.want)
			}
		})
	}
}

// Prepending unconditionally printed "oauth, oauth" when auth_methods had it too.
func TestMcpSignInDoesNotRepeatOauth(t *testing.T) {
	tests := []struct {
		name  string
		entry clients.McpCatalogEntry
		want  string
	}{
		{
			"oauth already in auth_methods",
			clients.McpCatalogEntry{
				AuthMethods:   []string{"oauth"},
				GrantsAllowed: []string{"pkce"},
			},
			"oauth",
		},
		{
			"oauth alongside another method",
			clients.McpCatalogEntry{
				AuthMethods:   []string{"bearer", "oauth"},
				GrantsAllowed: []string{"pkce"},
			},
			"oauth, bearer",
		},
		{
			"no grants recorded falls back to auth_methods as-is",
			clients.McpCatalogEntry{AuthMethods: []string{"bearer"}},
			"bearer",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := mcpSignIn(tc.entry); got != tc.want {
				t.Errorf("mcpSignIn() = %q, want %q", got, tc.want)
			}
		})
	}
}
