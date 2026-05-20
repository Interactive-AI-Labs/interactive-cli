package output

import (
	"encoding/json"
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

	headers := []string{"NAME", "REVISION", "STATUS", "UPDATED"}
	rows := make([][]string, len(agents))
	for i, a := range agents {
		rows[i] = []string{
			a.Name,
			fmt.Sprintf("%d", a.Revision),
			a.Status,
			LocalTime(a.Updated),
		}
	}

	return PrintTable(out, headers, rows)
}

func PrintAgentDescribe(out io.Writer, agent *clients.DescribeAgentResponse) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", agent.Name)
	if agent.StackId != "" {
		fmt.Fprintf(w, "Stack Id:\t%s\n", agent.StackId)
	}
	fmt.Fprintf(w, "Id:\t%s\n", agent.Id)
	fmt.Fprintf(w, "Version:\t%s\n", agent.Version)
	fmt.Fprintf(w, "Revision:\t%d\n", agent.Revision)
	fmt.Fprintf(w, "Status:\t%s\n", agent.Status)
	if agent.Message != "" {
		fmt.Fprintf(w, "Message:\t%s\n", agent.Message)
	}

	if agent.Updated != "" {
		fmt.Fprintf(w, "Updated:\t%s\n", LocalTime(agent.Updated))
	}
	if agent.Endpoint != "" {
		fmt.Fprintf(w, "Endpoint:\t%s\n", agent.Endpoint)
	}
	if len(agent.Env) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Environment:")
		for _, e := range agent.Env {
			fmt.Fprintf(w, "  %s=%s\n", e.Name, e.Value)
		}
	}

	if len(agent.SecretRefs) > 0 {
		names := make([]string, len(agent.SecretRefs))
		for i, ref := range agent.SecretRefs {
			names[i] = ref.SecretName
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Secrets:\t%s\n", strings.Join(names, ", "))
	}

	if agent.Schedule != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Schedule:")
		if agent.Schedule.Uptime != "" {
			fmt.Fprintf(w, "  Uptime:\t%s\n", agent.Schedule.Uptime)
		}
		if agent.Schedule.Downtime != "" {
			fmt.Fprintf(w, "  Downtime:\t%s\n", agent.Schedule.Downtime)
		}
		if agent.Schedule.Timezone != "" {
			fmt.Fprintf(w, "  Timezone:\t%s\n", agent.Schedule.Timezone)
		}
	}

	if agent.AgentConfig != nil {
		cfgBytes, err := yaml.Marshal(agent.AgentConfig)
		if err == nil && len(cfgBytes) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Agent Config:")
			for _, line := range strings.Split(strings.TrimRight(string(cfgBytes), "\n"), "\n") {
				fmt.Fprintf(w, "  %s\n", line)
			}
		}
	}

	return w.Flush()
}

func PrintAgentRevisions(out io.Writer, revisions []clients.RevisionMeta) error {
	if len(revisions) == 0 {
		fmt.Fprintln(out, "No revisions found.")
		return nil
	}

	latest := 0
	for _, r := range revisions {
		if r.Revision > latest {
			latest = r.Revision
		}
	}

	headers := []string{"", "REVISION", "UPDATED"}
	rows := make([][]string, len(revisions))
	for i, r := range revisions {
		marker := ""
		if r.Revision == latest {
			marker = "*"
		}
		rows[i] = []string{marker, fmt.Sprintf("%d", r.Revision), LocalTime(r.Updated)}
	}

	return PrintTable(out, headers, rows)
}

func PrintAgentRevision(out io.Writer, rev *clients.AgentRevisionResponse) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Revision:\t%d\n", rev.Revision)
	if rev.StackId != "" {
		fmt.Fprintf(w, "Stack Id:\t%s\n", rev.StackId)
	}
	fmt.Fprintf(w, "Status:\t%s\n", rev.Status)
	if rev.Updated != "" {
		fmt.Fprintf(w, "Updated:\t%s\n", LocalTime(rev.Updated))
	}

	fmt.Fprintf(w, "Id:\t%s\n", rev.Id)
	fmt.Fprintf(w, "Version:\t%s\n", rev.Version)

	if rev.Endpoint != "" {
		fmt.Fprintf(w, "Endpoint:\t%s\n", rev.Endpoint)
	}
	if len(rev.Env) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Environment:")
		for _, e := range rev.Env {
			fmt.Fprintf(w, "  %s=%s\n", e.Name, e.Value)
		}
	}

	if len(rev.SecretRefs) > 0 {
		names := make([]string, len(rev.SecretRefs))
		for i, ref := range rev.SecretRefs {
			names[i] = ref.SecretName
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Secrets:\t%s\n", strings.Join(names, ", "))
	}

	if rev.Schedule != nil {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Schedule:")
		if rev.Schedule.Uptime != "" {
			fmt.Fprintf(w, "  Uptime:\t%s\n", rev.Schedule.Uptime)
		}
		if rev.Schedule.Downtime != "" {
			fmt.Fprintf(w, "  Downtime:\t%s\n", rev.Schedule.Downtime)
		}
		if rev.Schedule.Timezone != "" {
			fmt.Fprintf(w, "  Timezone:\t%s\n", rev.Schedule.Timezone)
		}
	}

	if rev.AgentConfig != nil {
		cfgBytes, err := yaml.Marshal(rev.AgentConfig)
		if err == nil && len(cfgBytes) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Agent Config:")
			for _, line := range strings.Split(strings.TrimRight(string(cfgBytes), "\n"), "\n") {
				fmt.Fprintf(w, "  %s\n", line)
			}
		}
	}

	return w.Flush()
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

func PrintCompatibilityMatrix(
	out io.Writer,
	matrix []clients.CompatibilityEntry,
	asJSON bool,
) error {
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(matrix)
	}

	if len(matrix) == 0 {
		fmt.Fprintln(out, "No compatibility data available.")
		return nil
	}

	headers := []string{"AGENT VERSION", "SCHEMA VERSION"}
	rows := make([][]string, len(matrix))
	for i, e := range matrix {
		rows[i] = []string{e.AgentVersion, e.SchemaVersion}
	}

	return PrintTable(out, headers, rows)
}
