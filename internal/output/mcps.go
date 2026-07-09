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
	headers := []string{"NAME", "TYPE", "STATUS", "VERIFY", "TOOLS", "CATALOG", "UPDATED"}
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
			orDash(m.CatalogID),
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
	fmt.Fprintf(w, "Revision:\t%d\n", m.Revision)
	fmt.Fprintf(w, "Updated:\t%s\n", LocalTime(m.Updated))

	fmt.Fprintf(w, "Verify Status:\t%s\n", orDash(m.Verify.Status))
	if m.Verify.VerifiedAt != "" {
		fmt.Fprintf(w, "Last Verified:\t%s\n", LocalTime(m.Verify.VerifiedAt))
	}
	if m.Verify.Error != "" {
		fmt.Fprintf(w, "Verify Error:\t%s\n", m.Verify.Error)
	}
	if m.Verify.Version != "" {
		fmt.Fprintf(w, "Verified Image Version:\t%s\n", m.Verify.Version)
	}
	if len(m.ToolVersions) > 0 {
		fmt.Fprintf(w, "Cached Versions:\t%s\n", strings.Join(m.ToolVersions, ", "))
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if len(m.Tools) > 0 {
		fmt.Fprintf(out, "\nTools (%d):\n", len(m.Tools))
		for _, t := range m.Tools {
			name, _ := t["name"].(string)
			desc, _ := t["description"].(string)
			if desc != "" {
				fmt.Fprintf(out, "  %s — %s\n", name, desc)
			} else {
				fmt.Fprintf(out, "  %s\n", name)
			}
		}
	}
	return nil
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
