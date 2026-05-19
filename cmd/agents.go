package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

The --file flag takes a YAML file matching the agent_config schema — run
'iai agents schema' to see the expected shape. Pass the agent name as the
positional argument and id/version/env/secrets/endpoint/schedule via flags;
do not include them inside the file.

Examples:
  iai agents create chat-agent --id interactive-agent --version 0.0.1 --file agent-config.yaml
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

--file takes a YAML file matching the agent_config schema — run
'iai agents schema' to see the expected shape — and replaces the entire agent
config in full when provided (no per-field merge).

Lists (--env, --secret) replace the entire current list when provided — pass
every value you want to keep.

For schedules, passing --schedule-uptime auto-clears any existing downtime,
and --schedule-downtime auto-clears any existing uptime. Pass --schedule-timezone
alongside either to change the timezone.

Use --clear-env, --clear-secret, --clear-schedule, or --clear-stack-id to
remove those configurations entirely.

Examples:
  iai agents update chat-agent --version 0.0.3
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
	Long: `List agents in a specific project.

Examples:
  iai agents list
  iai agents list -p my-project`,
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

		return output.PrintAgentList(out, agents)
	},
}

var agentDescribeRevision int

var agentDescribeCmd = &cobra.Command{
	Use:     "describe <agent_name>",
	Aliases: []string{"desc"},
	Short:   "Describe an agent in detail",
	Long: `Show detailed information about a specific agent including its configuration.

Use --version to view a specific past version instead of the current state.

Examples:
  iai agents describe my-agent
  iai agents describe my-agent --version 3`,
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

		return output.PrintAgentDescribe(out, agent)
	},
}

var agentDeleteCmd = &cobra.Command{
	Use:   "delete <agent_name>",
	Short: "Delete an agent from a project",
	Long: `Delete an agent from a specific project.

Examples:
  iai agents delete my-agent`,
	Args: cobra.ExactArgs(1),
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
	Use:   "restart <agent_name>",
	Short: "Restart an agent in a project",
	Long: `Restart an agent in a specific project.

Examples:
  iai agents restart my-agent`,
	Args: cobra.ExactArgs(1),
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

var (
	agentLogsFollow    bool
	agentLogsSince     string
	agentLogsStartTime string
	agentLogsEndTime   string
)

var agentLogsCmd = &cobra.Command{
	Use:   "logs <agent_name>",
	Short: "Show logs for an agent",
	Long: `Show logs for an agent in a project.

Returns up to 5000 log entries in chronological order.

Examples:
  iai agents logs my-agent
  iai agents logs my-agent --follow
  iai agents logs my-agent --since 30m
  iai agents logs my-agent --start-time 2026-01-01T00:00:00Z --end-time 2026-01-01T01:00:00Z`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		agentName := strings.TrimSpace(args[0])

		ctx := cmd.Context()
		if agentLogsFollow {
			var stop func()
			ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
		}

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
		}

		logsResp, err := deployClient.GetAgentLogs(ctx, pCtx.orgId, pCtx.projectId, agentName, opts)
		if err != nil {
			return err
		}
		defer logsResp.Body.Close()

		meta := output.LogsMeta{
			Since:     logsResp.Since,
			Truncated: logsResp.Truncated,
			Empty:     logsResp.Empty,
		}
		err = output.PrintLogStream(out, logsResp.Body, true, meta)
		if agentLogsFollow && ctx.Err() != nil {
			return nil
		}
		return err
	},
}

var agentCatalogCmd = &cobra.Command{
	Use:   "catalog [agent_id]",
	Short: "List available agent types and versions",
	Long: `List agent types available in the catalog.

Without arguments, lists all available agent IDs.
With an agent ID argument, lists available versions for that agent.

Examples:
  iai agents catalog
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
	Long: `Fetch and display the current JSON Schema for the agent_config block.

Examples:
  iai agents schema`,
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

		result, err := apiClient.GetAgentSchema(cmd.Context())
		if err != nil {
			return err
		}

		fmt.Fprintf(out, "Schema version: %s\n\n", result.SchemaVersion)

		var indented bytes.Buffer
		if err := json.Indent(&indented, result.Schema, "", "  "); err != nil {
			return fmt.Errorf("failed to format schema: %w", err)
		}
		fmt.Fprintln(out, indented.String())

		return nil
	},
}

var agentRevisionsCmd = &cobra.Command{
	Use:     "revisions <agent_name>",
	Aliases: []string{"revs"},
	Short:   "List revisions of an agent",
	Long: `Show past revisions of an agent, sorted newest-first.
Up to 50 revisions are retained per agent.

Examples:
  iai agents revisions my-agent`,
	Args: cobra.ExactArgs(1),
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
	Use:   "diff <agent_name> <revision_a> <revision_b>",
	Short: "Compare two revisions of an agent",
	Long: `Show the differences between two revisions of an agent.

Examples:
  iai agents diff my-agent 1 3`,
	Args: cobra.ExactArgs(3),
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
--port when set, or an available OS-assigned port otherwise).

Examples:
  iai agents port-forward my-agent
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

	// Flags for "agents describe"
	agentDescribeCmd.Flags().
		StringVarP(&agentProject, "project", "p", "", "Project name")
	agentDescribeCmd.Flags().
		StringVarP(&agentOrganization, "organization", "o", "", "Organization name")
	agentDescribeCmd.Flags().
		IntVar(&agentDescribeRevision, "revision", 0, "Show a specific past revision instead of the current state")

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

	// Register commands
	rootCmd.AddCommand(agentsCmd)
	agentsCmd.AddCommand(agentCreateCmd)
	agentsCmd.AddCommand(agentUpdateCmd)
	agentsCmd.AddCommand(agentListCmd)
	agentsCmd.AddCommand(agentDescribeCmd)
	agentsCmd.AddCommand(agentDeleteCmd)
	agentsCmd.AddCommand(agentRestartCmd)
	agentsCmd.AddCommand(agentLogsCmd)
	agentsCmd.AddCommand(agentSchemaCmd)
	agentsCmd.AddCommand(agentCatalogCmd)
	agentsCmd.AddCommand(agentRevisionsCmd)
	agentsCmd.AddCommand(agentDiffCmd)
	agentsCmd.AddCommand(agentPortForwardCmd)
}
