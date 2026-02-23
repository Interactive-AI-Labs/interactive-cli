package cmd

import (
	"fmt"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
	"github.com/spf13/cobra"
)

var (
	vsProject      string
	vsOrganization string
	vsCPU          int
	vsMemory       float64
	vsStorageSize  int
	vsAutoResize   bool
	vsAutoResLimit int
	vsHA           bool
	vsBackups      bool
)

var vectorStoresCmd = &cobra.Command{
	Use:     "vector-stores",
	Aliases: []string{"vector-store", "vs"},
	Short:   "Manage vector stores",
	Long:    `Manage vector stores in InteractiveAI projects.`,
}

var vsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List vector stores in a project",
	Long: `List vector stores in a specific project.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, vsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, vsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		stores, err := deployClient.ListVectorStores(cmd.Context(), orgId, projectId)
		if err != nil {
			return err
		}

		return output.PrintVectorStoreList(out, stores)
	},
}

var vsGetCmd = &cobra.Command{
	Use:   "get <vectorStoreName>",
	Short: "Get status of a vector store",
	Long: `Get the status of a specific vector store.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		vectorStoreName := strings.TrimSpace(args[0])

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, vsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, vsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		store, err := deployClient.GetVectorStore(cmd.Context(), orgId, projectId, vectorStoreName)
		if err != nil {
			return err
		}

		return output.PrintVectorStoreList(out, []clients.VectorStoreInfo{*store})
	},
}

var vsCreateCmd = &cobra.Command{
	Use:   "create <vectorStoreName>",
	Short: "Create a vector store",
	Long: `Create a vector store in a specific project.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		vectorStoreName := strings.TrimSpace(args[0])

		if vsAutoResize && !cmd.Flags().Changed("auto-resize-limit") {
			return fmt.Errorf("--auto-resize-limit is required when --auto-resize is enabled")
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, vsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, vsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		reqBody := clients.CreateVectorStoreBody{
			Resources: clients.VectorStoreResources{
				CPU:    vsCPU,
				Memory: vsMemory,
			},
			Storage: clients.VectorStoreStorage{
				Size:            vsStorageSize,
				AutoResize:      vsAutoResize,
				AutoResizeLimit: vsAutoResLimit,
			},
			HA:      vsHA,
			Backups: vsBackups,
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting vector store creation request...")

		serverMessage, err := deployClient.CreateVectorStore(cmd.Context(), orgId, projectId, vectorStoreName, reqBody)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var vsDeleteCmd = &cobra.Command{
	Use:     "delete <vectorStoreName>",
	Aliases: []string{"rm"},
	Short:   "Delete a vector store",
	Long: `Delete a vector store in a specific project.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		vectorStoreName := strings.TrimSpace(args[0])

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, vsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, vsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting vector store deletion request...")

		serverMessage, err := deployClient.DeleteVectorStore(cmd.Context(), orgId, projectId, vectorStoreName)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

func init() {
	// vector-stores list
	vsListCmd.Flags().StringVarP(&vsProject, "project", "p", "", "Project name that owns the vector stores")
	vsListCmd.Flags().StringVarP(&vsOrganization, "organization", "o", "", "Organization name that owns the project")

	// vector-stores get
	vsGetCmd.Flags().StringVarP(&vsProject, "project", "p", "", "Project name that owns the vector stores")
	vsGetCmd.Flags().StringVarP(&vsOrganization, "organization", "o", "", "Organization name that owns the project")

	// vector-stores create
	vsCreateCmd.Flags().StringVarP(&vsProject, "project", "p", "", "Project name that owns the vector stores")
	vsCreateCmd.Flags().StringVarP(&vsOrganization, "organization", "o", "", "Organization name that owns the project")
	vsCreateCmd.Flags().IntVar(&vsCPU, "cpu", 0, "CPU cores (2-80, must be even)")
	vsCreateCmd.Flags().Float64Var(&vsMemory, "memory", 0, "Memory in GB, numeric value only (0.25 increments)")
	vsCreateCmd.Flags().IntVar(&vsStorageSize, "storage-size", 0, "Storage size in GB, numeric value only (min 20)")
	vsCreateCmd.Flags().BoolVar(&vsAutoResize, "auto-resize", false, "Enable automatic storage resizing")
	vsCreateCmd.Flags().IntVar(&vsAutoResLimit, "auto-resize-limit", 0, "Auto-resize limit in GB (0 = unlimited, requires --auto-resize)")
	vsCreateCmd.Flags().BoolVar(&vsHA, "ha", false, "Enable high availability with a standby replica in a separate zone for automatic failover")
	vsCreateCmd.Flags().BoolVar(&vsBackups, "backups", false, "Enable automated daily backups with point-in-time recovery")
	vsCreateCmd.MarkFlagRequired("cpu")
	vsCreateCmd.MarkFlagRequired("memory")
	vsCreateCmd.MarkFlagRequired("storage-size")

	// vector-stores delete
	vsDeleteCmd.Flags().StringVarP(&vsProject, "project", "p", "", "Project name that owns the vector stores")
	vsDeleteCmd.Flags().StringVarP(&vsOrganization, "organization", "o", "", "Organization name that owns the project")

	// Wire up the command hierarchy
	vectorStoresCmd.AddCommand(vsListCmd, vsGetCmd, vsCreateCmd, vsDeleteCmd)
	rootCmd.AddCommand(vectorStoresCmd)
}
