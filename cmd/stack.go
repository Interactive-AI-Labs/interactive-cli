package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
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
	stackSyncNoWait       bool
	stackSyncWaitTimeout  time.Duration

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
	Long:    `Manage stacks and their resources (services, agents, databases, mcps, jobs) from stack configuration files.`,
}

var stackSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync services, agents, databases, mcps, and jobs from a stack config file",
	Long: `Sync services, agents, databases, mcps, and jobs in a project from a stack configuration file.

Services, agents, databases, mcps, and jobs are created and updated to match the config
file. Resources the config file no longer mentions are NOT deleted by
default: a config that omits a resource looks identical to a stale one, so
the sync refuses each deletion, reports it on stderr, and continues with the
creates and updates. Pass --allow-delete with the resource types you intend
to decommission (services, agents, databases, mcps, jobs, or all) to delete them;
within each resource type, deletes run after that type's creates and updates.

Resource types sync in order: services, databases, mcps, then agents once
the self-hosted mcps are ready, and finally jobs. Self-hosted mcps run on the
deployment operator; remote ones are registered on the platform.

Updates replace the whole live spec of each resource. For every service, agent,
mcp, or job updated, the live revision being replaced is printed to stderr so a
sync from a stale config file is visible before it lands. Jobs can only be
updated or deleted when all their runs have finished.

Script jobs reference their files with scriptFile and pyprojectFile, resolved
relative to the config file.

Use --dry-run to print the full plan — creates, updates, deletes, and
refused deletions — without applying anything.

The organization and project are read from the config file, flags, or resolved via 'iai organizations select' / 'iai projects select'.`,
	Example: `  iai stacks sync --file stack.yaml
  iai stacks sync --file stack.yaml --project my-project --organization my-org
  iai stacks sync --file stack.yaml --dry-run
  iai stacks sync --file stack.yaml --wait-timeout 10m
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

		if stackSyncWaitTimeout <= 0 {
			return fmt.Errorf("--wait-timeout must be positive; pass --no-wait to skip the wait")
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := platform.NewAPIClient(
			hostname,
			defaultHTTPTimeout,
			token,
			apiKey,
			cookies,
		)
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

		runPhase := func(
			label string,
			run func(sync.Options) (*sync.Result, error),
		) (*sync.Result, error) {
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
					return nil, err
				}
				sync.PrintPlan(out, label, result)
				return result, nil
			}
			return result, sync.PrintResult(out, label, result, err)
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
			_, err := runPhase("services", func(opts sync.Options) (*sync.Result, error) {
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
			_, err := runPhase("databases", func(opts sync.Options) (*sync.Result, error) {
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
		remoteMcps := make(map[string]platform.McpCreateRequest)
		for name, mcpCfg := range cfg.Mcps {
			if mcpCfg.Type == deployment.McpTypeRemote {
				remoteMcps[name] = mcpCfg.ToPlatformCreateRequest(name, cfg.StackId)
				continue
			}
			mcpBodies[name] = mcpCfg.ToCreateRequest(cfg.StackId)
		}

		hasMcps := false
		if len(mcpBodies) == 0 && len(remoteMcps) == 0 {
			hasMcps, err = sync.HasMcps(
				cmd.Context(),
				deployClient,
				apiClient,
				orgId,
				projectId,
				cfg.StackId,
			)
			if err != nil {
				return err
			}
		}

		var mcpResult *sync.Result
		if len(mcpBodies) > 0 || len(remoteMcps) > 0 || hasMcps {
			mcpResult, err = runPhase("mcps", func(opts sync.Options) (*sync.Result, error) {
				return sync.Mcps(
					cmd.Context(),
					cmd.ErrOrStderr(),
					deployClient,
					apiClient,
					orgId,
					projectId,
					cfg.StackId,
					mcpBodies,
					remoteMcps,
					opts,
				)
			})
			if err != nil {
				return err
			}
		}

		// Only self-hosted mcps roll out; remote ones are registered on the platform with nothing to wait for.
		var changedMcps []string
		if mcpResult != nil {
			for _, name := range slices.Concat(mcpResult.Created, mcpResult.Updated) {
				if _, selfHosted := mcpBodies[name]; selfHosted {
					changedMcps = append(changedMcps, name)
				}
			}
		}
		if len(changedMcps) > 0 && !stackSyncDryRun && !stackSyncNoWait {
			fmt.Fprint(out, "Waiting for mcps to be ready")
			done := output.PrintLoadingDots(out)
			err := sync.WaitForMcps(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				changedMcps,
				stackSyncWaitTimeout,
			)
			close(done)
			fmt.Fprintln(out)
			if err != nil {
				return fmt.Errorf(
					"%w; agents were not synced — check with 'iai mcps describe <mcp_name>'",
					err,
				)
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
			_, err := runPhase("agents", func(opts sync.Options) (*sync.Result, error) {
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

		jobBodies := make(map[string]deployment.CreateJobBody)
		for name, jobCfg := range cfg.Jobs {
			jobBodies[name] = jobCfg.ToCreateRequest(cfg.StackId)
		}

		hasJobs := false
		if len(jobBodies) == 0 {
			hasJobs, err = sync.HasJobs(
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

		if len(jobBodies) > 0 || hasJobs {
			_, err := runPhase("jobs", func(opts sync.Options) (*sync.Result, error) {
				return sync.Jobs(
					cmd.Context(),
					cmd.ErrOrStderr(),
					deployClient,
					orgId,
					projectId,
					cfg.StackId,
					jobBodies,
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
	Long: `Fetch the live services, agents, databases, mcps, and jobs for a stack and
