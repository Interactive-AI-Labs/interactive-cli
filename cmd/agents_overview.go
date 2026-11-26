package cmd

import (
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	overviewWindow        string
	overviewFromTimestamp string
	overviewToTimestamp   string
	overviewJSON          bool
	overviewYAML          bool

	performanceAgents        []string
	performanceFromTimestamp string
	performanceToTimestamp   string
	performanceJSON          bool
	performanceYAML          bool
)

var agentOverviewCmd = &cobra.Command{
	Use:     "overview [agent_name]",
	Short:   "Show an agent's performance, cost, usage and errors",
	GroupID: groupAgentObservability,
	Long: "Show an agent's Overview: performance, cost, tokens, usage and errors for a window, " +
		"with the previous window of the same length.\n\n" + selectedAgentHelp,
	Example: `  iai agents overview my-agent
  iai agents overview my-agent --window 7d --json
  iai agents overview my-agent --window custom --from-timestamp 2026-11-01T00:00:00Z --to-timestamp 2026-11-08T00:00:00Z`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agent, _, err := agentArg(args, 1, agentOrganization, agentProject)
		if err != nil {
			return err
		}
		pCtx, apiClient, _, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}
		overview, raw, err := apiClient.GetAgentOverview(
			cmd.Context(), pCtx.orgId, pCtx.projectId, agent,
			platform.AgentOverviewOptions{
				Window: overviewWindow,
				From:   overviewFromTimestamp,
				To:     overviewToTimestamp,
				Groups: true,
			},
		)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if overviewJSON {
			return output.PrintRawJSON(out, raw)
		}
		if overviewYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintAgentOverview(out, overview)
	},
}

var agentPerformanceCmd = &cobra.Command{
	Use:     "performance",
	Short:   "Show traces, sessions, latency and cost per agent",
	GroupID: groupAgentObservability,
	Long: "Show the project's agent table: traces, sessions, latency and cost per agent, " +
		"and the total cost of the agents shown.\n\n" +
		"Without --agent, the selected agent is used (see 'iai agents select'); " +
		"with neither, every agent.",
	Example: `  iai agents performance
  iai agents performance --agent my-agent --agent other-agent
  iai agents performance --from-timestamp 2026-11-01T00:00:00Z --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		agents, err := performanceAgentsOrSelected(
			performanceAgents,
			agentOrganization,
			agentProject,
		)
		if err != nil {
			return err
		}
		body := platform.AgentPerformanceBody{
			FromTimestamp: performanceFromTimestamp,
			ToTimestamp:   performanceToTimestamp,
			Filters:       agentTagFilters(agents),
			Groups:        true,
			AgentIDs:      agents,
		}
		if body.FromTimestamp == "" {
			body.FromTimestamp = time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
		}
		if body.ToTimestamp == "" {
			body.ToTimestamp = time.Now().UTC().Format(time.RFC3339)
		}
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(), agentOrganization, agentProject,
		)
		if err != nil {
			return err
		}
		perf, raw, err := apiClient.GetAgentPerformance(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			body,
		)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if performanceJSON {
			return output.PrintRawJSON(out, raw)
		}
		if performanceYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintAgentPerformance(out, perf)
	},
}

func performanceAgentsOrSelected(agents []string, org, project string) ([]string, error) {
	if len(agents) > 0 {
		return agents, nil
	}
	selected, err := selectedAgent(org, project)
	if err != nil || selected == "" {
		return nil, err
	}
	return []string{selected}, nil
}

// agentTagFilters scopes the table to agents the way the project home does: by agent tag.
func agentTagFilters(agents []string) []platform.PerformanceFilter {
	if len(agents) == 0 {
		return nil
	}
	tags := make([]string, len(agents))
	for i, a := range agents {
		tags[i] = "agent:" + a
	}
	return []platform.PerformanceFilter{
		{Column: "Tags", Type: "arrayOptions", Operator: "any of", Value: tags},
	}
}

func init() {
	f := agentOverviewCmd.Flags()
	f.StringVar(
		&overviewWindow,
		"window",
		"24h",
		"Window length: 30m, 1h, 6h, 24h, 3d, 7d, 14d, 1mo, 3mo, all, or custom with --from-timestamp and --to-timestamp",
	)
	indexFlag(agentOverviewCmd, "window", "30m, 1h, 6h, 24h, 3d, 7d, 14d, 1mo, 3mo, all, or custom")
	f.StringVar(&overviewFromTimestamp, "from-timestamp", "",
		"Start of a custom window (ISO 8601)")
	f.StringVar(&overviewToTimestamp, "to-timestamp", "",
		"End of a custom window (ISO 8601)")
	f.BoolVar(&overviewJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&overviewYAML, "yaml", false, "Output raw API response as YAML")
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentOverviewCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	f = agentPerformanceCmd.Flags()
	f.StringArrayVar(&performanceAgents, "agent", nil,
		"Agent name (repeatable); defaults to the selected agent, else every agent")
	f.StringVar(&performanceFromTimestamp, "from-timestamp", "",
		"Show agents from this timestamp (ISO 8601, default: 7 days ago)")
	indexFlag(agentPerformanceCmd, "from-timestamp", "default: 7 days ago")
	f.StringVar(&performanceToTimestamp, "to-timestamp", "",
		"Show agents to this timestamp (ISO 8601, default: now)")
	indexFlag(agentPerformanceCmd, "to-timestamp", "default: now")
	f.BoolVar(&performanceJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&performanceYAML, "yaml", false, "Output raw API response as YAML")
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentPerformanceCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	agentsCmd.AddCommand(agentOverviewCmd, agentPerformanceCmd)
}
