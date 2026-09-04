package output

import (
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

func PrintMcpList(out io.Writer, mcps []platform.McpSchema) error {
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
		credential := "none"
		if m.HasCredential {
			credential = "set"
		}
		stack := "-"
		if m.StackId != nil && *m.StackId != "" {
			stack = *m.StackId
		}
		rows[i] = []string{
			m.Name,
			string(m.Backend),
			status,
			verify,
			strconv.Itoa(m.ToolCount),
			credential,
			stack,
		}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpCatalog(out io.Writer, entries []platform.McpCatalogEntry) error {
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

func mcpSignIn(e platform.McpCatalogEntry) string {
	// auth_methods is what a create is checked against: a populated list without
	// "oauth" refuses the sign-in whatever grants the provider advertises, and an
	// empty list is unrestricted.
	oauthAccepted := len(e.AuthMethods) == 0 || slices.Contains(e.AuthMethods, "oauth")
	if len(e.GrantsAllowed) == 0 || !oauthAccepted {
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

func mcpPermissions(e platform.McpCatalogEntry) string {
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

func PrintMcpDetail(out io.Writer, m *platform.McpSchema) error {
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
func NeedsSignIn(m platform.McpSchema) bool {
	if m.AuthType == nil || m.HasCredential {
		return false
	}
	switch *m.AuthType {
	case "oauth", "oauth2", "oauth_delegate":
		return true
	}
	return false
}

func PrintMcpTools(
	out io.Writer,
	backend platform.McpBackend,
	tools []platform.McpToolSchema,
) error {
	if len(tools) == 0 {
		// `verify` refuses an internal MCP, so only an external one can be sent there.
		if backend == platform.McpBackendInternal {
			fmt.Fprintln(out, "No tools cached - an internal mcp verifies itself when its "+
				"workload becomes ready; redeploy it to run that again.")
		} else {
			fmt.Fprintln(out, "No tools cached - run 'iai mcps verify' first.")
		}
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
