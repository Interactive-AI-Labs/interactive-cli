package cmd

import (
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	analyticsTimeRange string
	analyticsStep      string
	analyticsJSON      bool
	analyticsYAML      bool
	analyticsOrg       string
	analyticsProject   string
)

var agentAnalyticsCmd = &cobra.Command{
	Use:     "analytics [agent_name]",
	Short:   "Show an agent's CPU, memory, requests, status codes and latency",
	GroupID: groupAgentObservability,
	Long: "Show an agent's deployment analytics for a time range: the latest and peak value of each " +
		"CPU, memory, requests, status code and latency series. --json and --yaml print the five " +
		"raw responses keyed cpu_usage, memory_usage, status_codes, requests_per_domain and latency." +
		"\n\n" + selectedAgentHelp,
	Example: `  iai agents analytics my-agent
  iai agents analytics my-agent --time-range 24h
  iai agents analytics my-agent --time-range 6h --step 5m --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agent, _, err := agentArg(args, 1, analyticsOrg, analyticsProject)
		if err != nil {
			return err
		}
		if agent, err = agentName(agent); err != nil {
			return err
		}
		pCtx, apiClient, _, err := resolveProject(cmd.Context(), analyticsOrg, analyticsProject)
		if err != nil {
			return err
		}
		analytics, raw, err := apiClient.GetAgentAnalytics(
			cmd.Context(), pCtx.orgId, pCtx.projectId, agent,
			platform.AgentAnalyticsOptions{TimeRange: analyticsTimeRange, Step: analyticsStep},
		)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if analyticsJSON {
			return output.PrintRawJSON(out, raw)
		}
		if analyticsYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintAgentAnalytics(out, analytics)
	},
}

func init() {
	f := agentAnalyticsCmd.Flags()
	f.StringVar(&analyticsTimeRange, "time-range", "",
		"Time range: 5m, 15m, 1h, 6h, 24h or 7d (default 1h)")
	indexFlag(agentAnalyticsCmd, "time-range", "5m, 15m, 1h, 6h, 24h or 7d")
	f.StringVar(&analyticsStep, "step", "",
		"Resolution, e.g. 60s or 5m (defaults by time range)")
	f.BoolVar(&analyticsJSON, "json", false, "Output raw API responses as JSON")
	f.BoolVar(&analyticsYAML, "yaml", false, "Output raw API responses as YAML")
	f.StringVarP(&analyticsOrg, "organization", "o", "", "Organization name that owns the project")
	f.StringVarP(&analyticsProject, "project", "p", "", "Project name")
	agentAnalyticsCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	agentsCmd.AddCommand(agentAnalyticsCmd)
}
