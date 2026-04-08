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
	Use:     "stack",
	Aliases: []string{"stacks"},
	Short:   "Manage stacks",
	Long:    `Manage stacks and their resources (services, vector stores) from stack configuration files.`,
}

var stackSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync services and vector stores from a stack config file",
	Long: `Sync services and vector stores in a project from a stack configuration file.

For services, sync will:
- Create services that exist in the config but not in the project
- Update services that exist in both the config and the project
- Delete services that exist in the project but not in the config (for the specified stack)

For vector stores, sync will:
- Create vector stores that exist in the config but not in the project
- Delete vector stores that exist in the project but not in the config (requires --allow-delete=vector-stores)

Stateful resources (vector stores) are protected from deletion by default.
Use --allow-delete to opt in per resource type.

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

		if len(cfg.Services) == 0 && len(cfg.VectorStores) == 0 {
			return fmt.Errorf("config file must define at least one service or vector store")
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

		if len(cfg.Services) > 0 {
			fmt.Fprint(out, "Syncing services")
			done := output.PrintLoadingDots(out)

			svcBodies := make(map[string]clients.CreateServiceBody)
			for name, svcCfg := range cfg.Services {
				svcBodies[name] = svcCfg.ToCreateRequest(cfg.StackId)
			}

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

		if len(cfg.VectorStores) > 0 {
			fmt.Fprint(out, "Syncing vector stores")
			done := output.PrintLoadingDots(out)

			vsBodies := make(map[string]clients.CreateVectorStoreBody)
			for name, vsCfg := range cfg.VectorStores {
				vsBodies[name] = vsCfg.ToCreateRequest(cfg.StackId)
			}

			allowDeleteVS := sync.AllowDeleteResource(
				stackSyncAllowDelete,
				"vector-stores",
			)
			vsResult, err := sync.VectorStores(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg.StackId,
				vsBodies,
				allowDeleteVS,
			)
			close(done)
			fmt.Fprintln(out)
			if err := sync.PrintResult(out, "vector stores", vsResult, err); err != nil {
				return err
			}
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
		StringSliceVar(&stackSyncAllowDelete, "allow-delete", nil, "Resource types to allow deletion for (e.g. vector-stores)")

	stackCmd.AddCommand(stackSyncCmd)
	rootCmd.AddCommand(stackCmd)
}
