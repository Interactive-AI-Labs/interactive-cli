package cmd

import (
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	runItemsListRunName     string
	runItemsListDatasetName string
	runItemsListPage        int
	runItemsListLimit       int
	runItemsListColumns     []string
	runItemsListJSON        bool
	runItemsListOrg         string
	runItemsListProject     string

	runItemsCreateRunName        string
	runItemsCreateRunDescription string
	runItemsCreateDatasetItemID  string
	runItemsCreateTraceID        string
	runItemsCreateObservationID  string
	runItemsCreateMetadataJSON   string
	runItemsCreateJSON           bool
	runItemsCreateOrg            string
	runItemsCreateProject        string
)

var runItemsCmd = &cobra.Command{
	Use:              "run-items",
	Aliases:          []string{"run-item"},
	Short:            "Manage dataset run items",
	Long:             `Manage items within dataset runs.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var runItemsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List run items",
	Long:    `List dataset run items. Requires at least one of --run-name or --dataset-name.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := runItemsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultRunItemColumns
		}
		if !runItemsListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllRunItemColumns); err != nil {
				return err
			}
		}

		opts := clients.DatasetRunItemListOptions{
			RunName:     runItemsListRunName,
			DatasetName: runItemsListDatasetName,
			Page:        runItemsListPage,
			Limit:       runItemsListLimit,
		}
		if err := inputs.ValidateRunItemListOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), runItemsListOrg, runItemsListProject)
		if err != nil {
			return err
		}

		items, meta, rawJSON, err := pCtx.apiClient.ListDatasetRunItems(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if runItemsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintRunItemList(out, items, meta, columns)
	},
}

var runItemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a run item",
	Long: `Create a new dataset run item linking a trace/observation to a dataset item.

This command requires API key authentication.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		body, err := inputs.BuildRunItemCreateBody(inputs.RunItemCreateInput{
			RunName:        runItemsCreateRunName,
			RunDescription: runItemsCreateRunDescription,
			DatasetItemID:  runItemsCreateDatasetItemID,
			TraceID:        runItemsCreateTraceID,
			ObservationID:  runItemsCreateObservationID,
			MetadataJSON:   runItemsCreateMetadataJSON,
		})
		if err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), runItemsCreateOrg, runItemsCreateProject)
		if err != nil {
			return err
		}

		item, rawJSON, err := pCtx.apiClient.CreateDatasetRunItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			body,
		)
		if err != nil {
			return err
		}

		if runItemsCreateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintRunItemCreateResult(out, item)
	},
}

func init() {
	runItemsListCmd.Flags().
		StringVar(&runItemsListRunName, "run-name", "", "Filter by run name")
	runItemsListCmd.Flags().
		StringVar(&runItemsListDatasetName, "dataset-name", "", "Filter by dataset name")
	runItemsListCmd.Flags().IntVar(&runItemsListPage, "page", 1, "Page number (starts at 1)")
	runItemsListCmd.Flags().IntVar(&runItemsListLimit, "limit", 50, "Items per page")
	runItemsListCmd.Flags().
		StringSliceVar(&runItemsListColumns, "columns", nil, "Columns to display (comma-separated)")
	runItemsListCmd.Flags().
		BoolVar(&runItemsListJSON, "json", false, "Output raw API response as JSON")
	runItemsListCmd.Flags().
		StringVarP(&runItemsListOrg, "organization", "o", "", "Organization name")
	runItemsListCmd.Flags().
		StringVarP(&runItemsListProject, "project", "p", "", "Project name")

	runItemsCreateCmd.Flags().
		StringVar(&runItemsCreateRunName, "run-name", "", "Run name (required)")
	runItemsCreateCmd.Flags().
		StringVar(&runItemsCreateDatasetItemID, "dataset-item-id", "", "Dataset item ID (required)")
	_ = runItemsCreateCmd.MarkFlagRequired("run-name")
	_ = runItemsCreateCmd.MarkFlagRequired("dataset-item-id")
	runItemsCreateCmd.Flags().
		StringVar(&runItemsCreateRunDescription, "run-description", "", "Run description")
	runItemsCreateCmd.Flags().
		StringVar(&runItemsCreateTraceID, "trace-id", "", "Trace ID")
	runItemsCreateCmd.Flags().
		StringVar(&runItemsCreateObservationID, "observation-id", "", "Observation ID")
	runItemsCreateCmd.Flags().
		StringVar(&runItemsCreateMetadataJSON, "metadata-json", "", "Metadata as JSON object")
	runItemsCreateCmd.Flags().
		BoolVar(&runItemsCreateJSON, "json", false, "Output raw API response as JSON")
	runItemsCreateCmd.Flags().
		StringVarP(&runItemsCreateOrg, "organization", "o", "", "Organization name")
	runItemsCreateCmd.Flags().
		StringVarP(&runItemsCreateProject, "project", "p", "", "Project name")

	runItemsCmd.AddCommand(runItemsListCmd, runItemsCreateCmd)
	rootCmd.AddCommand(runItemsCmd)
}
