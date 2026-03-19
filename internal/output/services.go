package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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
			LocalTime(svc.Updated),
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintSyncResult(out io.Writer, label string, created, updated, deleted []string) {
	if len(created) > 0 {
		fmt.Fprintf(out, "Created %s: %s\n", label, strings.Join(created, ", "))
	}
	if len(updated) > 0 {
		fmt.Fprintf(out, "Updated %s: %s\n", label, strings.Join(updated, ", "))
	}
	if len(deleted) > 0 {
		fmt.Fprintf(out, "Deleted %s: %s\n", label, strings.Join(deleted, ", "))
	}
	if len(created) == 0 && len(updated) == 0 && len(deleted) == 0 {
		fmt.Fprintf(out, "No changes required; %s already match config.\n", label)
	}
}
