package output

import (
	"fmt"
	"io"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintReplicaList(out io.Writer, replicas []clients.ReplicaInfo) error {
	headers := []string{"NAME", "STATUS", "CPU", "MEMORY", "STARTED"}
	rows := make([][]string, len(replicas))
	for i, r := range replicas {
		readinessLabel := "Not Ready"
		if r.Ready {
			readinessLabel = "Ready"
		}

		combinedStatus := strings.TrimSpace(r.Status)
		if combinedStatus == "" {
			combinedStatus = strings.TrimSpace(r.Phase)
		}
		if combinedStatus == "" {
			combinedStatus = "Unknown"
		}

		combinedStatus = fmt.Sprintf("%s [%s]", combinedStatus, readinessLabel)

		rows[i] = []string{
			r.Name,
			combinedStatus,
			r.CPU,
			r.Memory,
			r.StartTime,
		}
	}

	return PrintTable(out, headers, rows)
}

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
