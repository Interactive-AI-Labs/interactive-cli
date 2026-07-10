package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	agentProject      string
	agentOrganization string

	agentId      string
	agentVersion string
	agentFile    string

	agentEndpoint   bool
	agentEnvVars    []string
	agentSecretRefs []string

	agentScheduleUptime   string
	agentScheduleDowntime string
	agentScheduleTimezone string

	agentClearSchedule bool
	agentClearEnv      bool
	agentClearSecret   bool
	agentClearStackId  bool

	agentStackId string

	agentListJSON     bool
	agentListYAML     bool
	agentDescribeJSON bool
	agentDescribeYAML bool
)

var (
	agentSchemaVersion string
	agentSchemaJSON    bool
	agentSchemaYAML    bool
)

var agentsCmd = &cobra.Command{
	Use:     "agents",
	Aliases: []string{"agent"},
	Short:   "Deploy AI agents with policies, routines, and tools",
	GroupID: groupInfra,
	Long:    `Manage deployment of agents to InteractiveAI projects.`,
}

var agentCreateCmd = &cobra.Command{
	Use:   "create <agent_name>",
	Short: "Create an agent in a project",
	Long: `Create an agent in a specific project.

The --file flag takes a YAML file matching the agent_config schema. Pass the
agent name as the positional argument and id/version/env/secrets/endpoint/schedule
via flags; do not include them inside the file.

The config schema depends on the agent version. Run
'iai agents compatibility-matrix' to find which schema version applies, then
'iai agents schema --schema-version <schema>' to see the expected fields.

Routines and policies referenced in the config must already exist in the project
and should be validated against the matching schema version (see --schema-version
on their create/update commands).`,
	Example: `  iai agents create chat-agent --id interactive-agent --version 0.0.1 --file agent-config.yaml
  iai agents create chat-agent --id interactive-agent --version 0.0.1 --file agent-config.yaml --endpoint
  iai agents create chat-agent --id interactive-agent --version 0.0.1 --file agent-config.yaml --secret api-keys --env LOG_LEVEL=info`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		reqBody, err := inputs.BuildAgentRequestBody(inputs.AgentInput{
			Id:               agentId,
			Version:          agentVersion,
			FilePath:         agentFile,
			Endpoint:         agentEndpoint,
			EnvVars:          agentEnvVars,
			SecretRefs:       agentSecretRefs,
			ScheduleUptime:   agentScheduleUptime,
			ScheduleDowntime: agentScheduleDowntime,
			ScheduleTimezone: agentScheduleTimezone,
			StackId:          agentStackId,
		})
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting agent creation request...")

		serverMessage, err := deployClient.CreateAgent(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
			reqBody,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var agentUpdateCmd = &cobra.Command{
	Use:   "update <agent_name>",
	Short: "Update an agent in a project",
	Long: `Update an agent in a specific project.

Only the flags you pass are applied; everything else is left at its current
value.

--file takes a YAML file matching the agent_config schema and replaces the
entire agent config in full when provided (no per-field merge). The config
schema depends on the agent version — run 'iai agents compatibility-matrix'
to find which schema version applies, then 'iai agents schema --schema-version <schema>'
to see the expected fields.

When upgrading to a new agent version with a different schema, update your
routines and policies first using --schema-version on their create/update
commands, then update the agent with the new config and version.

Lists (--env, --secret) replace the entire current list when provided — pass
every value you want to keep.

For schedules, passing --schedule-uptime auto-clears any existing downtime,
and --schedule-downtime auto-clears any existing uptime. Pass --schedule-timezone
alongside either to change the timezone.

Use --clear-env, --clear-secret, --clear-schedule, or --clear-stack-id to
remove those configurations entirely.`,
	Example: `  iai agents update chat-agent --version 0.0.3
  iai agents update chat-agent --file agent-config.yaml
  iai agents update chat-agent --endpoint=false
  iai agents update chat-agent --schedule-uptime "Mon-Fri 07:30-20:30" --schedule-timezone Europe/Berlin
  iai agents update chat-agent --clear-schedule
  iai agents update chat-agent --stack-id my-stack
  iai agents update chat-agent --clear-stack-id`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		patch, err := inputs.BuildAgentUpdatePatch(inputs.AgentInput{
			Id:               agentId,
			Version:          agentVersion,
			FilePath:         agentFile,
			Endpoint:         agentEndpoint,
			EnvVars:          agentEnvVars,
			SecretRefs:       agentSecretRefs,
			ScheduleUptime:   agentScheduleUptime,
			ScheduleDowntime: agentScheduleDowntime,
			ScheduleTimezone: agentScheduleTimezone,
			StackId:          agentStackId,
		}, agentClearEnv, agentClearSecret, agentClearSchedule, agentClearStackId, cmd.Flags().Changed)
		if err != nil {
			return err
		}
		if len(patch) == 0 {
			return fmt.Errorf("no fields to update; pass at least one flag")
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting agent update request...")

		serverMessage, err := deployClient.PatchAgent(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
			patch,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var agentListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List agents in a project",
	Long:    `List agents in a specific project.`,
	Example: `  iai agents list
  iai agents list -p my-project
  iai agents list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		agents, err := deployClient.ListAgents(cmd.Context(), pCtx.orgId, pCtx.projectId, "")
		if err != nil {
			return err
		}

		if agentListJSON {
			return output.PrintStructuredJSON(out, agents)
		}
		if agentListYAML {
			return output.PrintStructuredYAML(out, agents)
		}

		return output.PrintAgentList(out, agents)
	},
}

var agentDescribeRevision int

var agentDescribeCmd = &cobra.Command{
	Use:     "describe <agent_name>",
	Aliases: []string{"desc"},
	Short:   "Describe an agent in detail",
	Long: `Show detailed information about a specific agent including its configuration.

Use --version to view a specific past version instead of the current state.`,
	Example: `  iai agents describe my-agent
  iai agents describe my-agent --version 3
  iai agents describe my-agent --yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		if agentDescribeRevision > 0 {
			rev, err := deployClient.DescribeAgentRevision(
				cmd.Context(),
				pCtx.orgId,
				pCtx.projectId,
				agentName,
				agentDescribeRevision,
			)
			if err != nil {
				return err
			}
			if agentDescribeJSON {
				return output.PrintStructuredJSON(out, rev)
			}
			if agentDescribeYAML {
				return output.PrintStructuredYAML(out, rev)
			}
			return output.PrintAgentRevision(out, rev)
		}

		agent, err := deployClient.DescribeAgent(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
		)
		if err != nil {
			return err
		}

		if agentDescribeJSON {
			return output.PrintStructuredJSON(out, agent)
		}
		if agentDescribeYAML {
			return output.PrintStructuredYAML(out, agent)
		}

		return output.PrintAgentDescribe(out, agent)
	},
}

var agentDeleteCmd = &cobra.Command{
	Use:     "delete <agent_name>",
	Short:   "Delete an agent from a project",
	Long:    `Delete an agent from a specific project.`,
	Example: `  iai agents delete my-agent`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting agent deletion request...")

		serverMessage, err := deployClient.DeleteAgent(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var agentRestartCmd = &cobra.Command{
	Use:     "restart <agent_name>",
	Short:   "Restart an agent in a project",
	Long:    `Restart an agent in a specific project.`,
	Example: `  iai agents restart my-agent`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting agent restart request...")

		serverMessage, err := deployClient.RestartAgent(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var agentDeactivateCmd = &cobra.Command{
	Use:   "deactivate <agent_name>",
	Short: "Deactivate an agent in a project",
	Long: `Deactivate an agent, stopping all running instances. The current configuration
is preserved and will be restored when the agent is activated again.`,
	Example: `  iai agents deactivate my-agent`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			agentOrganization,
			agentProject,
		)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting agent deactivate request...")

		serverMessage, err := deployClient.DeactivateAgent(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var agentActivateCmd = &cobra.Command{
	Use:     "activate <agent_name>",
	Short:   "Activate a deactivated agent in a project",
	Long:    `Activate a deactivated agent, restoring it to its previous configuration.`,
	Example: `  iai agents activate my-agent`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			agentOrganization,
			agentProject,
		)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting agent activate request...")

		serverMessage, err := deployClient.ActivateAgent(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
		)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var (
	agentLogsFollow     bool
	agentLogsSince      string
	agentLogsStartTime  string
	agentLogsEndTime    string
	agentLogsRaw        bool
	agentLogsDecode     bool
	agentLogsFields     []string
	agentLogsAllFields  bool
	agentLogsTimestamps bool
	agentLogsLimit      int
)

var agentLogsCmd = &cobra.Command{
	Use:   "logs <agent_name>",
	Short: "Show logs for an agent",
	Long: `Show logs for an agent in a project.

Returns up to 1000 log entries in chronological order by default; use
--limit to request up to 5000.

Structured (JSON) logs are automatically formatted: the level and message
fields are extracted and displayed as "LEVEL message". Use --fields or
--all-fields to include additional top-level fields after the message. Use
--raw for exact server JSON, or --decode to decode embedded JSON strings into
nested JSON values.`,
	Example: `  iai agents logs my-agent
  iai agents logs my-agent --follow
  iai agents logs my-agent --since 30m
  iai agents logs my-agent --timestamps
  iai agents logs my-agent --start-time 2026-01-01T00:00:00Z --end-time 2026-01-01T01:00:00Z`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		timeout := 1 * time.Minute
		if agentLogsFollow {
			timeout = 0
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(), agentOrganization, agentProject,
			resolveOpts{deployTimeout: timeout},
		)
		if err != nil {
			return err
		}

		opts := clients.LogsOptions{
			Follow:    agentLogsFollow,
			Since:     agentLogsSince,
			StartTime: agentLogsStartTime,
			EndTime:   agentLogsEndTime,
			Limit:     agentLogsLimit,
		}

		ctx, stop := logFollowContext(cmd.Context(), agentLogsFollow)
		defer stop()

		logsResp, err := deployClient.GetAgentLogs(ctx, pCtx.orgId, pCtx.projectId, agentName, opts)
		if err != nil {
			return finishLogStream(cmd.ErrOrStderr(), agentLogsFollow, ctx, err)
		}
		defer logsResp.Body.Close()

		meta := output.LogsMeta{
			Start:     logsResp.Start,
			End:       logsResp.End,
			Truncated: logsResp.Truncated,
			Empty:     logsResp.Empty,
			Limit:     logsResp.Limit,
		}
		fmtOpts := output.LogFormatOptions{
			Raw:        agentLogsRaw || agentLogsDecode,
			Decode:     agentLogsDecode,
			Fields:     agentLogsFields,
			AllFields:  agentLogsAllFields,
			Timestamps: agentLogsTimestamps,
		}
		err = output.PrintLogStream(out, logsResp.Body, true, meta, fmtOpts)
		return finishLogStream(cmd.ErrOrStderr(), agentLogsFollow, ctx, err)
	},
}

var agentCatalogCmd = &cobra.Command{
	Use:   "catalog [agent_id]",
	Short: "List available agent types and versions",
	Long: `List agent types available in the catalog.

Without arguments, lists all available agent IDs.
With an agent ID argument, lists available versions for that agent.`,
	Example: `  iai agents catalog
  iai agents catalog interactive-agent`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			agents, err := deployClient.ListCatalogAgents(cmd.Context())
			if err != nil {
				return err
			}
			return output.PrintAgentCatalog(out, agents)
		}

		id := strings.TrimSpace(args[0])
		versions, err := deployClient.ListCatalogAgentVersions(cmd.Context(), id)
		if err != nil {
			return err
		}
		return output.PrintAgentVersions(out, id, versions)
	},
}

var agentSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Display the JSON Schema for agent configuration",
	Long: `Fetch and display the JSON Schema for the agent_config block.

Defaults to the latest schema version. Use --schema-version to fetch a specific
version (run 'iai agents compatibility-matrix' to see available versions).

Use --json or --yaml for structured schema output.`,
	Example: `  iai agents schema
  iai agents schema --schema-version 2.1.0
  iai agents schema --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		result, err := apiClient.GetAgentSchema(cmd.Context(), agentSchemaVersion)
		if err != nil {
			return err
		}

		if agentSchemaJSON {
			return output.PrintStructuredJSON(out, result)
		}
		if agentSchemaYAML {
			return output.PrintStructuredYAML(out, result)
		}

		return output.PrintSchemaPretty(out, result.Schema, result.SchemaVersion)
	},
}

var (
	agentCompatibilityMatrixJSON bool
	agentCompatibilityMatrixYAML bool
)

var agentCompatibilityMatrixCmd = &cobra.Command{
	Use:   "compatibility-matrix",
	Short: "Show agent version to schema version compatibility",
	Long: `Display the compatibility matrix between agent versions and schema versions.

Each agent version requires a specific config schema. Use this command to find
the schema version for your target agent version, then run
'iai agents schema --schema-version <schema>' to see the expected config fields.

Prompt types (routines, policies, etc.) also support versioned schemas — use
--schema-version on their create/update commands to validate against the
matching version.

By default, output is a formatted table. Use --json for machine-readable output.`,
	Example: `  iai agents compatibility-matrix
  iai agents compatibility-matrix --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		matrix, err := clients.GetAgentCompatibilityMatrix(
			cmd.Context(),
			hostname,
			defaultHTTPTimeout,
		)
		if err != nil {
			return err
		}

		if agentCompatibilityMatrixYAML {
			return output.PrintStructuredYAML(out, matrix)
		}

		return output.PrintCompatibilityMatrix(out, matrix, agentCompatibilityMatrixJSON)
	},
}

var agentRevisionsCmd = &cobra.Command{
	Use:     "revisions <agent_name>",
	Aliases: []string{"revs"},
	Short:   "List revisions of an agent",
	Long: `Show past revisions of an agent, sorted newest-first.
Up to 50 revisions are retained per agent.`,
	Example: `  iai agents revisions my-agent`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		revisions, err := deployClient.ListAgentRevisions(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
		)
		if err != nil {
			return err
		}

		return output.PrintAgentRevisions(out, revisions)
	},
}

var agentDiffCmd = &cobra.Command{
	Use:     "diff <agent_name> <revision_a> <revision_b>",
	Short:   "Compare two revisions of an agent",
	Long:    `Show the differences between two revisions of an agent.`,
	Example: `  iai agents diff my-agent 1 3`,
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		revA, err := inputs.ParseRevisionArg(args[1])
		if err != nil {
			return err
		}
		revB, err := inputs.ParseRevisionArg(args[2])
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		a, err := deployClient.DescribeAgentRevision(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
			revA,
		)
		if err != nil {
			return err
		}

		b, err := deployClient.DescribeAgentRevision(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
			revB,
		)
		if err != nil {
			return err
		}

		return output.PrintRevisionDiff(out, args[1], a, args[2], b)
	},
}

var (
	agentPFPort      int
	agentPFLocalPort int
)

var agentPortForwardCmd = &cobra.Command{
	Use:   "port-forward <agent_name>",
	Short: "Forward a local port to an agent",
	Long: `Open a local TCP listener and tunnel traffic through the deployment operator
to an agent running in the cluster.

The remote port defaults to the agent's configured port. Use --port to
override. Use --local-port to choose the local listening port (defaults to
--port when set, or an available OS-assigned port otherwise).`,
	Example: `  iai agents port-forward my-agent
  iai agents port-forward my-agent --port 8080
  iai agents port-forward my-agent --port 8080 --local-port 9090`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		agentName := strings.TrimSpace(args[0])
		localPort := agentPFLocalPort
		if localPort == 0 {
			localPort = agentPFPort
		}
		return runPortForward(cmd.Context(), portForwardOpts{
			resourceType: "agents",
			resourceName: agentName,
			remotePort:   agentPFPort,
			localPort:    localPort,
			org:          agentOrganization,
			project:      agentProject,
		})
	},
}

