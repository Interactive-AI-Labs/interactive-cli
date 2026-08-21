package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintMcpList(out io.Writer, mcps []clients.McpSchema) error {
	if len(mcps) == 0 {
		fmt.Fprintln(out, "No mcps found.")
		return nil
	}
	headers := []string{"NAME", "BACKEND", "STATUS", "VERIFY", "TOOLS", "CREDENTIAL", "STACK"}
	rows := make([][]string, len(mcps))
	for i, m := range mcps {
		status := "-"
		if m.Status != nil {
			status = *m.Status
		}
		verify := "-"
		if m.VerifyStatus != nil {
			verify = *m.VerifyStatus
		}
		if NeedsSignIn(m) {
			verify = "needs sign-in"
		}
		rows[i] = []string{
			m.Name,
			string(m.Backend),
			status,
			verify,
			descOr(m.ToolCount),
			boolOr(m.HasCredential),
			strOr(m.StackId),
		}
	}
	return PrintTable(out, headers, rows)
}

func strOr(v *string) string {
	if v == nil || *v == "" {
		return "-"
	}
	return *v
}

func descOr(n int) string {
	return fmt.Sprintf("%d", n)
}

func boolOr(b bool) string {
	if b {
		return "set"
	}
	return "none"
}

func PrintMcpCatalog(out io.Writer, entries []clients.McpCatalogEntry) error {
	if len(entries) == 0 {
		fmt.Fprintln(out, "No catalog entries found.")
		return nil
	}
	headers := []string{"ID", "NAME", "CATEGORY", "SIGN-IN", "PERMISSIONS"}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		rows[i] = []string{
			e.ID, e.Name, e.Category, mcpSignIn(e), mcpPermissions(e),
		}
	}
	return PrintTable(out, headers, rows)
}

func mcpSignIn(e clients.McpCatalogEntry) string {
	if len(e.GrantsAllowed) == 0 {
		return TruncateList(e.AuthMethods, 3)
	}
	// auth_methods may already carry "oauth", so filter before prepending.
	options := []string{"oauth"}
	for _, m := range e.AuthMethods {
		if m != "oauth" {
			options = append(options, m)
		}
	}
	signIn := strings.Join(options, ", ")
	if selfRegisters, known := e.SelfRegisters[e.GrantsAllowed[0]]; known && !selfRegisters {
		signIn += " · needs admin setup"
	}
	return signIn
}

func mcpPermissions(e clients.McpCatalogEntry) string {
	if len(e.ScopesSupported) == 0 {
		return "—"
	}
	short := make([]string, 0, len(e.ScopesSupported))
	for _, s := range e.ScopesSupported {
		short = append(short, s)
	}
	listed := TruncateList(short, 3)
	if e.ScopesSelectable != nil && !*e.ScopesSelectable {
		return listed + " (all required)"
	}
	return listed
}

func PrintMcpDetail(out io.Writer, m *clients.McpSchema) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", m.Name)
	fmt.Fprintf(w, "Backend:\t%s\n", m.Backend)
	if m.StackId != nil && *m.StackId != "" {
		fmt.Fprintf(w, "Stack:\t%s\n", *m.StackId)
	}
	if m.Description != nil {
		fmt.Fprintf(w, "Description:\t%s\n", *m.Description)
	}
	if m.EndpointURL != nil {
		fmt.Fprintf(w, "Endpoint URL:\t%s\n", *m.EndpointURL)
	}
	if m.Transport != nil {
		fmt.Fprintf(w, "Transport:\t%s\n", *m.Transport)
	}
	if m.AuthType != nil {
		fmt.Fprintf(w, "Auth Type:\t%s\n", *m.AuthType)
	}
	if m.CatalogID != nil {
		fmt.Fprintf(w, "Catalog ID:\t%s\n", *m.CatalogID)
	}
	if m.Status != nil {
		fmt.Fprintf(w, "Status:\t%s\n", *m.Status)
	}
	if m.VerifyStatus != nil {
		fmt.Fprintf(w, "Verify Status:\t%s\n", *m.VerifyStatus)
	}
	fmt.Fprintf(w, "Credential Set:\t%t\n", m.HasCredential)
	if NeedsSignIn(*m) {
		fmt.Fprintf(
			w,
			"Tools:\t%d (needs a sign-in first — run 'iai mcps connect %s')\n",
			m.ToolCount,
			m.Name,
		)
	} else {
		fmt.Fprintf(w, "Tools:\t%d (see 'iai mcps tools %s')\n", m.ToolCount, m.Name)
	}
	if len(m.AttachedAgents) > 0 {
		fmt.Fprintf(w, "Attached Agents:\t%s\n", strings.Join(m.AttachedAgents, ", "))
	}
	return w.Flush()
}

// NeedsSignIn reports an MCP whose provider credential can only arrive through
// `iai mcps connect`. AuthType may be Platform's name for the choice ("oauth") or
// LiteLLM's name for the flow, which older rows still carry — both mean sign-in.
func NeedsSignIn(m clients.McpSchema) bool {
	if m.AuthType == nil || m.HasCredential {
		return false
	}
	switch *m.AuthType {
	case "oauth", "oauth2", "oauth_delegate":
		return true
	}
	return false
}

func PrintMcpTools(out io.Writer, tools []clients.McpToolSchema) error {
	if len(tools) == 0 {
		fmt.Fprintln(out, "No tools cached - run 'iai mcps verify' first.")
		return nil
	}
	fmt.Fprintf(out, "Tools (%d):\n", len(tools))
	for _, t := range tools {
		desc := ""
		if t.Description != nil {
			desc = *t.Description
		}
		if desc != "" {
			fmt.Fprintf(out, "  %s - %s\n", t.Name, desc)
		} else {
			fmt.Fprintf(out, "  %s\n", t.Name)
		}
	}
	return nil
}
