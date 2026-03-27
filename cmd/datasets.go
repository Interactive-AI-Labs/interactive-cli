package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	datasetsListPage    int
	datasetsListLimit   int
	datasetsListColumns []string
	datasetsListJSON    bool
	datasetsListOrg     string
	datasetsListProject string

	datasetsGetJSON    bool
	datasetsGetOrg     string
	datasetsGetProject string

	datasetsCreateDescription  string
	datasetsCreateMetadataJSON string
	datasetsCreateOrg          string
	datasetsCreateProject      string
	datasetsCreateJSON         bool
)

var datasetsCmd = &cobra.Command{
	Use:              "datasets",
	Aliases:          []string{"dataset"},
	Short:            "Manage evaluation datasets",
	Long:             `Manage evaluation datasets. Works with API key or session login.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var datasetsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List datasets",
	Long:    `List evaluation datasets with pagination.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := datasetsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultDatasetColumns
		}
		if !datasetsListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllDatasetColumns); err != nil {
				return err
			}
		}

		opts := clients.DatasetListOptions{
			Page:  datasetsListPage,
			Limit: datasetsListLimit,
		}
		if err := inputs.ValidateDatasetListOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), datasetsListOrg, datasetsListProject)
		if err != nil {
			return err
		}

		datasets, meta, rawJSON, err := pCtx.apiClient.ListDatasets(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if datasetsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintDatasetList(out, datasets, meta, columns)
	},
}

var datasetsGetCmd = &cobra.Command{
	Use:     "get <name>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a dataset by name",
	Long:    `Get detailed information about a specific dataset.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		name := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(cmd.Context(), datasetsGetOrg, datasetsGetProject)
		if err != nil {
			return err
		}

		dataset, rawJSON, err := pCtx.apiClient.GetDataset(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			name,
		)
		if err != nil {
			return err
		}

		if datasetsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintDatasetDetail(out, dataset)
	},
}

var datasetsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a dataset",
	Long:  `Create a new evaluation dataset.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		body, err := inputs.BuildDatasetCreateBody(
			args[0],
			datasetsCreateDescription,
			datasetsCreateMetadataJSON,
		)
		if err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), datasetsCreateOrg, datasetsCreateProject)
		if err != nil {
			return err
		}

		dataset, rawJSON, err := pCtx.apiClient.CreateDataset(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			body,
		)
		if err != nil {
			return err
		}

		if datasetsCreateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintDatasetCreateResult(out, dataset)
	},
}

func init() {
	datasetsListCmd.Flags().IntVar(&datasetsListPage, "page", 1, "Page number (starts at 1)")
	datasetsListCmd.Flags().IntVar(&datasetsListLimit, "limit", 50, "Items per page")
	datasetsListCmd.Flags().
		StringSliceVar(&datasetsListColumns, "columns", nil, "Columns to display (comma-separated)")
	datasetsListCmd.Flags().
		BoolVar(&datasetsListJSON, "json", false, "Output raw API response as JSON")
	datasetsListCmd.Flags().
		StringVarP(&datasetsListOrg, "organization", "o", "", "Organization name that owns the project")
	datasetsListCmd.Flags().
		StringVarP(&datasetsListProject, "project", "p", "", "Project name")

	datasetsGetCmd.Flags().
		BoolVar(&datasetsGetJSON, "json", false, "Output raw API response as JSON")
	datasetsGetCmd.Flags().
		StringVarP(&datasetsGetOrg, "organization", "o", "", "Organization name that owns the project")
	datasetsGetCmd.Flags().
		StringVarP(&datasetsGetProject, "project", "p", "", "Project name")

	datasetsCreateCmd.Flags().
		StringVar(&datasetsCreateDescription, "description", "", "Dataset description")
	datasetsCreateCmd.Flags().
		StringVar(&datasetsCreateMetadataJSON, "metadata-json", "", "Metadata as JSON object")
	datasetsCreateCmd.Flags().
		BoolVar(&datasetsCreateJSON, "json", false, "Output raw API response as JSON")
	datasetsCreateCmd.Flags().
		StringVarP(&datasetsCreateOrg, "organization", "o", "", "Organization name that owns the project")
	datasetsCreateCmd.Flags().
		StringVarP(&datasetsCreateProject, "project", "p", "", "Project name")

	datasetsCmd.AddCommand(datasetsListCmd, datasetsGetCmd, datasetsCreateCmd)
	rootCmd.AddCommand(datasetsCmd)
}
