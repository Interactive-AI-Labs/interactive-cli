package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	datasetItemsListDatasetName         string
	datasetItemsListSourceTraceID       string
	datasetItemsListSourceObservationID string
	datasetItemsListPage                int
	datasetItemsListLimit               int
	datasetItemsListColumns             []string
	datasetItemsListJSON                bool
	datasetItemsListOrg                 string
	datasetItemsListProject             string

	datasetItemsGetJSON    bool
	datasetItemsGetOrg     string
	datasetItemsGetProject string

	datasetItemsCreateID                  string
	datasetItemsCreateDatasetName         string
	datasetItemsCreateInput               string
	datasetItemsCreateExpectedOutput      string
	datasetItemsCreateMetadataJSON        string
	datasetItemsCreateSourceTraceID       string
	datasetItemsCreateSourceObservationID string
	datasetItemsCreateStatus              string
	datasetItemsCreateJSON                bool
	datasetItemsCreateOrg                 string
	datasetItemsCreateProject             string

	datasetItemsDeleteOrg     string
	datasetItemsDeleteProject string
)

var datasetItemsCmd = &cobra.Command{
	Use:              "dataset-items",
	Aliases:          []string{"dataset-item"},
	Short:            "Manage dataset items",
	Long:             `Manage individual items within evaluation datasets.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var datasetItemsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List dataset items",
	Long:    `List items in a dataset with optional filters.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := datasetItemsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultDatasetItemColumns
		}
		if !datasetItemsListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllDatasetItemColumns); err != nil {
				return err
			}
		}

		opts := clients.DatasetItemListOptions{
			DatasetName:         datasetItemsListDatasetName,
			SourceTraceID:       datasetItemsListSourceTraceID,
			SourceObservationID: datasetItemsListSourceObservationID,
			Page:                datasetItemsListPage,
			Limit:               datasetItemsListLimit,
		}
		if err := inputs.ValidateDatasetItemListOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(
			cmd.Context(),
			datasetItemsListOrg,
			datasetItemsListProject,
		)
		if err != nil {
			return err
		}

		items, meta, rawJSON, err := pCtx.apiClient.ListDatasetItems(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if datasetItemsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintDatasetItemList(out, items, meta, columns)
	},
}

var datasetItemsGetCmd = &cobra.Command{
	Use:     "get <id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a dataset item",
	Long:    `Get detailed information about a specific dataset item.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		itemID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(cmd.Context(), datasetItemsGetOrg, datasetItemsGetProject)
		if err != nil {
			return err
		}

		item, rawJSON, err := pCtx.apiClient.GetDatasetItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			itemID,
		)
		if err != nil {
			return err
		}

		if datasetItemsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintDatasetItemDetail(out, item)
	},
}

var datasetItemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a dataset item",
	Long:  `Create or upsert an item in a dataset.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		body, err := inputs.BuildDatasetItemCreateBody(inputs.DatasetItemCreateInput{
			ID:                  datasetItemsCreateID,
			DatasetName:         datasetItemsCreateDatasetName,
			InputJSON:           datasetItemsCreateInput,
			ExpectedOutputJSON:  datasetItemsCreateExpectedOutput,
			MetadataJSON:        datasetItemsCreateMetadataJSON,
			SourceTraceID:       datasetItemsCreateSourceTraceID,
			SourceObservationID: datasetItemsCreateSourceObservationID,
			Status:              datasetItemsCreateStatus,
		})
		if err != nil {
			return err
		}

		pCtx, err := resolveProject(
			cmd.Context(),
			datasetItemsCreateOrg,
			datasetItemsCreateProject,
		)
		if err != nil {
			return err
		}

		item, rawJSON, err := pCtx.apiClient.CreateDatasetItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			body,
		)
		if err != nil {
			return err
		}

		if datasetItemsCreateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintDatasetItemCreateResult(out, item)
	},
}

var datasetItemsDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm"},
	Short:   "Delete a dataset item",
	Long:    `Delete a dataset item by ID.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		itemID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(
			cmd.Context(),
			datasetItemsDeleteOrg,
			datasetItemsDeleteProject,
		)
		if err != nil {
			return err
		}

		message, err := pCtx.apiClient.DeleteDatasetItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			itemID,
		)
		if err != nil {
			return err
		}

		return output.PrintDeleteSuccess(out, itemID, "dataset item", message)
	},
}

func init() {
	datasetItemsListCmd.Flags().
		StringVar(&datasetItemsListDatasetName, "dataset-name", "", "Dataset name (required)")
	datasetItemsListCmd.Flags().
		StringVar(&datasetItemsListSourceTraceID, "source-trace-id", "", "Filter by source trace ID")
	datasetItemsListCmd.Flags().
		StringVar(
			&datasetItemsListSourceObservationID, "source-observation-id", "",
			"Filter by source observation ID",
		)
	datasetItemsListCmd.Flags().
		IntVar(&datasetItemsListPage, "page", 1, "Page number (starts at 1)")
	datasetItemsListCmd.Flags().IntVar(&datasetItemsListLimit, "limit", 50, "Items per page")
	datasetItemsListCmd.Flags().
		StringSliceVar(&datasetItemsListColumns, "columns", nil, "Columns to display (comma-separated)")
	datasetItemsListCmd.Flags().
		BoolVar(&datasetItemsListJSON, "json", false, "Output raw API response as JSON")
	datasetItemsListCmd.Flags().
		StringVarP(&datasetItemsListOrg, "organization", "o", "", "Organization name")
	datasetItemsListCmd.Flags().
		StringVarP(&datasetItemsListProject, "project", "p", "", "Project name")

	datasetItemsGetCmd.Flags().
		BoolVar(&datasetItemsGetJSON, "json", false, "Output raw API response as JSON")
	datasetItemsGetCmd.Flags().
		StringVarP(&datasetItemsGetOrg, "organization", "o", "", "Organization name")
	datasetItemsGetCmd.Flags().
		StringVarP(&datasetItemsGetProject, "project", "p", "", "Project name")

	datasetItemsCreateCmd.Flags().
		StringVar(&datasetItemsCreateID, "id", "", "Explicit item ID (for upsert)")
	datasetItemsCreateCmd.Flags().
		StringVar(&datasetItemsCreateDatasetName, "dataset-name", "", "Dataset name (required)")
	_ = datasetItemsCreateCmd.MarkFlagRequired("dataset-name")
	datasetItemsCreateCmd.Flags().
		StringVar(&datasetItemsCreateInput, "input", "", "Input as JSON string")
	datasetItemsCreateCmd.Flags().
		StringVar(
			&datasetItemsCreateExpectedOutput, "expected-output", "",
			"Expected output as JSON string",
		)
	datasetItemsCreateCmd.Flags().
		StringVar(&datasetItemsCreateMetadataJSON, "metadata-json", "", "Metadata as JSON object")
	datasetItemsCreateCmd.Flags().
		StringVar(&datasetItemsCreateSourceTraceID, "source-trace-id", "", "Source trace ID")
	datasetItemsCreateCmd.Flags().
		StringVar(
			&datasetItemsCreateSourceObservationID, "source-observation-id", "",
			"Source observation ID",
		)
	datasetItemsCreateCmd.Flags().
		StringVar(&datasetItemsCreateStatus, "status", "", "Item status (ACTIVE/ARCHIVED)")
	datasetItemsCreateCmd.Flags().
		BoolVar(&datasetItemsCreateJSON, "json", false, "Output raw API response as JSON")
	datasetItemsCreateCmd.Flags().
		StringVarP(&datasetItemsCreateOrg, "organization", "o", "", "Organization name")
	datasetItemsCreateCmd.Flags().
		StringVarP(&datasetItemsCreateProject, "project", "p", "", "Project name")

	datasetItemsDeleteCmd.Flags().
		StringVarP(&datasetItemsDeleteOrg, "organization", "o", "", "Organization name")
	datasetItemsDeleteCmd.Flags().
		StringVarP(&datasetItemsDeleteProject, "project", "p", "", "Project name")

	datasetItemsCmd.AddCommand(
		datasetItemsListCmd,
		datasetItemsGetCmd,
		datasetItemsCreateCmd,
		datasetItemsDeleteCmd,
	)
	rootCmd.AddCommand(datasetItemsCmd)
}
