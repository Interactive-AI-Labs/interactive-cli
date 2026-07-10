package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintMcpList(out io.Writer, mcps []clients.McpOutput) error {
	if len(mcps) == 0 {
		fmt.Fprintln(out, "No mcps found.")
		return nil
	}
	headers := []string{"NAME", "TYPE", "STATUS", "VERIFY", "TOOLS", "UPDATED"}
	rows := make([][]string, len(mcps))
	for i, m := range mcps {
		status := m.Status
		if status == "" {
			status = "-" // external: no workload
		}
		verify := m.Verify.Status
		if verify == "" {
			verify = "never"
		}
		rows[i] = []string{
			m.Name,
			m.Type,
			status,
			verify,
			fmt.Sprintf("%d", m.Verify.ToolCount),
			LocalTime(m.Updated),
		}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpDetail(out io.Writer, m *clients.DescribeMcpResponse) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", m.Name)
	fmt.Fprintf(w, "Type:\t%s\n", m.Type)
	fmt.Fprintf(w, "Endpoint:\t%s\n", m.EndpointURL)
	fmt.Fprintf(w, "Transport:\t%s\n", m.Transport)
	fmt.Fprintf(w, "Slug:\t%s\n", m.Slug)
	if m.CatalogID != "" {
		fmt.Fprintf(w, "Catalog ID:\t%s\n", m.CatalogID)
	}
	if m.Status != "" {
		fmt.Fprintf(w, "Status:\t%s\n", m.Status)
	}
	fmt.Fprintf(w, "Credential Set:\t%t\n", m.HasCredential)
	if len(m.SecretRefs) > 0 {
		names := make([]string, len(m.SecretRefs))
		for i, ref := range m.SecretRefs {
			names[i] = ref.SecretName
		}
		fmt.Fprintf(w, "Secrets:\t%s\n", strings.Join(names, ", "))
	}
	if len(m.AttachedAgents) > 0 {
		fmt.Fprintf(w, "Attached Agents:\t%s\n", strings.Join(m.AttachedAgents, ", "))
	}
	fmt.Fprintf(w, "Revision:\t%d\n", m.Revision)
	fmt.Fprintf(w, "Updated:\t%s\n", LocalTime(m.Updated))

	verifyStatus := m.Verify.Status
	if verifyStatus == "" {
		verifyStatus = "-"
	}
	fmt.Fprintf(w, "Verify Status:\t%s\n", verifyStatus)
	if m.Verify.VerifiedAt != "" {
		fmt.Fprintf(w, "Last Verified:\t%s\n", LocalTime(m.Verify.VerifiedAt))
	}
	if m.Verify.Error != "" {
		fmt.Fprintf(w, "Verify Error:\t%s\n", m.Verify.Error)
	}
	fmt.Fprintf(w, "Tools:\t%d (see 'iai mcps tools get %s', 'iai mcps tools revisions %s')\n", len(m.Tools), m.Name, m.Name)

	return w.Flush()
}

// PrintMcpTools lists an mcp's cached tools with their descriptions — the
// current revision's (DescribeMcp) or a past one's (DescribeMcpToolRevision),
// both of which carry a Tools []map[string]any field.
func PrintMcpTools(out io.Writer, tools []map[string]any) error {
	if len(tools) == 0 {
		fmt.Fprintln(out, "No tools cached — run 'iai mcps verify' first.")
		return nil
	}
	fmt.Fprintf(out, "Tools (%d):\n", len(tools))
	for _, t := range tools {
		name, _ := t["name"].(string)
		desc, _ := t["description"].(string)
		if desc != "" {
			fmt.Fprintf(out, "  %s — %s\n", name, desc)
		} else {
			fmt.Fprintf(out, "  %s\n", name)
		}
	}
	return nil
}
