package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
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
	Use:        "vector-stores",
	Aliases:    []string{"vector-store", "vs"},
	Short:      "Manage vector stores",
	GroupID:    groupInfra,
	Long:       `Manage vector stores in InteractiveAI projects.`,
	Deprecated: "use 'iai databases' with the vector extension instead",
	Hidden:     true,
}

var vsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List vector stores in a project",
	Long:    `List vector stores in a specific project.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), vsOrganization, vsProject)
		if err != nil {
			return err
		}

		stores, err := deployClient.ListVectorStores(cmd.Context(), pCtx.orgId, pCtx.projectId, "")
		if err != nil {
			return err
		}

		return output.PrintVectorStoreList(out, stores)
	},
}

var vsDescribeCmd = &cobra.Command{
	Use:     "describe <vectorStoreName>",
	Aliases: []string{"desc"},
	Short:   "Describe a vector store in detail",
	Long:    `Show detailed information about a specific vector store including status, resources, storage, HA, and backup settings.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		vectorStoreName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), vsOrganization, vsProject)
		if err != nil {
			return err
		}

		store, err := deployClient.DescribeVectorStore(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			vectorStoreName,
		)
		if err != nil {
			return err
		}

		return output.PrintVectorStoreDescribe(out, store)
	},
}

var vsCreateCmd = &cobra.Command{
	Use:   "create <vectorStoreName>",
	Short: "Create a vector store",
	Long:  `Create a vector store in a specific project.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		vectorStoreName := strings.TrimSpace(args[0])

		if vsAutoResize && !cmd.Flags().Changed("auto-resize-limit") {
			return fmt.Errorf("--auto-resize-limit is required when --auto-resize is enabled")
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), vsOrganization, vsProject)
		if err != nil {
			return err
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

		serverMessage, err := deployClient.CreateVectorStore(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			vectorStoreName,
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

var vsDeleteCmd = &cobra.Command{
	Use:     "delete <vectorStoreName>",
	Aliases: []string{"rm"},
	Short:   "Delete a vector store",
	Long:    `Delete a vector store in a specific project.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		vectorStoreName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), vsOrganization, vsProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting vector store deletion request...")

		serverMessage, err := deployClient.DeleteVectorStore(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			vectorStoreName,
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

func init() {
	// vector-stores list
	vsListCmd.Flags().
		StringVarP(&vsProject, "project", "p", "", "Project name that owns the vector stores")
	vsListCmd.Flags().
		StringVarP(&vsOrganization, "organization", "o", "", "Organization name that owns the project")

	// vector-stores describe
	vsDescribeCmd.Flags().
		StringVarP(&vsProject, "project", "p", "", "Project name that owns the vector stores")
	vsDescribeCmd.Flags().
		StringVarP(&vsOrganization, "organization", "o", "", "Organization name that owns the project")

	// vector-stores create
	vsCreateCmd.Flags().
		StringVarP(&vsProject, "project", "p", "", "Project name that owns the vector stores")
	vsCreateCmd.Flags().
		StringVarP(&vsOrganization, "organization", "o", "", "Organization name that owns the project")
	vsCreateCmd.Flags().IntVar(&vsCPU, "cpu", 0, "CPU cores (2-80, must be even)")
	vsCreateCmd.Flags().
		Float64Var(&vsMemory, "memory", 0, "Memory in GB (2-8 per vCPU, 0.25 increments)")
	vsCreateCmd.Flags().
		IntVar(&vsStorageSize, "storage-size", 0, "Storage size in GB, numeric value only (min 20)")
	vsCreateCmd.Flags().
		BoolVar(&vsAutoResize, "auto-resize", false, "Enable automatic storage resizing")
	vsCreateCmd.Flags().
		IntVar(&vsAutoResLimit, "auto-resize-limit", 0, "Auto-resize limit in GB (0 = unlimited, requires --auto-resize)")
	vsCreateCmd.Flags().
		BoolVar(&vsHA, "ha", false, "Enable high availability with a standby replica in a separate zone for automatic failover")
	vsCreateCmd.Flags().
		BoolVar(&vsBackups, "backups", false, "Enable automated daily backups with point-in-time recovery")
	vsCreateCmd.MarkFlagRequired("cpu")
	vsCreateCmd.MarkFlagRequired("memory")
	vsCreateCmd.MarkFlagRequired("storage-size")

	// vector-stores delete
	vsDeleteCmd.Flags().
		StringVarP(&vsProject, "project", "p", "", "Project name that owns the vector stores")
	vsDeleteCmd.Flags().
		StringVarP(&vsOrganization, "organization", "o", "", "Organization name that owns the project")

	// Wire up the command hierarchy
	vectorStoresCmd.AddCommand(vsListCmd, vsDescribeCmd, vsCreateCmd, vsDeleteCmd)
	rootCmd.AddCommand(vectorStoresCmd)
}
