package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
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
	datasetItemsListYAML                bool
	datasetItemsListOrg                 string
	datasetItemsListProject             string

	datasetItemsGetJSON    bool
	datasetItemsGetYAML    bool
	datasetItemsGetOrg     string
	datasetItemsGetProject string

	datasetItemsCreateID                  string
	datasetItemsCreateName                string
	datasetItemsCreateDatasetName         string
	datasetItemsCreateInput               string
	datasetItemsCreateExpectedOutput      string
	datasetItemsCreateMetadataJSON        string
	datasetItemsCreateSourceTraceID       string
	datasetItemsCreateSourceObservationID string
	datasetItemsCreateStatus              string
	datasetItemsCreateJSON                bool
	datasetItemsCreateYAML                bool
	datasetItemsCreateOrg                 string
	datasetItemsCreateProject             string

	datasetItemsUpdateName         string
	datasetItemsUpdateStatus       string
	datasetItemsUpdateMetadataJSON string
	datasetItemsUpdateJSON         bool
	datasetItemsUpdateYAML         bool
	datasetItemsUpdateOrg          string
	datasetItemsUpdateProject      string

	datasetItemsDeleteOrg     string
	datasetItemsDeleteProject string
)

var datasetItemsCmd = &cobra.Command{
	Use:              "dataset-items",
	Aliases:          []string{"dataset-item"},
	Short:            "Manage items in evaluation datasets",
	GroupID:          groupEvaluation,
	Long:             `Manage individual items within evaluation datasets.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var datasetItemsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List dataset items",
	Long:    `List items in a dataset with optional filters.`,
	Example: `  iai dataset-items list --dataset-name my-dataset
  iai dataset-items list --dataset-name my-dataset --source-trace-id trace-123 --limit 50
  iai dataset-items list --dataset-name my-dataset --columns id,name,status
  iai dataset-items list --dataset-name my-dataset --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := datasetItemsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultDatasetItemColumns
		}
		if err := validateTableOnlyColumns(
			cmd,
			datasetItemsListJSON,
			datasetItemsListYAML,
		); err != nil {
			return err
		}
		if !datasetItemsListJSON && !datasetItemsListYAML {
			if err := inputs.ValidateColumns(columns, inputs.AllDatasetItemColumns); err != nil {
				return err
			}
		}

		opts := platform.DatasetItemListOptions{
			DatasetName:         datasetItemsListDatasetName,
			SourceTraceID:       datasetItemsListSourceTraceID,
			SourceObservationID: datasetItemsListSourceObservationID,
			Page:                datasetItemsListPage,
			Limit:               datasetItemsListLimit,
		}
		if err := inputs.ValidateDatasetItemListOptions(opts); err != nil {
			return err
		}

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			datasetItemsListOrg,
			datasetItemsListProject,
		)
		if err != nil {
			return err
		}

		items, meta, rawJSON, err := apiClient.ListDatasetItems(
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

		if datasetItemsListYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintDatasetItemList(out, items, meta, columns)
	},
}

