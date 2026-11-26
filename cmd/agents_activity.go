package cmd

import (
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	activityPage   int
	activityLimit  int
	activitySearch string
	activityActors []string
	activityJSON   bool
	activityYAML   bool
)

var agentActivityCmd = &cobra.Command{
	Use:     "activity [agent_name]",
	Short:   "List an agent's deploys, context changes and secret edits",
	GroupID: groupAgentObservability,
	Long: "List an agent's activity: deploys, context changes and secret edits, " +
		"most recent first.\n\n" + selectedAgentHelp,
	Example: `  iai agents activity my-agent
  iai agents activity my-agent --search deploy --limit 50
  iai agents activity my-agent --actor user:123 --actor api_key:iai_ab --json`,
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
		activity, raw, err := apiClient.ListAgentActivity(
			cmd.Context(), pCtx.orgId, pCtx.projectId, agent,
			platform.AgentActivityOptions{
				Page:   activityPage,
				Limit:  activityLimit,
				Search: activitySearch,
				Actors: activityActors,
			},
		)
		if err != nil {
			return err
		}
		out := cmd.OutOrStdout()
		if activityJSON {
			return output.PrintRawJSON(out, raw)
		}
		if activityYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintAgentActivity(out, activity)
	},
}

func init() {
	f := agentActivityCmd.Flags()
	f.IntVar(&activityPage, "page", 1, "Page number (starts at 1)")
	f.IntVar(&activityLimit, "limit", 0, "Items per page (max 100)")
	indexFlag(agentActivityCmd, "limit", "max 100")
	f.StringVar(&activitySearch, "search", "",
		"Search in actor, action and data (max 200 characters)")
	indexFlag(agentActivityCmd, "search", "max 200 characters")
	f.StringArrayVar(
		&activityActors,
		"actor",
		nil,
		"Filter by actor: user:<id>, api_key:<prefix> or system:<component> (repeatable; --json lists them under actors)",
	)
	f.BoolVar(&activityJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&activityYAML, "yaml", false, "Output raw API response as YAML")
	f.StringVarP(&agentProject, "project", "p", "", "Project name")
	f.StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentActivityCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	agentsCmd.AddCommand(agentActivityCmd)
}