var agentLogFieldsSince string

var agentLogFieldsCmd = &cobra.Command{
	Use:   "log-fields <agent_name>",
	Short: "List available fields in structured logs",
	Long: `Scan recent logs and list the extra top-level fields present in structured (JSON) log entries.

Use the reported field names with 'iai agents logs --fields' to include them in output.`,
	Example: `  iai agents log-fields my-agent
  iai agents log-fields my-agent --since 1h`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		since := agentLogFieldsSince

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), agentOrganization, agentProject)
		if err != nil {
			return err
		}

		opts := clients.LogsOptions{Since: since}
		logsResp, err := deployClient.GetAgentLogs(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			agentName,
			opts,
		)
		if err != nil {
			return err
		}
		defer logsResp.Body.Close()

		if logsResp.Empty {
			output.PrintNoLogsFound(cmd.ErrOrStderr(), logsResp.Start, logsResp.End)
			return nil
		}

		fields, err := output.DiscoverLogFields(logsResp.Body)
		if err != nil {
			return err
		}
		if err := output.PrintLogFields(out, fields); err != nil {
			return err
		}
		if logsResp.Truncated {
			output.PrintLogFieldDiscoveryTruncationWarning(cmd.ErrOrStderr(), logsResp.Limit)
		}
		return nil
	},
}

