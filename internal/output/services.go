package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintServiceList(out io.Writer, services []clients.ServiceOutput) error {
	if len(services) == 0 {
		fmt.Fprintln(out, "No services found.")
		return nil
	}

	headers := []string{"NAME", "REVISION", "STATUS", "UPDATED"}
	rows := make([][]string, len(services))
	for i, svc := range services {
		rows[i] = []string{
			svc.Name,
			fmt.Sprintf("%d", svc.Revision),
			svc.Status,
			LocalTime(svc.Updated),
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintServiceDescribe(out io.Writer, svc *clients.DescribeServiceResponse) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", svc.Name)
	fmt.Fprintf(w, "Revision:\t%d\n", svc.Revision)
	fmt.Fprintf(w, "Status:\t%s\n", svc.Status)
	if svc.Message != "" {
		fmt.Fprintf(w, "Message:\t%s\n", svc.Message)
	}

	if svc.Updated != "" {
		fmt.Fprintf(w, "Updated:\t%s\n", LocalTime(svc.Updated))
	}

	fmt.Fprintf(w, "Port:\t%d\n", svc.ServicePort)
	fmt.Fprintln(w, "Image:")
	fmt.Fprintf(w, "  Type:\t%s\n", svc.Image.Type)
	fmt.Fprintf(w, "  Name:\t%s\n", svc.Image.Name)
	fmt.Fprintf(w, "  Tag:\t%s\n", svc.Image.Tag)
	if svc.Image.Repository != "" {
		fmt.Fprintf(w, "  Repository:\t%s\n", svc.Image.Repository)
	}

	fmt.Fprintln(w, "Resources:")
	fmt.Fprintf(w, "  CPU:\t%s\n", svc.Resources.CPU)
	fmt.Fprintf(w, "  Memory:\t%s\n", svc.Resources.Memory)

	if svc.Autoscaling != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Autoscaling:")
		fmt.Fprintf(w, "  Min Replicas:\t%d\n", svc.Autoscaling.MinReplicas)
		fmt.Fprintf(w, "  Max Replicas:\t%d\n", svc.Autoscaling.MaxReplicas)
		if svc.Autoscaling.CPUPercentage != nil {
			fmt.Fprintf(w, "  CPU%%:\t%d\n", *svc.Autoscaling.CPUPercentage)
		}
		if svc.Autoscaling.MemoryPercentage != nil {
			fmt.Fprintf(w, "  Memory%%:\t%d\n", *svc.Autoscaling.MemoryPercentage)
		}
	} else {
		fmt.Fprintf(w, "Replicas:\t%d\n", svc.Replicas)
	}

	if svc.Endpoint != "" {
		fmt.Fprintf(w, "Endpoint:\t%s\n", svc.Endpoint)
	}

	if svc.StackId != "" {
		fmt.Fprintf(w, "Stack:\t%s\n", svc.StackId)
	}

	if svc.Healthcheck != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Healthcheck:")
		fmt.Fprintf(w, "  Path:\t%s\n", svc.Healthcheck.Path)
		if svc.Healthcheck.InitialDelaySeconds != nil {
			fmt.Fprintf(w, "  Initial Delay:\t%ds\n", *svc.Healthcheck.InitialDelaySeconds)
		}
	}

	if len(svc.Env) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Environment:")
		for _, e := range svc.Env {
			fmt.Fprintf(w, "  %s=%s\n", e.Name, e.Value)
		}
	}

	if len(svc.SecretRefs) > 0 {
		names := make([]string, len(svc.SecretRefs))
		for i, ref := range svc.SecretRefs {
			names[i] = ref.SecretName
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Secrets:\t%s\n", strings.Join(names, ", "))
	}

	if svc.Schedule != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Schedule:")
		if svc.Schedule.Uptime != "" {
			fmt.Fprintf(w, "  Uptime:\t%s\n", svc.Schedule.Uptime)
		}
		if svc.Schedule.Downtime != "" {
			fmt.Fprintf(w, "  Downtime:\t%s\n", svc.Schedule.Downtime)
		}
		if svc.Schedule.Timezone != "" {
			fmt.Fprintf(w, "  Timezone:\t%s\n", svc.Schedule.Timezone)
		}
	}

	return w.Flush()
}

func PrintSyncResult(out io.Writer, label string, created, updated, deleted, skipped []string) {
	if len(created) > 0 {
		fmt.Fprintf(out, "Created %s: %s\n", label, strings.Join(created, ", "))
	}
	if len(updated) > 0 {
		fmt.Fprintf(out, "Updated %s: %s\n", label, strings.Join(updated, ", "))
	}
	if len(deleted) > 0 {
		fmt.Fprintf(out, "Deleted %s: %s\n", label, strings.Join(deleted, ", "))
	}
	if len(skipped) > 0 {
		fmt.Fprintf(
			out,
			"Skipped %s (already exist, updates not supported): %s\n",
			label,
			strings.Join(skipped, ", "),
		)
	}
	if len(created) == 0 && len(updated) == 0 && len(deleted) == 0 && len(skipped) == 0 {
		fmt.Fprintf(out, "No changes required; %s already match config.\n", label)
	}
}
