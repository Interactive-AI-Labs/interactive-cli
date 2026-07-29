package cmd

import (
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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
)

var stackCmd = &cobra.Command{
	Use:     "stacks",
	Aliases: []string{"stack", "st"},
	Short:   "Declarative resource sync from config files",
	GroupID: groupInfra,
	Long:    `Manage stacks and their resources (services, agents, databases) from stack configuration files.`,
}

var stackSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync services, agents, and databases from a stack config file",
	Long: `Sync services, agents, and databases in a project from a stack configuration file.

Services, agents, and databases are created and updated to match the config
file. Resources the config file no longer mentions are NOT deleted by
default: a config that omits a resource looks identical to a stale one, so
the sync refuses each deletion, reports it on stderr, and continues with the
creates and updates. Pass --allow-delete with the resource types you intend
to decommission (services, agents, databases, or all) to delete them; within
each resource type, deletes run after that type's creates and updates.

Updates replace the whole live spec of each resource. For every service or
agent updated, the live revision being replaced is printed to stderr so a
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

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
		if err != nil {
			return err
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

		// In a dry run the plan on stdout is the deliverable; otherwise
		// results print as applied, with refused deletions reported.
		printOutcome := func(label string, result *sync.Result, err error) error {
			if stackSyncDryRun {
				if err != nil {
					return err
				}
				sync.PrintPlan(out, label, result)
				return nil
			}
			return sync.PrintResult(out, label, result, err)
		}

		svcBodies := make(map[string]clients.CreateServiceBody)
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
			ranSync = true
			fmt.Fprint(out, verb+" services")
			done := output.PrintLoadingDots(out)

			svcResult, err := sync.Services(
				cmd.Context(),
				cmd.ErrOrStderr(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
				svcBodies,
				sync.Options{
					AllowDelete: sync.AllowDeleteResource(stackSyncAllowDelete, "services"),
					DryRun:      stackSyncDryRun,
				},
			)
			close(done)
			fmt.Fprintln(out)
			if err := printOutcome("services", svcResult, err); err != nil {
				return err
			}
		}

		agentBodies := make(map[string]clients.CreateAgentBody)
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
			ranSync = true
			fmt.Fprint(out, verb+" agents")
			done := output.PrintLoadingDots(out)

			agentResult, err := sync.Agents(
				cmd.Context(),
				cmd.ErrOrStderr(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
				agentBodies,
				sync.Options{
					AllowDelete: sync.AllowDeleteResource(stackSyncAllowDelete, "agents"),
					DryRun:      stackSyncDryRun,
				},
			)
			close(done)
			fmt.Fprintln(out)
			if err := printOutcome("agents", agentResult, err); err != nil {
				return err
			}
		}

		dbBodies := make(map[string]clients.CreateDatabaseBody)
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
			ranSync = true
			fmt.Fprint(out, verb+" databases")
			done := output.PrintLoadingDots(out)

			dbResult, err := sync.Databases(
				cmd.Context(),
				cmd.ErrOrStderr(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
				dbBodies,
				sync.Options{
					AllowDelete: sync.AllowDeleteResource(stackSyncAllowDelete, "databases"),
					DryRun:      stackSyncDryRun,
				},
			)
			close(done)
			fmt.Fprintln(out)
			if err := printOutcome("databases", dbResult, err); err != nil {
				return err
			}
		}

		if !ranSync {
			fmt.Fprintf(out, "No resources to sync for stack %q.\n", cfg.StackId)
		}

		return nil
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
		StringSliceVar(&stackSyncAllowDelete, "allow-delete", nil, "Resource types the sync may delete when the config omits them (services, agents, databases, or all); deletions are refused otherwise")
	stackSyncCmd.Flags().
		BoolVar(&stackSyncDryRun, "dry-run", false, "Print the full plan (creates, updates, deletes, refused deletions) without applying anything")

	stackCmd.AddCommand(stackSyncCmd)
	rootCmd.AddCommand(stackCmd)
}
