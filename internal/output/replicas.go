package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintReplicaList(out io.Writer, replicas []clients.ReplicaInfo) error {
	if len(replicas) == 0 {
		fmt.Fprintln(out, "No replicas found.")
		return nil
	}

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
			LocalTime(r.StartTime),
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintReplicaDescribe(out io.Writer, status *clients.ReplicaStatus) error {
	readyStr := "No"
	if status.Ready {
		readyStr = "Yes"
	}

	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", status.Name)
	fmt.Fprintf(w, "Status:\t%s\n", status.Status)
	fmt.Fprintf(w, "Ready:\t%s\n", readyStr)
	if status.StartTime != "" {
		fmt.Fprintf(w, "Start Time:\t%s\n", LocalTime(status.StartTime))
	}
	fmt.Fprintf(w, "Restart Count:\t%d\n", status.RestartCount)

	if status.LastTerminationState != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Last Termination State:")
		fmt.Fprintf(w, "  Reason:\t%s\n", status.LastTerminationState.Reason)
		fmt.Fprintf(w, "  Exit Code:\t%d\n", status.LastTerminationState.ExitCode)
		if status.LastTerminationState.StartedAt != "" {
			fmt.Fprintf(w, "  Started At:\t%s\n", LocalTime(status.LastTerminationState.StartedAt))
		}
		if status.LastTerminationState.FinishedAt != "" {
			fmt.Fprintf(
				w,
				"  Finished At:\t%s\n",
				LocalTime(status.LastTerminationState.FinishedAt),
			)
		}
	}

	if status.Resources != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Resources:")
		fmt.Fprintf(w, "  CPU:\t%s\n", status.Resources.CPU)
		fmt.Fprintf(w, "  Memory:\t%s\n", status.Resources.Memory)
	}

	if status.Healthcheck != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Healthcheck:")
		fmt.Fprintf(w, "  Path:\t%s\n", status.Healthcheck.Path)
		fmt.Fprintf(w, "  Initial Delay (secs):\t%d\n", status.Healthcheck.InitialDelaySeconds)
	}

	if err := w.Flush(); err != nil {
		return err
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
				LocalTime(e.LastTimestamp),
			}
		}
		if err := PrintTable(out, headers, rows); err != nil {
			return err
		}
	}

	return nil
}
