package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"gopkg.in/yaml.v3"
)

func PrintAgentList(out io.Writer, agents []clients.AgentOutput) error {
	if len(agents) == 0 {
		fmt.Fprintln(out, "No agents found.")
		return nil
	}

	headers := []string{"NAME", "REVISION", "STATUS", "ENDPOINT", "UPDATED"}
	rows := make([][]string, len(agents))
	for i, a := range agents {
		rows[i] = []string{
			a.Name,
			fmt.Sprintf("%d", a.Revision),
			a.Status,
			a.Endpoint,
			LocalTime(a.Updated),
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintAgentDescribe(out io.Writer, agent *clients.DescribeAgentResponse) error {
	fmt.Fprintf(out, "Name:      %s\n", agent.Name)
	fmt.Fprintf(out, "Id:        %s\n", agent.Id)
	fmt.Fprintf(out, "Version:   %s\n", agent.Version)
	fmt.Fprintf(out, "Revision:  %d\n", agent.Revision)
	fmt.Fprintf(out, "Status:    %s\n", agent.Status)

	if agent.Updated != "" {
		fmt.Fprintf(out, "Updated:   %s\n", LocalTime(agent.Updated))
	}
	if agent.Endpoint != "" {
		fmt.Fprintf(out, "Endpoint:  %s\n", agent.Endpoint)
	}

	if len(agent.Env) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Environment:")
		for _, e := range agent.Env {
			fmt.Fprintf(out, "  %s=%s\n", e.Name, e.Value)
		}
	}

	if len(agent.SecretRefs) > 0 {
		names := make([]string, len(agent.SecretRefs))
		for i, ref := range agent.SecretRefs {
			names[i] = ref.SecretName
		}
		fmt.Fprintf(out, "\nSecrets:   %s\n", strings.Join(names, ", "))
	}

	if agent.Schedule != nil {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Schedule:")
		if agent.Schedule.Uptime != "" {
			fmt.Fprintf(out, "  Uptime:   %s\n", agent.Schedule.Uptime)
		}
		if agent.Schedule.Downtime != "" {
			fmt.Fprintf(out, "  Downtime: %s\n", agent.Schedule.Downtime)
		}
		if agent.Schedule.Timezone != "" {
			fmt.Fprintf(out, "  Timezone: %s\n", agent.Schedule.Timezone)
		}
	}

	if agent.AgentConfig != nil {
		cfgBytes, err := yaml.Marshal(agent.AgentConfig)
		if err == nil && len(cfgBytes) > 0 {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Agent Config:")
			for _, line := range strings.Split(strings.TrimRight(string(cfgBytes), "\n"), "\n") {
				fmt.Fprintf(out, "  %s\n", line)
			}
		}
	}

	return nil
}

func PrintAgentCatalog(out io.Writer, agents []clients.CatalogAgent) error {
	if len(agents) == 0 {
		fmt.Fprintln(out, "No agents available.")
		return nil
	}

	headers := []string{"AGENT ID"}
	rows := make([][]string, len(agents))
	for i, a := range agents {
		rows[i] = []string{a.Id}
	}

	return PrintTable(out, headers, rows)
}

func PrintAgentVersions(out io.Writer, agentId string, versions []string) error {
	if len(versions) == 0 {
		fmt.Fprintf(out, "No versions found for agent %q.\n", agentId)
		return nil
	}

	headers := []string{"VERSION"}
	rows := make([][]string, len(versions))
	for i, v := range versions {
		rows[i] = []string{v}
	}

	return PrintTable(out, headers, rows)
}