write them as a stack configuration file.

Use this to rebase your local stack config on the live state before making
changes. MCP credentials are never exported; include auth.credential before
syncing credentialed MCPs.

With --file, each script job's files are written to jobs/<name>/main.py and
jobs/<name>/pyproject.toml next to the config file, overwriting existing
files. Other outputs omit script job files; add scriptFile and pyprojectFile
before syncing.

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

		pCtx, apiClient, deployClient, err := resolveProject(
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
			apiClient,
			pCtx.orgId,
			pCtx.projectId,
			stackGetStackID,
		)
		if err != nil {
			return err
		}

		liveCfg.Organization = pCtx.orgName
		liveCfg.Project = pCtx.projectName

		if stackGetFile == "" {
			if names := files.ScriptJobNames(liveCfg); len(names) > 0 {
				fmt.Fprintf(
					cmd.ErrOrStderr(),
					"Warning: files of script jobs %s are not exported; use --file to write them.\n",
					strings.Join(names, ", "),
				)
			}
			if stackGetJSON {
				return output.PrintStructuredJSON(out, liveCfg)
			}
			if stackGetYAML {
				return output.PrintStructuredYAML(out, liveCfg)
			}
			yamlData, err := files.MarshalStackConfig(liveCfg)
			if err != nil {
				return err
			}
			fmt.Fprint(out, string(yamlData))
			return nil
		}

		if err := files.WriteJobFiles(liveCfg, filepath.Dir(stackGetFile)); err != nil {
			return err
		}
		yamlData, err := files.MarshalStackConfig(liveCfg)
		if err != nil {
			return err
		}
		if err := os.WriteFile(stackGetFile, yamlData, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(out, "Stack configuration written to %s\n", stackGetFile)
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

		pCtx, apiClient, deployClient, err := resolveProject(
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
			apiClient,
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
	Long: `List stacks and their resource counts (services, agents, databases, mcps,
jobs) in a project. Stacks are discovered from the live resources that belong
to them.

The organization and project are read from flags or resolved via
'iai organizations select' / 'iai projects select'.`,
	Example: `  iai stacks list
  iai stacks list --json
  iai stacks list -o my-org -p my-project`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, apiClient, deployClient, err := resolveProject(
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
			apiClient,
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

		headers := []string{"STACK ID", "SERVICES", "AGENTS", "DATABASES", "MCPS", "JOBS"}
		rows := make([][]string, len(stacks))
		for i, s := range stacks {
			rows[i] = []string{
				s.StackID,
				fmt.Sprintf("%d", s.ServiceCount),
				fmt.Sprintf("%d", s.AgentCount),
				fmt.Sprintf("%d", s.DatabaseCount),
				fmt.Sprintf("%d", s.McpCount),
				fmt.Sprintf("%d", s.JobCount),
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
		StringSliceVar(&stackSyncAllowDelete, "allow-delete", nil, "Resource types the sync may delete when the config omits them (services, agents, databases, mcps, jobs, or all); deletions are refused otherwise")
	stackSyncCmd.Flags().
		BoolVar(&stackSyncDryRun, "dry-run", false, "Print the full plan (creates, updates, deletes, refused deletions) without applying anything")
	stackSyncCmd.Flags().
		BoolVar(&stackSyncNoWait, "no-wait", false, "Sync agents without waiting for the self-hosted mcps to be ready")
	stackSyncCmd.Flags().
		DurationVar(&stackSyncWaitTimeout, "wait-timeout", 5*time.Minute, "How long to wait for every self-hosted mcp in the config to be ready (tools verified) before syncing agents; if one is not ready in time the sync fails and agents are left unchanged")
	stackSyncCmd.MarkFlagsMutuallyExclusive("no-wait", "wait-timeout")

	stackGetCmd.Flags().
		StringVar(&stackGetStackID, "stack-id", "", "Stack ID to export")
	stackGetCmd.Flags().
		StringVarP(&stackGetFile, "file", "f", "", "Write output to file instead of stdout; cannot combine with --json or --yaml")
	stackGetCmd.Flags().
		StringVarP(&stackGetOrg, "organization", "o", "", "Organization name")
	stackGetCmd.Flags().
		StringVarP(&stackGetProject, "project", "p", "", "Project name")
	stackGetCmd.Flags().
		BoolVar(&stackGetJSON, "json", false, "Output as JSON")
	stackGetCmd.Flags().
		BoolVar(&stackGetYAML, "yaml", false, "Output as YAML")
	stackGetCmd.MarkFlagsMutuallyExclusive("file", "json", "yaml")

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
