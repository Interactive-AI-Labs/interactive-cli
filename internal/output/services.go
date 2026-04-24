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

func PrintServiceDescribe(out io.Writer, svc *clients.DescribeServiceResponse) error {
	fmt.Fprintf(out, "Name:        %s\n", svc.Name)
	fmt.Fprintf(out, "Project Id:  %s\n", svc.ProjectId)
	fmt.Fprintf(out, "Revision:    %d\n", svc.Revision)
	fmt.Fprintf(out, "Status:      %s\n", svc.Status)

	if svc.Updated != "" {
		fmt.Fprintf(out, "Updated:     %s\n", LocalTime(svc.Updated))
	}

	fmt.Fprintf(out, "Port:        %d\n", svc.ServicePort)
	fmt.Fprintf(out, "Image:\n")
	fmt.Fprintf(out, "  Type:       %s\n", svc.Image.Type)
	fmt.Fprintf(out, "  Name:       %s\n", svc.Image.Name)
	fmt.Fprintf(out, "  Tag:        %s\n", svc.Image.Tag)
	if svc.Image.Repository != "" {
		fmt.Fprintf(out, "  Repository: %s\n", svc.Image.Repository)
	}

	fmt.Fprintf(out, "Resources:\n")
	fmt.Fprintf(out, "  CPU:     %s\n", svc.Resources.CPU)
	fmt.Fprintf(out, "  Memory:  %s\n", svc.Resources.Memory)

	if svc.Autoscaling != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Autoscaling:")
		fmt.Fprintf(out, "  Min Replicas: %d\n", svc.Autoscaling.MinReplicas)
		fmt.Fprintf(out, "  Max Replicas: %d\n", svc.Autoscaling.MaxReplicas)
		if svc.Autoscaling.CPUPercentage != nil {
			fmt.Fprintf(out, "  CPU%%:         %d\n", *svc.Autoscaling.CPUPercentage)
		}
		if svc.Autoscaling.MemoryPercentage != nil {
			fmt.Fprintf(out, "  Memory%%:      %d\n", *svc.Autoscaling.MemoryPercentage)
		}
	} else {
		fmt.Fprintf(out, "Replicas:    %d\n", svc.Replicas)
	}

	if svc.Endpoint != "" {
		fmt.Fprintf(out, "Endpoint:    %s\n", svc.Endpoint)
	}

	if svc.StackId != "" {
		fmt.Fprintf(out, "Stack Id:    %s\n", svc.StackId)
	}

	if svc.Healthcheck != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Healthcheck:")
		fmt.Fprintf(out, "  Path:           %s\n", svc.Healthcheck.Path)
		if svc.Healthcheck.InitialDelaySeconds != nil {
			fmt.Fprintf(out, "  Initial Delay:  %ds\n", *svc.Healthcheck.InitialDelaySeconds)
		}
	}

	if len(svc.Env) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Environment:")
		for _, e := range svc.Env {
			fmt.Fprintf(out, "  %s=%s\n", e.Name, e.Value)
		}
	}

	if len(svc.SecretRefs) > 0 {
		names := make([]string, len(svc.SecretRefs))
		for i, ref := range svc.SecretRefs {
			names[i] = ref.SecretName
		}
		fmt.Fprintf(out, "\nSecrets:     %s\n", strings.Join(names, ", "))
	}

	if svc.Schedule != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Schedule:")
		if svc.Schedule.Uptime != "" {
			fmt.Fprintf(out, "  Uptime:   %s\n", svc.Schedule.Uptime)
		}
		if svc.Schedule.Downtime != "" {
			fmt.Fprintf(out, "  Downtime: %s\n", svc.Schedule.Downtime)
		}
		if svc.Schedule.Timezone != "" {
			fmt.Fprintf(out, "  Timezone: %s\n", svc.Schedule.Timezone)
		}
	}

	return nil
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
