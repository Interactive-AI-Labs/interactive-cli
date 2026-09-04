package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func TestPrintMcpCatalog(t *testing.T) {
	tests := []struct {
		name    string
		entries []platform.McpCatalogEntry
		want    string
	}{
		{
			name: "static key entry has no sign-in details to show",
			entries: []platform.McpCatalogEntry{
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
			entries: []platform.McpCatalogEntry{
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
			// The provider does OAuth, but this entry is not allowed to be created
			// with it — offering the sign-in would send the user down a dead end.
			name: "provider publishes a grant the entry does not accept",
			entries: []platform.McpCatalogEntry{
				{
					ID:            "hf",
					Name:          "Hugging Face",
					Category:      "dev",
					Type:          "platform",
					AuthMethods:   []string{"bearer", "api_key"},
					GrantsAllowed: []string{"pkce"},
				},
			},
			want: "ID   NAME           CATEGORY   SIGN-IN           PERMISSIONS\n" +
				"hf   Hugging Face   dev        bearer, api_key   —\n",
		},
		{
			name: "provider needing our app, that also takes a plain token",
			entries: []platform.McpCatalogEntry{
				{
					ID:              "github",
					Name:            "GitHub",
					Category:        "Development",
					Type:            "platform",
					AuthMethods:     []string{"oauth", "bearer"},
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
			entries: []platform.McpCatalogEntry{
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

// The unified list carries stack_id so `iai stacks` needs no describe-per-mcp.
func TestPrintMcpListIncludesStack(t *testing.T) {
	tests := []struct {
		name string
		mcp  platform.McpSchema
		want string
	}{
		{
			name: "internal mcp in a stack",
			mcp: platform.McpSchema{
				Name: "tools", Backend: "internal",
				Status: strPtr("healthy"), VerifyStatus: strPtr("ok"),
				ToolCount: 3, StackId: strPtr("stack-123"),
			},
			want: "stack-123",
		},
		{
			name: "external mcp never has one",
			mcp: platform.McpSchema{
				Name: "notion", Backend: "external", AuthType: strPtr("oauth"),
			},
			want: "-",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintMcpList(&buf, []platform.McpSchema{tt.mcp}); err != nil {
				t.Fatalf("PrintMcpList() error = %v", err)
			}
			got := buf.String()
			if !strings.Contains(got, "STACK") {
				t.Errorf("header is missing the STACK column:\n%q", got)
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("output %q does not contain %q", got, tt.want)
			}
		})
	}
}

func TestPrintMcpTools(t *testing.T) {
	tests := []struct {
		name    string
		backend platform.McpBackend
		tools   []platform.McpToolSchema
		want    string
	}{
		{
			name:    "one tool with description",
			backend: platform.McpBackendExternal,
			tools: []platform.McpToolSchema{
				{Name: "search", Description: strPtr("Search the knowledge base")},
				{Name: "ping", Description: strPtr("No-arg health check")},
			},
			want: "Tools (2):\n  search - Search the knowledge base\n  ping - No-arg health check\n",
		},
		{
			name:    "an external mcp with no tools is sent to verify",
			backend: platform.McpBackendExternal,
			tools:   nil,
			want:    "No tools cached - run 'iai mcps verify' first.\n",
		},
		{
			// `verify` refuses an internal mcp, so sending the user there is a dead end.
			name:    "an internal mcp is not sent to verify",
			backend: platform.McpBackendInternal,
			tools:   nil,
			want: "No tools cached - an internal mcp verifies itself when its workload " +
				"becomes ready; redeploy it to run that again.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintMcpTools(&buf, tt.backend, tt.tools); err != nil {
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
		mcp  platform.McpSchema
		want string
	}{
		{
			name: "oauth mcp that was never connected",
			mcp: platform.McpSchema{
				Name:     "notion",
				Backend:  "external",
				AuthType: strPtr("oauth"),
			},
			want: "needs sign-in",
		},
		{
			name: "oauth mcp already connected",
			mcp: platform.McpSchema{
				Name:          "notion",
				Backend:       "external",
				AuthType:      strPtr("oauth"),
				HasCredential: true,
			},
			want: "set",
		},
		{
			name: "static mcp with credential",
			mcp: platform.McpSchema{
				Name:          "acme",
				Backend:       "external",
				AuthType:      strPtr("bearer"),
				HasCredential: true,
			},
			want: "set",
		},
		{
			// Rows written before Platform recorded its own auth type carry
			// LiteLLM's name for the flow. Same state, so the same column.
			name: "unconnected oauth mcp carrying LiteLLM's auth type",
			mcp: platform.McpSchema{
				Name:     "linear",
				Backend:  "external",
				AuthType: strPtr("oauth2"),
			},
			want: "needs sign-in",
		},
		{
			name: "unconnected oauth_delegate mcp",
			mcp: platform.McpSchema{
				Name:     "bridge",
				Backend:  "external",
				AuthType: strPtr("oauth_delegate"),
			},
			want: "needs sign-in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintMcpList(&buf, []platform.McpSchema{tt.mcp}); err != nil {
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
		entry platform.McpCatalogEntry
		want  string
	}{
		{
			"oauth already in auth_methods",
			platform.McpCatalogEntry{
				AuthMethods:   []string{"oauth"},
				GrantsAllowed: []string{"pkce"},
			},
			"oauth",
		},
		{
			"oauth alongside another method",
			platform.McpCatalogEntry{
				AuthMethods:   []string{"bearer", "oauth"},
				GrantsAllowed: []string{"pkce"},
			},
			"oauth, bearer",
		},
		{
			"no grants recorded falls back to auth_methods as-is",
			platform.McpCatalogEntry{AuthMethods: []string{"bearer"}},
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

// describe used to send an unconnected oauth mcp to `iai mcps tools`, which
// errors. It has to name the command that actually helps.
func TestPrintMcpDetailPointsAtConnectWhenUnsigned(t *testing.T) {
	tests := []struct {
		name    string
		mcp     platform.McpSchema
		want    string
		notWant string
	}{
		{
			name: "unconnected oauth mcp is told to connect",
			mcp: platform.McpSchema{
				Name:     "asana",
				Backend:  "external",
				AuthType: strPtr("oauth"),
			},
			want:    "iai mcps connect asana",
			notWant: "iai mcps tools asana",
		},
		{
			name: "connected oauth mcp is told where the tools are",
			mcp: platform.McpSchema{
				Name:          "asana",
				Backend:       "external",
				AuthType:      strPtr("oauth"),
				HasCredential: true,
				ToolCount:     46,
			},
			want:    "iai mcps tools asana",
			notWant: "iai mcps connect asana",
		},
		{
			name: "static credential mcp is never told to connect",
			mcp: platform.McpSchema{
				Name:          "acme",
				Backend:       "external",
				AuthType:      strPtr("bearer"),
				HasCredential: true,
			},
			want:    "iai mcps tools acme",
			notWant: "iai mcps connect acme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			mcp := tt.mcp
			if err := PrintMcpDetail(&buf, &mcp); err != nil {
				t.Fatalf("PrintMcpDetail() error = %v", err)
			}
			got := buf.String()
			if !strings.Contains(got, tt.want) {
				t.Errorf("output %q does not contain %q", got, tt.want)
			}
			if strings.Contains(got, tt.notWant) {
				t.Errorf("output %q must not contain %q", got, tt.notWant)
			}
		})
	}
}
