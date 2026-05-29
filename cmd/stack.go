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

Services are created, updated, or deleted to match the config file.
Agents are created, updated, or deleted to match the config file.
Databases are created, updated, or deleted (--allow-delete=databases) to match the config file.

The organization and project are read from the config file, flags, or resolved via 'iai organizations select' / 'iai projects select'.`,
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
		fmt.Fprintf(out, "Syncing stack %q...\n", cfg.StackId)
		ranSync := false

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
			fmt.Fprint(out, "Syncing services")
			done := output.PrintLoadingDots(out)

			svcResult, err := sync.Services(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
				svcBodies,
			)
			close(done)
			fmt.Fprintln(out)
			if err := sync.PrintResult(out, "services", svcResult, err); err != nil {
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
			fmt.Fprint(out, "Syncing agents")
			done := output.PrintLoadingDots(out)

			agentResult, err := sync.Agents(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
				agentBodies,
			)
			close(done)
			fmt.Fprintln(out)
			if err := sync.PrintResult(out, "agents", agentResult, err); err != nil {
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
			fmt.Fprint(out, "Syncing databases")
			done := output.PrintLoadingDots(out)

			allowDeleteDB := sync.AllowDeleteResource(
				stackSyncAllowDelete,
				"databases",
			)
			dbResult, err := sync.Databases(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
				dbBodies,
				allowDeleteDB,
			)
			close(done)
			fmt.Fprintln(out)
			if err := sync.PrintResult(out, "databases", dbResult, err); err != nil {
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
		StringSliceVar(&stackSyncAllowDelete, "allow-delete", nil, "Resource types to allow deletion for (e.g. databases)")

	stackCmd.AddCommand(stackSyncCmd)
	rootCmd.AddCommand(stackCmd)
}