var datasetItemsGetCmd = &cobra.Command{
	Use:     "get <id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a dataset item",
	Long:    `Get detailed information about a specific dataset item. Takes the item id only; there is no --dataset-name flag.`,
	Example: `  iai dataset-items get item-123
  iai dataset-items get item-123 --json
  iai dataset-items get item-123 --yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		itemID := strings.TrimSpace(args[0])

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			datasetItemsGetOrg,
			datasetItemsGetProject,
		)
		if err != nil {
			return err
		}

		item, rawJSON, err := apiClient.GetDatasetItem(
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

		if datasetItemsGetYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintDatasetItemDetail(out, item)
	},
}

var datasetItemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a dataset item",
	Long:  `Create or upsert an item in a dataset.`,
	Example: `  iai dataset-items create --dataset-name my-dataset --input '{"question":"2+2?"}'
  iai dataset-items create --dataset-name my-dataset --name defensores-37-pdf-excel-v2 --input '{"q":"hi"}'
  iai dataset-items create --dataset-name my-dataset --input '{"q":"hi"}' --status ARCHIVED
  iai dataset-items create --dataset-name my-dataset --input '{"question":"2+2?"}' --expected-output '{"answer":"4"}'
  iai dataset-items create --dataset-name my-dataset --input '{"q":"hi"}' --metadata-json '{"source":"manual"}'
  iai dataset-items create --dataset-name my-dataset --input '{"q":"hi"}' --source-trace-id trace-123 --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		body, err := inputs.BuildDatasetItemCreateBody(inputs.DatasetItemCreateInput{
			ID:                  datasetItemsCreateID,
			Name:                datasetItemsCreateName,
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

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			datasetItemsCreateOrg,
			datasetItemsCreateProject,
		)
		if err != nil {
			return err
		}

		item, rawJSON, err := apiClient.CreateDatasetItem(
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

		if datasetItemsCreateYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintDatasetItemCreateResult(out, item)
	},
}

var datasetItemsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a dataset item",
	Long: `Change the name, status or metadata of a dataset item. Flags you leave out keep their current value.
--name "" removes the name. --metadata-json replaces the metadata and keeps the name.`,
	Example: `  iai dataset-items update item-123 --name defensores-37-pdf-excel-v2
  iai dataset-items update item-123 --status ARCHIVED
  iai dataset-items update item-123 --status ACTIVE --metadata-json '{"source":"manual"}'
  iai dataset-items update item-123 --name ""`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		itemID := strings.TrimSpace(args[0])

		var name *string
		if cmd.Flags().Changed("name") {
			name = &datasetItemsUpdateName
		}
		body, err := inputs.BuildDatasetItemUpdateBody(inputs.DatasetItemUpdateInput{
			Name:         name,
			Status:       datasetItemsUpdateStatus,
			MetadataJSON: datasetItemsUpdateMetadataJSON,
		})
		if err != nil {
			return err
		}

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			datasetItemsUpdateOrg,
			datasetItemsUpdateProject,
		)
		if err != nil {
			return err
		}

		item, rawJSON, err := apiClient.UpdateDatasetItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			itemID,
			body,
		)
		if err != nil {
			return err
		}

		if datasetItemsUpdateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		if datasetItemsUpdateYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintDatasetItemDetail(out, item)
	},
}

var datasetItemsDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm"},
	Short:   "Delete a dataset item",
	Long:    `Delete a dataset item by ID.`,
	Example: `  iai dataset-items delete item-123
  iai dataset-items delete item-123 --organization my-org --project my-project`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		itemID := strings.TrimSpace(args[0])

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			datasetItemsDeleteOrg,
			datasetItemsDeleteProject,
		)
		if err != nil {
			return err
		}

		message, err := apiClient.DeleteDatasetItem(
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
	indexJSONShape(datasetItemsCmd, "env")
	datasetItemsListCmd.Flags().
		StringVar(&datasetItemsListDatasetName, "dataset-name", "", "Dataset name (required)")
	_ = datasetItemsListCmd.MarkFlagRequired("dataset-name")
	datasetItemsListCmd.Flags().
		StringVar(&datasetItemsListSourceTraceID, "source-trace-id", "", "Filter by source trace ID")
	datasetItemsListCmd.Flags().
		StringVar(
			&datasetItemsListSourceObservationID, "source-observation-id", "",
			"Filter by source observation ID",
		)
	datasetItemsListCmd.Flags().
		IntVar(&datasetItemsListPage, "page", 1, "Page number (starts at 1)")
	datasetItemsListCmd.Flags().
		IntVar(&datasetItemsListLimit, "limit", 0, "Items per page (max 100)")
	indexFlag(datasetItemsListCmd, "limit", "max 100")
	datasetItemsListCmd.Flags().
		StringSliceVar(&datasetItemsListColumns, "columns", nil, "Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml")
	datasetItemsListCmd.Flags().
		BoolVar(&datasetItemsListJSON, "json", false, "Output raw API response as JSON")
	datasetItemsListCmd.Flags().
		BoolVar(&datasetItemsListYAML, "yaml", false, "Output raw API response as YAML")
	datasetItemsListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	datasetItemsListCmd.Flags().
		StringVarP(&datasetItemsListOrg, "organization", "o", "", "Organization name that owns the project")
	datasetItemsListCmd.Flags().
		StringVarP(&datasetItemsListProject, "project", "p", "", "Project name")

	indexCommand(datasetItemsGetCmd, "no --dataset-name")
	datasetItemsGetCmd.Flags().
		BoolVar(&datasetItemsGetJSON, "json", false, "Output raw API response as JSON")
	datasetItemsGetCmd.Flags().
		BoolVar(&datasetItemsGetYAML, "yaml", false, "Output raw API response as YAML")
	datasetItemsGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	datasetItemsGetCmd.Flags().
		StringVarP(&datasetItemsGetOrg, "organization", "o", "", "Organization name that owns the project")
	datasetItemsGetCmd.Flags().
		StringVarP(&datasetItemsGetProject, "project", "p", "", "Project name")

	datasetItemsCreateCmd.Flags().
		StringVar(&datasetItemsCreateID, "id", "", "Explicit item ID; upserts an existing item (needs datasets:update)")
	datasetItemsCreateCmd.Flags().
		StringVar(&datasetItemsCreateName, "name", "", "Item name")
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
	addChoiceFlag(datasetItemsCreateCmd, &datasetItemsCreateStatus, "status", "",
		"Item status; ARCHIVED needs datasets:update", datasetItemStatuses)
	datasetItemsCreateCmd.Flags().
		BoolVar(&datasetItemsCreateJSON, "json", false, "Output raw API response as JSON")
	datasetItemsCreateCmd.Flags().
		BoolVar(&datasetItemsCreateYAML, "yaml", false, "Output raw API response as YAML")
	datasetItemsCreateCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	datasetItemsCreateCmd.Flags().
		StringVarP(&datasetItemsCreateOrg, "organization", "o", "", "Organization name that owns the project")
	datasetItemsCreateCmd.Flags().
		StringVarP(&datasetItemsCreateProject, "project", "p", "", "Project name")

	datasetItemsUpdateCmd.Flags().
		StringVar(&datasetItemsUpdateName, "name", "", `Item name ("" removes it)`)
	addChoiceFlag(datasetItemsUpdateCmd, &datasetItemsUpdateStatus, "status", "",
		"Item status", datasetItemStatuses)
	datasetItemsUpdateCmd.Flags().
		StringVar(&datasetItemsUpdateMetadataJSON, "metadata-json", "", "Metadata as JSON object (replaces the current metadata)")
	datasetItemsUpdateCmd.Flags().
		BoolVar(&datasetItemsUpdateJSON, "json", false, "Output raw API response as JSON")
	datasetItemsUpdateCmd.Flags().
		BoolVar(&datasetItemsUpdateYAML, "yaml", false, "Output raw API response as YAML")
	datasetItemsUpdateCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	datasetItemsUpdateCmd.Flags().
		StringVarP(&datasetItemsUpdateOrg, "organization", "o", "", "Organization name that owns the project")
	datasetItemsUpdateCmd.Flags().
		StringVarP(&datasetItemsUpdateProject, "project", "p", "", "Project name")

	datasetItemsDeleteCmd.Flags().
		StringVarP(&datasetItemsDeleteOrg, "organization", "o", "", "Organization name that owns the project")
	datasetItemsDeleteCmd.Flags().
		StringVarP(&datasetItemsDeleteProject, "project", "p", "", "Project name")

	datasetItemsCmd.AddCommand(
		datasetItemsListCmd,
		datasetItemsGetCmd,
		datasetItemsCreateCmd,
		datasetItemsUpdateCmd,
		datasetItemsDeleteCmd,
	)
	rootCmd.AddCommand(datasetItemsCmd)
}

var datasetItemStatuses = []string{"ACTIVE", "ARCHIVED"}
