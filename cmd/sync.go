package cmd

import (
	"context"
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
	"github.com/spf13/cobra"
)

var (
	syncProject      string
	syncOrganization string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync services and vector stores from a stack config file",
	Long: `Sync services and vector stores in a project from a stack configuration file.

For services, sync will:
- Create services that exist in the config but not in the project
- Update services that exist in both the config and the project
- Delete services that exist in the project but not in the config (for the specified stack)

For vector stores, sync will:
- Create vector stores that exist in the config but not in the project
- Delete vector stores that exist in the project but not in the config (for the specified stack)

The project is selected with --project or via 'iai projects select', and the config file with --cfg-file.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if cfgFilePath == "" {
			return fmt.Errorf("config file is required; please provide --cfg-file")
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
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

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := clients.NewDeploymentClient(
			deploymentHostname,
			defaultHTTPTimeout,
			apiKey,
			cookies,
		)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, syncOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, syncProject)
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

			svcResult, err := syncServices(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg,
			)
			close(done)
			fmt.Fprintln(out)
			if err != nil {
				return err
			}

			output.PrintSyncResult(out, "services", svcResult.Created, svcResult.Updated, svcResult.Deleted)
		}

		if len(cfg.VectorStores) > 0 {
			fmt.Fprint(out, "Syncing vector stores")
			done := output.PrintLoadingDots(out)

			vsResult, err := syncVectorStores(
				cmd.Context(),
				deployClient,
				orgId,
				projectId,
				cfg,
			)
			close(done)
			fmt.Fprintln(out)
			if err != nil {
				return err
			}

			output.PrintSyncResult(out, "vector stores", vsResult.Created, nil, vsResult.Deleted)
		}

		return nil
	},
}

// syncResult holds the outcome of a sync operation.
type syncResult struct {
	Created []string
	Updated []string
	Deleted []string
}

// syncServices creates, updates, and deletes services to match the stack config.
// It is scoped to the stack identified by cfg.StackId.
func syncServices(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId string,
	cfg *files.StackConfig,
) (*syncResult, error) {
	existing, err := deployClient.ListServices(ctx, orgId, projectId, cfg.StackId)
	if err != nil {
		return nil, err
	}

	existingByName := make(map[string]clients.ServiceOutput)
	for _, svc := range existing {
		existingByName[svc.Name] = svc
	}

	result := &syncResult{
		Created: []string{},
		Updated: []string{},
		Deleted: []string{},
	}

	for name, svcCfg := range cfg.Services {
		req := svcCfg.ToCreateRequest(cfg.StackId)

		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateService(ctx, orgId, projectId, name, req)
			if err != nil {
				return nil, err
			}
			result.Created = append(result.Created, name)
		} else {
			_, err := deployClient.UpdateService(ctx, orgId, projectId, name, req)
			if err != nil {
				return nil, err
			}
			result.Updated = append(result.Updated, name)
		}
	}

	for name := range existingByName {
		if _, desired := cfg.Services[name]; !desired {
			_, err := deployClient.DeleteService(ctx, orgId, projectId, name)
			if err != nil {
				return nil, err
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}

// syncVectorStores creates and deletes vector stores to match the stack config.
// Updates are not supported (no update endpoint). Scoped to cfg.StackId.
func syncVectorStores(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId string,
	cfg *files.StackConfig,
) (*syncResult, error) {
	existing, err := deployClient.ListVectorStores(ctx, orgId, projectId, cfg.StackId)
	if err != nil {
		return nil, err
	}

	existingByName := make(map[string]clients.VectorStoreInfo)
	for _, vs := range existing {
		existingByName[vs.VectorStoreName] = vs
	}

	result := &syncResult{
		Created: []string{},
		Deleted: []string{},
	}

	for name, vsCfg := range cfg.VectorStores {
		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateVectorStore(ctx, orgId, projectId, name, vsCfg.ToCreateRequest(cfg.StackId))
			if err != nil {
				return nil, err
			}
			result.Created = append(result.Created, name)
		}
		// no update: vector stores have no update endpoint yet
	}

	for name := range existingByName {
		if _, desired := cfg.VectorStores[name]; !desired {
			_, err := deployClient.DeleteVectorStore(ctx, orgId, projectId, name)
			if err != nil {
				return nil, err
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}

func init() {
	syncCmd.Flags().StringVarP(&syncProject, "project", "p", "", "Project name to sync in")
	syncCmd.Flags().
		StringVarP(&syncOrganization, "organization", "o", "", "Organization name that owns the project")

	rootCmd.AddCommand(syncCmd)
}
