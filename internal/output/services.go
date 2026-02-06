package output

import (
	"fmt"
	"io"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintServiceList(out io.Writer, services []clients.ServiceOutput) error {
	headers := []string{"NAME", "REVISION", "STATUS", "ENDPOINT", "UPDATED"}
	rows := make([][]string, len(services))
	for i, svc := range services {
		rows[i] = []string{
			svc.Name,
			fmt.Sprintf("%d", svc.Revision),
			svc.Status,
			svc.Endpoint,
			svc.Updated,
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintSyncResult(out io.Writer, created, updated, deleted []string) {
	if len(created) > 0 {
		fmt.Fprintf(out, "Created services: %s\n", strings.Join(created, ", "))
	}
	if len(updated) > 0 {
		fmt.Fprintf(out, "Updated services: %s\n", strings.Join(updated, ", "))
	}
	if len(deleted) > 0 {
		fmt.Fprintf(out, "Deleted services: %s\n", strings.Join(deleted, ", "))
	}
	if len(created) == 0 && len(updated) == 0 && len(deleted) == 0 {
		fmt.Fprintln(out, "No changes required; services already match config.")
	}
}