func init() {
	// Flags for "agents create"
	agentCreateCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentCreateCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentCreateCmd.Flags().
		StringVar(&agentId, "id", "", "Agent type from the marketplace (e.g. interactive-agent)")
	agentCreateCmd.Flags().
		StringVar(&agentVersion, "version", "", "Agent image version to deploy (e.g. 0.0.1)")
	agentCreateCmd.Flags().
		StringVar(&agentFile, "file", "", "Path to YAML file matching the agent_config schema (run 'iai agents schema' to see it)")
	agentCreateCmd.Flags().
		BoolVar(&agentEndpoint, "endpoint", false, "Expose the agent at <agent-name>-<project-hash>.interactive.ai")
	agentCreateCmd.Flags().
		StringArrayVar(&agentEnvVars, "env", nil, "Environment variable (NAME=VALUE); can be repeated")
	agentCreateCmd.Flags().
		StringArrayVar(&agentSecretRefs, "secret", nil, "Secret to inject as environment variables; can be repeated")
	agentCreateCmd.Flags().
		StringVar(&agentScheduleUptime, "schedule-uptime", "", "When the agent should be running (mutually exclusive with --schedule-downtime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Example: 'Mon-Fri 07:30-20:30'")
	agentCreateCmd.Flags().
		StringVar(&agentScheduleDowntime, "schedule-downtime", "", "When the agent should be scaled down (mutually exclusive with --schedule-uptime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Example: 'Sat-Sun 00:00-24:00'")
	agentCreateCmd.Flags().
		StringVar(&agentScheduleTimezone, "schedule-timezone", "", "IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC); required with --schedule-uptime or --schedule-downtime")
	agentCreateCmd.Flags().
		StringVar(&agentStackId, "stack-id", "", "Stack ID to assign the agent to")
	_ = agentCreateCmd.MarkFlagRequired("id")
	_ = agentCreateCmd.MarkFlagRequired("version")
	_ = agentCreateCmd.MarkFlagRequired("file")

	// Flags for "agents update"
	agentUpdateCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentUpdateCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentUpdateCmd.Flags().
		StringVar(&agentId, "id", "", "Agent type from the marketplace (e.g. interactive-agent)")
	agentUpdateCmd.Flags().
		StringVar(&agentVersion, "version", "", "Agent image version to deploy (e.g. 0.0.1)")
	agentUpdateCmd.Flags().
		StringVar(&agentFile, "file", "", "Path to YAML file matching the agent_config schema (run 'iai agents schema' to see it)")
	agentUpdateCmd.Flags().
		BoolVar(&agentEndpoint, "endpoint", false, "Expose the agent at <agent-name>-<project-hash>.interactive.ai")
	agentUpdateCmd.Flags().
		StringArrayVar(&agentEnvVars, "env", nil, "Environment variable (NAME=VALUE); can be repeated")
	agentUpdateCmd.Flags().
		StringArrayVar(&agentSecretRefs, "secret", nil, "Secret to inject as environment variables; can be repeated")
	agentUpdateCmd.Flags().
		StringVar(&agentScheduleUptime, "schedule-uptime", "", "When the agent should be running (mutually exclusive with --schedule-downtime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Example: 'Mon-Fri 07:30-20:30'")
	agentUpdateCmd.Flags().
		StringVar(&agentScheduleDowntime, "schedule-downtime", "", "When the agent should be scaled down (mutually exclusive with --schedule-uptime). Format: comma-separated entries of DAY_FROM-DAY_TO HH:MM-HH:MM. Example: 'Sat-Sun 00:00-24:00'")
	agentUpdateCmd.Flags().
		StringVar(&agentScheduleTimezone, "schedule-timezone", "", "IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC); required with --schedule-uptime or --schedule-downtime")
	agentUpdateCmd.Flags().
		BoolVar(&agentClearEnv, "clear-env", false, "Remove all environment variables from the agent")
	agentUpdateCmd.Flags().
		BoolVar(&agentClearSecret, "clear-secret", false, "Remove all secret references from the agent")
	agentUpdateCmd.Flags().
		BoolVar(&agentClearSchedule, "clear-schedule", false, "Remove the schedule configuration from the agent")
	agentUpdateCmd.Flags().
		StringVar(&agentStackId, "stack-id", "", "Stack ID to assign the agent to")
	agentUpdateCmd.Flags().
		BoolVar(&agentClearStackId, "clear-stack-id", false, "Remove the agent from its stack")

	// Flags for "agents list"
	agentListCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentListCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentListCmd.Flags().BoolVar(&agentListJSON, "json", false, "Output raw API response as JSON")
	agentListCmd.Flags().BoolVar(&agentListYAML, "yaml", false, "Output raw API response as YAML")
	agentListCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// Flags for "agents describe"
	agentDescribeCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentDescribeCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentDescribeCmd.Flags().
		IntVar(&agentDescribeRevision, "revision", 0, "Show a specific past revision instead of the current state")
	agentDescribeCmd.Flags().
		BoolVar(&agentDescribeJSON, "json", false, "Output raw API response as JSON")
	agentDescribeCmd.Flags().
		BoolVar(&agentDescribeYAML, "yaml", false, "Output raw API response as YAML")
	agentDescribeCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// Flags for "agents delete"
	agentDeleteCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentDeleteCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")

	// Flags for "agents restart"
	agentRestartCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentRestartCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")

	// Flags for "agents deactivate"
	agentDeactivateCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentDeactivateCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")

	// Flags for "agents activate"
	agentActivateCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentActivateCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")

	// Flags for "agents logs"
	agentLogsCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentLogsCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentLogsCmd.Flags().
		BoolVarP(&agentLogsFollow, "follow", "f", false, "Stream new log entries as they arrive; mutually exclusive with --end-time")
	agentLogsCmd.Flags().
		StringVar(&agentLogsSince, "since", "", "Relative duration to look back (e.g. 30m, 1h, 3d, 1w); default 1h; max 72h; mutually exclusive with --start-time and --end-time")
	agentLogsCmd.Flags().
		StringVar(&agentLogsStartTime, "start-time", "", "Absolute RFC3339 start timestamp (e.g. 2026-02-24T10:00:00Z); mutually exclusive with --since; max 72h window")
	agentLogsCmd.Flags().
		StringVar(&agentLogsEndTime, "end-time", "", "Absolute RFC3339 end timestamp (e.g. 2026-02-24T12:00:00Z); requires --start-time; mutually exclusive with --since and --follow")
	agentLogsCmd.Flags().
		BoolVar(&agentLogsRaw, "raw", false, "Output exact server JSON lines without formatting")
	agentLogsCmd.Flags().
		BoolVar(&agentLogsDecode, "decode", false, "Decode embedded JSON strings into nested JSON values; outputs raw JSON")
	agentLogsCmd.Flags().
		StringSliceVar(&agentLogsFields, "fields", nil, "Additional fields to show after the message for structured (JSON) logs (e.g. --fields logger,pid); ignored for plain-text logs; use --raw for exact server JSON")
	agentLogsCmd.Flags().
		BoolVar(&agentLogsAllFields, "all-fields", false, "Show all extra top-level fields from structured (JSON) logs after the message")
	agentLogsCmd.Flags().
		BoolVar(&agentLogsTimestamps, "timestamps", false, "Include platform log timestamps")
	agentLogsCmd.Flags().
		IntVar(&agentLogsLimit, "limit", 0, "Maximum number of log entries to return (1-5000); defaults to 1000")
	agentLogsCmd.MarkFlagsMutuallyExclusive("raw", "fields")
	agentLogsCmd.MarkFlagsMutuallyExclusive("raw", "all-fields")
	agentLogsCmd.MarkFlagsMutuallyExclusive("decode", "fields")
	agentLogsCmd.MarkFlagsMutuallyExclusive("decode", "all-fields")
	agentLogsCmd.MarkFlagsMutuallyExclusive("fields", "all-fields")

	// Flags for "agents log-fields"
	agentLogFieldsCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentLogFieldsCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentLogFieldsCmd.Flags().
		StringVar(&agentLogFieldsSince, "since", "1h", "Relative duration to scan (e.g. 5m, 1h)")

	// Flags for "agents revisions"
	agentRevisionsCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentRevisionsCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")

	// Flags for "agents diff"
	agentDiffCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentDiffCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")

	// Flags for "agents port-forward"
	agentPortForwardCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentPortForwardCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentPortForwardCmd.Flags().
		IntVar(&agentPFPort, "port", 0, "Remote port on the agent (defaults to the agent's configured port)")
	agentPortForwardCmd.Flags().
		IntVar(&agentPFLocalPort, "local-port", 0, "Local port to listen on (defaults to the remote port)")

	// Flags for "agents schema"
	agentSchemaCmd.Flags().
		StringVar(&agentSchemaVersion, "schema-version", "", "Schema version to fetch (defaults to latest stable)")
	agentSchemaCmd.Flags().
		BoolVar(&agentSchemaJSON, "json", false, "Output schema response as JSON")
	agentSchemaCmd.Flags().
		BoolVar(&agentSchemaYAML, "yaml", false, "Output schema response as YAML")
	agentSchemaCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// Flags for "agents compatibility-matrix"
	agentCompatibilityMatrixCmd.Flags().
		BoolVar(&agentCompatibilityMatrixJSON, "json", false, "Output raw JSON instead of a formatted table")
	agentCompatibilityMatrixCmd.Flags().
		BoolVar(&agentCompatibilityMatrixYAML, "yaml", false, "Output structured YAML")
	agentCompatibilityMatrixCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	// Register commands
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.AddCommand(agentCreateCmd)
	agentsCmd.AddCommand(agentUpdateCmd)
	agentsCmd.AddCommand(agentListCmd)
	agentsCmd.AddCommand(agentDescribeCmd)
	agentsCmd.AddCommand(agentDeleteCmd)
	agentsCmd.AddCommand(agentRestartCmd)
	agentsCmd.AddCommand(agentDeactivateCmd)
	agentsCmd.AddCommand(agentActivateCmd)
	agentsCmd.AddCommand(agentLogsCmd)
	agentsCmd.AddCommand(agentSchemaCmd)
	agentsCmd.AddCommand(agentCatalogCmd)
	agentsCmd.AddCommand(agentRevisionsCmd)
	agentsCmd.AddCommand(agentDiffCmd)
	agentsCmd.AddCommand(agentPortForwardCmd)
	agentsCmd.AddCommand(agentCompatibilityMatrixCmd)
	agentsCmd.AddCommand(agentLogFieldsCmd)
}
