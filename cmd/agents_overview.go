package cmd

import (
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
		if agent, err = agentName(agent); err != nil {
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
	agentsCmd.AddCommand(agentOverviewCmd)
}
