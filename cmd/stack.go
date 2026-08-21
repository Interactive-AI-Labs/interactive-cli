package cmd

import (
	"fmt"
	"os"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/sync"
	"github.com/spf13/cobra"
)

var (
	stackSyncFile         string
	stackSyncProject      string
	stackSyncOrganization string
	stackSyncAllowDelete  []string
	stackSyncDryRun       bool

	stackGetStackID string
	stackGetFile    string
	stackGetOrg     string
	stackGetProject string
	stackGetJSON    bool
	stackGetYAML    bool

	stackDiffFile    string
	stackDiffStackID string
	stackDiffOrg     string
	stackDiffProject string
	stackDiffJSON    bool

	stackListJSON bool
	stackListOrg  string
	stackListProj string
)

var stackCmd = &cobra.Command{
	Use:     "stacks",
	Aliases: []string{"stack", "st"},
	Short:   "Declarative resource sync from config files",
	GroupID: groupInfra,
	Long:    `Manage stacks and their resources (services, agents, databases, mcps) from stack configuration files.`,
}

var stackSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync services, agents, databases, and mcps from a stack config file",
	Long: `Sync services, agents, databases, and mcps in a project from a stack configuration file.

Services, agents, databases, and mcps are created and updated to match the config
file. Resources the config file no longer mentions are NOT deleted by
default: a config that omits a resource looks identical to a stale one, so
the sync refuses each deletion, reports it on stderr, and continues with the
creates and updates. Pass --allow-delete with the resource types you intend
to decommission (services, agents, databases, mcps, or all) to delete them; within
each resource type, deletes run after that type's creates and updates.

Updates replace the whole live spec of each resource. For every service, agent,
or mcp updated, the live revision being replaced is printed to stderr so a
sync from a stale config file is visible before it lands.

Use --dry-run to print the full plan — creates, updates, deletes, and
refused deletions — without applying anything.

The organization and project are read from the config file, flags, or resolved via 'iai organizations select' / 'iai projects select'.`,
	Example: `  iai stacks sync --file stack.yaml
  iai stacks sync --file stack.yaml --project my-project --organization my-org
  iai stacks sync --file stack.yaml --dry-run
  iai stacks sync --file stack.yaml --allow-delete services,agents`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		filePath := stackSyncFile
		if filePath == "" {
			filePath = cfgFilePath
		}
		if filePath == "" {
			return fmt.Errorf("config file is required; please provide --file or --cfg-file")
		}

		cfg, err := files.LoadStackConfig(filePath)
		if err != nil {
			return fmt.Errorf("failed to load stack config: %w", err)
		}

		if cfg.StackId == "" {
			return fmt.Errorf("stack-id is required for sync command")
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := api.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := deployment.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, stackSyncOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, stackSyncProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		verb := "Syncing"
		if stackSyncDryRun {
			verb = "Planning"
			fmt.Fprintf(
				out,
				"Dry run: planning stack %q — no changes will be applied.\n",
				cfg.StackId,
			)
		} else {
			fmt.Fprintf(out, "Syncing stack %q...\n", cfg.StackId)
		}
		ranSync := false

		runPhase := func(label string, run func(sync.Options) (*sync.Result, error)) error {
			ranSync = true
			fmt.Fprint(out, verb+" "+label)
			done := output.PrintLoadingDots(out)
			result, err := run(sync.Options{
				AllowDelete: sync.AllowDeleteResource(stackSyncAllowDelete, label),
				DryRun:      stackSyncDryRun,
			})
			close(done)
			fmt.Fprintln(out)
			if stackSyncDryRun {
				if err != nil {
					return err
				}
				sync.PrintPlan(out, label, result)
				return nil
			}
			return sync.PrintResult(out, label, result, err)
		}

		svcBodies := make(map[string]deployment.CreateServiceBody)
		for name, svcCfg := range cfg.Services {
			svcBodies[name] = svcCfg.ToCreateRequest(cfg.StackId)
		}

		hasServices := false
		if len(svcBodies) == 0 {
			hasServices, err = sync.HasServices(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
			)
			if err != nil {
				return err
			}
		}

		if len(svcBodies) > 0 || hasServices {
			err := runPhase("services", func(opts sync.Options) (*sync.Result, error) {
				return sync.Services(
					cmd.Context(),
					cmd.ErrOrStderr(),
					deployClient,
					orgId,
					projectId,
					cfg.StackId,
					svcBodies,
					opts,
				)
			})
			if err != nil {
				return err
			}
		}

		agentBodies := make(map[string]deployment.CreateAgentBody)
		for name, agentCfg := range cfg.Agents {
			agentBodies[name] = agentCfg.ToCreateRequest(cfg.StackId)
		}

		hasAgents := false
		if len(agentBodies) == 0 {
			hasAgents, err = sync.HasAgents(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
			)
			if err != nil {
				return err
			}
		}

		if len(agentBodies) > 0 || hasAgents {
			err := runPhase("agents", func(opts sync.Options) (*sync.Result, error) {
				return sync.Agents(
					cmd.Context(),
					cmd.ErrOrStderr(),
					deployClient,
					orgId,
					projectId,
					cfg.StackId,
					agentBodies,
					opts,
				)
			})
			if err != nil {
				return err
			}
		}

		dbBodies := make(map[string]deployment.CreateDatabaseBody)
		for name, dbCfg := range cfg.Databases {
			dbBodies[name] = dbCfg.ToCreateRequest(cfg.StackId)
		}

		hasDatabases := false
		if len(dbBodies) == 0 {
			hasDatabases, err = sync.HasDatabases(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
			)
			if err != nil {
				return err
			}
		}

		if len(dbBodies) > 0 || hasDatabases {
			err := runPhase("databases", func(opts sync.Options) (*sync.Result, error) {
				return sync.Databases(
					cmd.Context(),
					cmd.ErrOrStderr(),
					deployClient,
					orgId,
					projectId,
					cfg.StackId,
					dbBodies,
					opts,
				)
			})
			if err != nil {
				return err
			}
		}

		mcpBodies := make(map[string]deployment.CreateMcpBody)
		for name, mcpCfg := range cfg.Mcps {
			mcpBodies[name] = mcpCfg.ToCreateRequest(cfg.StackId)
		}

		hasMcps := false
		if len(mcpBodies) == 0 {
			hasMcps, err = sync.HasMcps(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
			)
			if err != nil {
				return err
			}
		}

		if len(mcpBodies) > 0 || hasMcps {
			err := runPhase("mcps", func(opts sync.Options) (*sync.Result, error) {
				return sync.Mcps(
					cmd.Context(),
					cmd.ErrOrStderr(),
					deployClient,
					orgId,
					projectId,
					cfg.StackId,
					mcpBodies,
					opts,
				)
			})
			if err != nil {
				return err
			}
		}

		if !ranSync {
			fmt.Fprintf(out, "No resources to sync for stack %q.\n", cfg.StackId)
		}

		return nil
	},
}

var stackGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Export live stack configuration",
	Long: `Fetch the live services, agents, databases, and mcps for a stack and write
them as a stack configuration file.

Use this to rebase your local stack config on the live state before making
changes. MCP credentials are never exported; include auth.credential before
syncing credentialed MCPs.

The organization and project are read from flags or resolved via 'iai
organizations select' / 'iai projects select'.`,
	Example: `  iai stacks get --stack-id my-stack
  iai stacks get --stack-id my-stack -f live-stack.yaml
  iai stacks get --stack-id my-stack -o my-org -p my-project`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if stackGetStackID == "" {
			return fmt.Errorf("--stack-id is required")
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "Exporting stack %q...\n", stackGetStackID)

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			stackGetOrg,
			stackGetProject,
		)
		if err != nil {
			return err
		}

		liveCfg, err := files.FetchLiveStack(
			cmd.Context(),
			deployClient,
			pCtx.orgId,
			pCtx.projectId,
			stackGetStackID,
		)
		if err != nil {
			return err
		}

		liveCfg.Organization = pCtx.orgName
		liveCfg.Project = pCtx.projectName

		yamlData, err := files.MarshalStackConfig(liveCfg)
		if err != nil {
			return err
		}

		if stackGetJSON {
			return output.PrintStructuredJSON(out, liveCfg)
		}
		if stackGetYAML {
			return output.PrintStructuredYAML(out, liveCfg)
		}

		if stackGetFile != "" {
			if err := os.WriteFile(stackGetFile, yamlData, 0o644); err != nil {
				return err
			}
			fmt.Fprintf(out, "Stack configuration written to %s\n", stackGetFile)
			return nil
		}

		fmt.Fprint(out, string(yamlData))
		return nil
	},
}

var stackDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between local config and live stack",
	Long: `Compare a local stack configuration file against the live state of a
stack and show creates, updates, deletes, and field-level changes.

The local file is read from --file or --cfg-file. The live state is fetched
from the deployment API using --stack-id.

Use --json for machine-readable output in CI pipelines.`,
	Example: `  iai stacks diff --file stack.yaml --stack-id my-stack
  iai stacks diff --file stack.yaml --stack-id my-stack --json
  iai stacks diff --file stack.yaml --stack-id my-stack -o my-org -p my-project`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		filePath := stackDiffFile
		if filePath == "" {
			filePath = cfgFilePath
		}
		if filePath == "" {
			return fmt.Errorf("config file is required; please provide --file or --cfg-file")
		}

		localCfg, err := files.LoadStackConfig(filePath)
		if err != nil {
			return fmt.Errorf("failed to load local config: %w", err)
		}

		if stackDiffOrg == "" {
			stackDiffOrg = localCfg.Organization
		}
		if stackDiffProject == "" {
			stackDiffProject = localCfg.Project
		}

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			stackDiffOrg,
			stackDiffProject,
		)
		if err != nil {
			return err
		}

		stackID := stackDiffStackID
		if stackID == "" {
			stackID = localCfg.StackId
		}
		if stackID == "" {
			return fmt.Errorf("--stack-id is required")
		}

		liveCfg, err := files.FetchLiveStack(
			cmd.Context(),
			deployClient,
			pCtx.orgId,
			pCtx.projectId,
			stackID,
		)
		if err != nil {
			return err
		}

		d := files.DiffStackConfigs(localCfg, liveCfg)

		if stackDiffJSON {
			return output.PrintStructuredJSON(out, d)
		}

		return files.PrintStackDiffDetailed(out, localCfg, liveCfg, d)
	},
}

var stackListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List stacks in a project",
	Long: `List stacks and their resource counts (services, agents, databases, mcps)
in a project. Stacks are discovered from the live resources that belong
to them.

The organization and project are read from flags or resolved via
'iai organizations select' / 'iai projects select'.`,
	Example: `  iai stacks list
  iai stacks list --json
  iai stacks list -o my-org -p my-project`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, _, deployClient, err := resolveProject(
			cmd.Context(),
			stackListOrg,
			stackListProj,
		)
		if err != nil {
			return err
		}

		stacks, err := files.ListStacks(
			cmd.Context(),
			deployClient,
			pCtx.orgId,
			pCtx.projectId,
		)
		if err != nil {
			return err
		}

		if stackListJSON {
			return output.PrintStructuredJSON(out, stacks)
		}

		if len(stacks) == 0 {
			fmt.Fprintln(out, "No stacks found.")
			return nil
		}

		headers := []string{"STACK ID", "SERVICES", "AGENTS", "DATABASES", "MCPS"}
		rows := make([][]string, len(stacks))
		for i, s := range stacks {
			rows[i] = []string{
				s.StackID,
				fmt.Sprintf("%d", s.ServiceCount),
				fmt.Sprintf("%d", s.AgentCount),
				fmt.Sprintf("%d", s.DatabaseCount),
				fmt.Sprintf("%d", s.McpCount),
			}
		}
		return output.PrintTable(out, headers, rows)
	},
}

func init() {
	stackSyncCmd.Flags().
		StringVarP(&stackSyncFile, "file", "f", "", "Path to stack configuration file")
	stackSyncCmd.Flags().
		StringVarP(&stackSyncProject, "project", "p", "", "Project name to sync resources in")
	stackSyncCmd.Flags().
		StringVarP(&stackSyncOrganization, "organization", "o", "", "Organization name that owns the project")
	stackSyncCmd.Flags().
		StringSliceVar(&stackSyncAllowDelete, "allow-delete", nil, "Resource types the sync may delete when the config omits them (services, agents, databases, mcps, or all); deletions are refused otherwise")
	stackSyncCmd.Flags().
		BoolVar(&stackSyncDryRun, "dry-run", false, "Print the full plan (creates, updates, deletes, refused deletions) without applying anything")

	stackGetCmd.Flags().
		StringVar(&stackGetStackID, "stack-id", "", "Stack ID to export")
	stackGetCmd.Flags().
		StringVarP(&stackGetFile, "file", "f", "", "Write output to file instead of stdout")
	stackGetCmd.Flags().
		StringVarP(&stackGetOrg, "organization", "o", "", "Organization name")
	stackGetCmd.Flags().
		StringVarP(&stackGetProject, "project", "p", "", "Project name")
	stackGetCmd.Flags().
		BoolVar(&stackGetJSON, "json", false, "Output as JSON")
	stackGetCmd.Flags().
		BoolVar(&stackGetYAML, "yaml", false, "Output as YAML")
	stackGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	stackDiffCmd.Flags().
		StringVarP(&stackDiffFile, "file", "f", "", "Path to local stack configuration file")
	stackDiffCmd.Flags().
		StringVar(&stackDiffStackID, "stack-id", "", "Stack ID to compare against live")
	stackDiffCmd.Flags().
		StringVarP(&stackDiffOrg, "organization", "o", "", "Organization name")
	stackDiffCmd.Flags().
		StringVarP(&stackDiffProject, "project", "p", "", "Project name")
	stackDiffCmd.Flags().
		BoolVar(&stackDiffJSON, "json", false, "Output diff as JSON")

	stackListCmd.Flags().
		BoolVar(&stackListJSON, "json", false, "Output as JSON")
	stackListCmd.Flags().
		StringVarP(&stackListOrg, "organization", "o", "", "Organization name")
	stackListCmd.Flags().
		StringVarP(&stackListProj, "project", "p", "", "Project name")

	stackCmd.AddCommand(stackSyncCmd)
	stackCmd.AddCommand(stackListCmd)
	stackCmd.AddCommand(stackGetCmd)
	stackCmd.AddCommand(stackDiffCmd)
	rootCmd.AddCommand(stackCmd)
}
