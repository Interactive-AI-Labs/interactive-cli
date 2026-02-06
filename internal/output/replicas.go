package output

import (
	"fmt"
	"io"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintReplicaDescribe(out io.Writer, status *clients.ReplicaStatus) error {
	readyStr := "No"
	if status.Ready {
		readyStr = "Yes"
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Name:          %s\n", status.Name)
	fmt.Fprintf(out, "Status:        %s\n", status.Status)
	fmt.Fprintf(out, "Ready:         %s\n", readyStr)
	if status.StartTime != "" {
		fmt.Fprintf(out, "Start Time:    %s\n", status.StartTime)
	}
	fmt.Fprintf(out, "Restart Count: %d\n", status.RestartCount)

	if status.Resources != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Resources:")
		fmt.Fprintf(out, "  CPU:    %s\n", status.Resources.CPU)
		fmt.Fprintf(out, "  Memory: %s\n", status.Resources.Memory)
	}

	if status.Healthcheck != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Healthcheck:")
		fmt.Fprintf(out, "  Path:                 %s\n", status.Healthcheck.Path)
		fmt.Fprintf(out, "  Initial Delay (secs): %d\n", status.Healthcheck.InitialDelaySeconds)
	}

	if len(status.Events) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Events:")
		headers := []string{"TYPE", "REASON", "COUNT", "MESSAGE", "LAST SEEN"}
		rows := make([][]string, len(status.Events))
		for i, e := range status.Events {
			rows[i] = []string{
				e.Type,
				e.Reason,
				fmt.Sprintf("%d", e.Count),
				e.Message,
				e.LastTimestamp,
			}
		}
		if err := PrintTable(out, headers, rows); err != nil {
			return err
		}
	}

	return nil
}
