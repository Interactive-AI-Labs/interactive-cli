package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
	"github.com/spf13/cobra"
)

func agentName(arg string) (string, error) {
	name := strings.TrimSpace(arg)
	if name == "" {
		return "", errors.New("agent name is required")
	}
	return name, nil
}

const selectedAgentHelp = "Without agent_name, the selected agent is used (see 'iai agents select')."

func selectedAgent(org, project string) (string, error) {
	cfg, err := files.LoadStackConfig(cfgFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to load config file: %w", err)
	}
	return session.NewSession(cfgDirName).SelectedAgent(cfg.Organization, org, cfg.Project, project)
}

// agentArg returns the agent and the remaining args, using the selected agent when it is left out.
func agentArg(args []string, n int, org, project string) (string, []string, error) {
	selected := ""
	if len(args) < n {
		var err error
		if selected, err = selectedAgent(org, project); err != nil {
			return "", nil, err
		}
	}
	return splitAgentArgs(args, n, selected)
}

// splitAgentArgs holds agentArg's decision: all n args means the first is the agent.
func splitAgentArgs(args []string, n int, selected string) (string, []string, error) {
	if len(args) == n {
		name, err := agentName(args[0])
		return name, args[1:], err
	}
	if selected == "" {
		return "", nil, fmt.Errorf(
			"accepts %d arg(s), received %d: pass the agent name or run 'iai agents select'",
			n, len(args),
		)
	}
	return selected, args, nil
}

// agentFlag returns the --agent value, or the selected agent when it is not passed.
func agentFlag(cmd *cobra.Command, agent, org, project string) (string, error) {
	if cmd.Flags().Changed("agent") {
		return agentName(agent)
	}
	return selectedAgent(org, project)
}

var agentTracesCmd = &cobra.Command{
	Use:     "traces [agent_name]",
	Short:   "List an agent's traces",
	GroupID: groupAgentObservability,
	Long:    "List an agent's traces.\n\n" + selectedAgentHelp,
	Example: `  iai agents traces my-agent
  iai agents traces my-agent --limit 20 --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if tracesAgent, _, err = agentArg(args, 1, tracesListOrg, tracesListProject); err != nil {
			return err
		}
		if tracesAgent, err = agentName(tracesAgent); err != nil {
			return err
		}
		return tracesListCmd.RunE(cmd, nil)
	},
}

var agentTraceSessionsCmd = &cobra.Command{
	Use:     "trace-sessions [agent_name]",
	Short:   "List an agent's trace sessions",
	GroupID: groupAgentObservability,
	Long:    "List an agent's trace sessions.\n\n" + selectedAgentHelp,
	Example: `  iai agents trace-sessions my-agent
  iai agents trace-sessions my-agent --limit 20 --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if sessionsAgent, _, err = agentArg(
			args,
			1,
			sessionsListOrg,
			sessionsListProject,
		); err != nil {
			return err
		}
		if sessionsAgent, err = agentName(sessionsAgent); err != nil {
			return err
		}
		return sessionsListCmd.RunE(cmd, nil)
	},
}

