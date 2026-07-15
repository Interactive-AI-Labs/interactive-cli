package output

import (
	"fmt"
	"io"
	"sort"
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

func PrintMcpCatalog(out io.Writer, entries []clients.McpCatalogEntry) error {
	if len(entries) == 0 {
		fmt.Fprintln(out, "No catalog entries found.")
		return nil
	}
	headers := []string{"ID", "NAME", "CATEGORY", "TYPE", "AUTH"}
	rows := make([][]string, len(entries))
	for i, e := range entries {
		rows[i] = []string{e.ID, e.Name, e.Category, e.Type, TruncateList(e.AuthMethods, 3)}
	}
	return PrintTable(out, headers, rows)
}

func PrintMcpDetail(out io.Writer, m *clients.DescribeMcpResponse) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", m.Name)
	fmt.Fprintf(w, "Type:\t%s\n", m.Type)
	fmt.Fprintf(w, "External URL:\t%s\n", m.EndpointURL)
	fmt.Fprintf(w, "Transport:\t%s\n", m.Transport)
	fmt.Fprintf(w, "Slug:\t%s\n", m.Slug)
	if m.CatalogID != "" {
		fmt.Fprintf(w, "Catalog ID:\t%s\n", m.CatalogID)
	}
	if m.Status != "" {
		fmt.Fprintf(w, "Status:\t%s\n", m.Status)
	}
	if m.AuthType != "" {
		fmt.Fprintf(w, "Auth Type:\t%s\n", m.AuthType)
	}
	if m.AuthHeader != "" {
		fmt.Fprintf(w, "Auth Header:\t%s\n", m.AuthHeader)
	}
	if m.AuthHeaderPrefix != "" {
		fmt.Fprintf(w, "Auth Header Prefix:\t%s\n", m.AuthHeaderPrefix)
	}
	if len(m.Headers) > 0 {
		names := make([]string, 0, len(m.Headers))
		for k := range m.Headers {
			names = append(names, k)
		}
		sort.Strings(names)
		pairs := make([]string, len(names))
		for i, k := range names {
			pairs[i] = k + "=" + m.Headers[k]
		}
		fmt.Fprintf(w, "Extra Headers:\t%s\n", strings.Join(pairs, ", "))
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
	fmt.Fprintf(w, "Tools:\t%d (see 'iai mcps tools %s')\n", m.Verify.ToolCount, m.Name)

	return w.Flush()
}

// PrintMcpTools lists an mcp's cached tools with their descriptions, plus the
// names-level diff vs the previous verify snapshot when one is recorded.
func PrintMcpTools(
	out io.Writer,
	tools []map[string]any,
	added, removed []string,
	changedFrom string,
) error {
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
	if len(added) > 0 || len(removed) > 0 {
		fmt.Fprintf(out, "\nChanged since revision %s:", changedFrom)
		for _, n := range added {
			fmt.Fprintf(out, " +%s", n)
		}
		for _, n := range removed {
			fmt.Fprintf(out, " -%s", n)
		}
		fmt.Fprintln(out)
	}
	return nil
}