var agentObservationsCmd = &cobra.Command{
	Use:     "observations [agent_name]",
	Short:   "List an agent's observations",
	GroupID: groupAgentObservability,
	Long:    "List an agent's observations.\n\n" + selectedAgentHelp,
	Example: `  iai agents observations my-agent
  iai agents observations my-agent --limit 20 --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if obsListAgent, _, err = agentArg(args, 1, obsListOrg, obsListProject); err != nil {
			return err
		}
		if obsListAgent, err = agentName(obsListAgent); err != nil {
			return err
		}
		return obsListCmd.RunE(cmd, nil)
	},
}

func init() {
	// Flags bind the list commands' variables, which their RunE reads.
	f := agentTracesCmd.Flags()
	f.IntVar(&tracesPage, "page", 1, "Page number (starts at 1)")
	f.IntVar(&tracesLimit, "limit", 0, "Items per page (max 100)")
	indexFlag(agentTracesCmd, "limit", "max 100")
	f.StringVar(&tracesUserID, "user-id", "", "Filter by user ID")
	f.StringVar(&tracesName, "name", "", "Filter by trace name")
	f.StringVar(&tracesSessionID, "session-id", "", "Filter by session ID")
	f.StringVar(&tracesFromTimestamp, "from-timestamp", "",
		"Filter traces from this timestamp (ISO 8601, default: 7 days ago)")
	indexFlag(agentTracesCmd, "from-timestamp", "default: 7 days ago")
	f.StringVar(
		&tracesToTimestamp,
		"to-timestamp",
		"",
		"Filter traces to this timestamp (ISO 8601)",
	)
	addChoiceFlag(
		agentTracesCmd,
		&tracesOrderBy,
		"order-by",
		"",
		"Order by field",
		traceOrderFields,
	)
	addChoiceFlag(agentTracesCmd, &tracesOrder, "order", "desc", "Sort direction", sortOrders)
	f.StringArrayVar(&tracesTags, "tags", nil, "Filter by tags (repeatable)")
	f.StringVar(&tracesVersion, "version", "", "Filter by version")
	f.StringVar(&tracesRelease, "release", "", "Filter by release")
	f.StringArrayVar(&tracesEnvironment, "environment", nil, "Filter by environment (repeatable)")
	f.Float64Var(&tracesMinCost, "min-cost", 0, "Minimum total cost filter")
	f.Float64Var(&tracesMaxCost, "max-cost", 0, "Maximum total cost filter")
	f.Float64Var(&tracesMinLatency, "min-latency", 0, "Minimum latency filter (seconds)")
	f.Float64Var(&tracesMaxLatency, "max-latency", 0, "Maximum latency filter (seconds)")
	f.IntVar(&tracesMinTokens, "min-tokens", 0, "Minimum total tokens filter")
	f.IntVar(&tracesMaxTokens, "max-tokens", 0, "Maximum total tokens filter")
	f.StringVar(&tracesModel, "model", "", "Filter by model name")
	f.BoolVar(&tracesHasError, "has-error", false, "Filter traces with errors")
	addChoiceFlag(
		agentTracesCmd,
		&tracesLevel,
		"level",
		"",
		"Filter by aggregated level",
		traceLevels,
	)
	f.StringVar(&tracesSearch, "search", "", "Search in trace name (max 200 characters)")
	indexFlag(agentTracesCmd, "search", "max 200 characters")
	f.StringVar(&tracesFields, "fields", "core,metrics",
		"Field groups to include: core, io, metrics (comma-separated)")
	indexFlag(agentTracesCmd, "fields", "core, io, metrics")
	f.BoolVar(&tracesJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&tracesYAML, "yaml", false, "Output raw API response as YAML")
	f.BoolVar(&tracesListSummary, "summary", false,
		"Show conditions, decisions, tools, and results for each listed turn")
	f.StringSliceVar(
		&tracesColumns,
		"columns",
		nil,
		"Columns to display for table output only (comma-separated, default: id,name,timestamp,latency,cost,tags). Cannot be used with --json or --yaml.\nAvailable: id,name,timestamp,user_id,session_id,release,version,environment,public,latency,cost,tags,observation_count,input_tokens,output_tokens,total_tokens,level",
	)
	f.StringVarP(&tracesListOrg, "organization", "o", "", "Organization name that owns the project")
	f.StringVarP(&tracesListProject, "project", "p", "", "Project name")
	agentTracesCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	agentTracesCmd.MarkFlagsMutuallyExclusive("summary", "columns")
	agentTracesCmd.MarkFlagsMutuallyExclusive("summary", "fields")

	f = agentTraceSessionsCmd.Flags()
	f.StringVar(&sessionsFromTimestamp, "from-timestamp", "",
		"Filter sessions from this timestamp (ISO 8601, default: 7 days ago)")
	indexFlag(agentTraceSessionsCmd, "from-timestamp", "default: 7 days ago")
	f.StringVar(
		&sessionsToTimestamp,
		"to-timestamp",
		"",
		"Filter sessions to this timestamp (ISO 8601)",
	)
	f.IntVar(&sessionsPage, "page", 1, "Page number (starts at 1)")
	f.IntVar(&sessionsLimit, "limit", 0, "Items per page (max 100)")
	indexFlag(agentTraceSessionsCmd, "limit", "max 100")
	f.StringVar(&sessionsEnvironment, "environment", "", "Filter by environment")
	f.StringSliceVar(
		&sessionsColumns,
		"columns",
		nil,
		"Columns to display for table output only (comma-separated, default: id,created_at,environment,trace_count,duration_seconds,total_cost,total_tokens). Cannot be used with --json or --yaml.\nAvailable: id,created_at,updated_at,environment,user_id,trace_count,duration_seconds,total_cost,input_tokens,output_tokens,total_tokens",
	)
	f.BoolVar(&sessionsListJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&sessionsListYAML, "yaml", false, "Output raw API response as YAML")
	f.StringVarP(
		&sessionsListOrg,
		"organization",
		"o",
		"",
		"Organization name that owns the project",
	)
	f.StringVarP(&sessionsListProject, "project", "p", "", "Project name")
	agentTraceSessionsCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	f = agentObservationsCmd.Flags()
	f.BoolVar(&obsListJSON, "json", false, "Output raw API response as JSON")
	f.BoolVar(&obsListYAML, "yaml", false, "Output raw API response as YAML")
	f.StringSliceVar(
		&obsColumns,
		"columns",
		nil,
		"Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml\nDefault: id,trace_id,type,name,model,latency_ms,total_cost,total_tokens",
	)
	f.StringVarP(&obsListOrg, "organization", "o", "", "Organization name that owns the project")
	f.StringVarP(&obsListProject, "project", "p", "", "Project name")
	f.StringVar(&obsListFromTimestamp, "from-timestamp", "",
		"Filter observations from this timestamp (ISO 8601, default: 7 days ago)")
	indexFlag(agentObservationsCmd, "from-timestamp", "default: 7 days ago")
	f.StringVar(
		&obsListToTimestamp,
		"to-timestamp",
		"",
		"Filter observations to this timestamp (ISO 8601)",
	)
	f.StringVar(&obsListCursor, "cursor", "", "Cursor for pagination")
	f.IntVar(&obsListLimit, "limit", 0, "Items per page (max 100)")
	indexFlag(agentObservationsCmd, "limit", "max 100")
	f.StringVar(&obsListFields, "fields", "", "Field groups to include (comma-separated)")
	f.StringVar(&obsListType, "type", "", "Filter by observation type")
	f.StringVar(&obsListName, "name", "", "Filter by observation name")
	f.StringVar(&obsListLevel, "level", "", "Filter by level")
	f.StringVar(&obsListModel, "model", "", "Filter by model")
	f.StringVar(&obsListEnvironment, "environment", "", "Filter by environment")
	f.StringVar(
		&obsListParentObservationID,
		"parent-observation-id",
		"",
		"Filter by parent observation ID",
	)
	f.StringVar(&obsListVersion, "version", "", "Filter by version")
	f.StringVar(&obsListUserID, "user-id", "", "Filter by user ID")
	agentObservationsCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	agentsCmd.AddCommand(agentTracesCmd, agentTraceSessionsCmd, agentObservationsCmd)
}
