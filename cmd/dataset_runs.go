package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	datasetRunsListDataset string
	datasetRunsListPage    int
	datasetRunsListLimit   int
	datasetRunsListColumns []string
	datasetRunsListJSON    bool
	datasetRunsListOrg     string
	datasetRunsListProject string

	datasetRunsGetDataset string
	datasetRunsGetJSON    bool
	datasetRunsGetOrg     string
	datasetRunsGetProject string

	datasetRunsDeleteDataset string
	datasetRunsDeleteOrg     string
	datasetRunsDeleteProject string
)

var datasetRunsCmd = &cobra.Command{
	Use:              "dataset-runs",
	Aliases:          []string{"dataset-run"},
	Short:            "Manage dataset runs",
	Long:             `Manage evaluation runs within datasets.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var datasetRunsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List dataset runs",
	Long:    `List runs for a given dataset.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		datasetName := strings.TrimSpace(datasetRunsListDataset)

		columns := datasetRunsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultDatasetRunColumns
		}
		if !datasetRunsListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllDatasetRunColumns); err != nil {
				return err
			}
		}

		opts := clients.DatasetRunListOptions{
			Page:  datasetRunsListPage,
			Limit: datasetRunsListLimit,
		}
		if err := inputs.ValidateDatasetRunListOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(
			cmd.Context(),
			datasetRunsListOrg,
			datasetRunsListProject,
		)
		if err != nil {
			return err
		}

		runs, meta, rawJSON, err := pCtx.apiClient.ListDatasetRuns(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			datasetName,
			opts,
		)
		if err != nil {
			return err
		}

		if datasetRunsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintDatasetRunList(out, runs, meta, columns)
	},
}

var datasetRunsGetCmd = &cobra.Command{
	Use:     "get <run-name>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a dataset run",
	Long:    `Get detailed information about a specific dataset run.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		datasetName := strings.TrimSpace(datasetRunsGetDataset)
		runName := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(
			cmd.Context(),
			datasetRunsGetOrg,
			datasetRunsGetProject,
		)
		if err != nil {
			return err
		}

		run, rawJSON, err := pCtx.apiClient.GetDatasetRun(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			datasetName,
			runName,
		)
		if err != nil {
			return err
		}

		if datasetRunsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintDatasetRunDetail(out, run)
	},
}

var datasetRunsDeleteCmd = &cobra.Command{
	Use:     "delete <run-name>",
	Aliases: []string{"rm"},
	Short:   "Delete a dataset run",
	Long:    `Delete a dataset run by name.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		datasetName := strings.TrimSpace(datasetRunsDeleteDataset)
		runName := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(
			cmd.Context(),
			datasetRunsDeleteOrg,
			datasetRunsDeleteProject,
		)
		if err != nil {
			return err
		}

		message, err := pCtx.apiClient.DeleteDatasetRun(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			datasetName,
			runName,
		)
		if err != nil {
			return err
		}

		return output.PrintDeleteSuccess(out, runName, "dataset run", message)
	},
}

func init() {
	datasetRunsListCmd.Flags().
		StringVar(&datasetRunsListDataset, "dataset-name", "", "Dataset name (required)")
	_ = datasetRunsListCmd.MarkFlagRequired("dataset-name")
	datasetRunsListCmd.Flags().
		IntVar(&datasetRunsListPage, "page", 1, "Page number (starts at 1)")
	datasetRunsListCmd.Flags().IntVar(&datasetRunsListLimit, "limit", 50, "Items per page")
	datasetRunsListCmd.Flags().
		StringSliceVar(&datasetRunsListColumns, "columns", nil, "Columns to display (comma-separated)")
	datasetRunsListCmd.Flags().
		BoolVar(&datasetRunsListJSON, "json", false, "Output raw API response as JSON")
	datasetRunsListCmd.Flags().
		StringVarP(&datasetRunsListOrg, "organization", "o", "", "Organization name that owns the project")
	datasetRunsListCmd.Flags().
		StringVarP(&datasetRunsListProject, "project", "p", "", "Project name")

	datasetRunsGetCmd.Flags().
		StringVar(&datasetRunsGetDataset, "dataset-name", "", "Dataset name (required)")
	_ = datasetRunsGetCmd.MarkFlagRequired("dataset-name")
	datasetRunsGetCmd.Flags().
		BoolVar(&datasetRunsGetJSON, "json", false, "Output raw API response as JSON")
	datasetRunsGetCmd.Flags().
		StringVarP(&datasetRunsGetOrg, "organization", "o", "", "Organization name that owns the project")
	datasetRunsGetCmd.Flags().
		StringVarP(&datasetRunsGetProject, "project", "p", "", "Project name")

	datasetRunsDeleteCmd.Flags().
		StringVar(&datasetRunsDeleteDataset, "dataset-name", "", "Dataset name (required)")
	_ = datasetRunsDeleteCmd.MarkFlagRequired("dataset-name")
	datasetRunsDeleteCmd.Flags().
		StringVarP(&datasetRunsDeleteOrg, "organization", "o", "", "Organization name that owns the project")
	datasetRunsDeleteCmd.Flags().
		StringVarP(&datasetRunsDeleteProject, "project", "p", "", "Project name")

	datasetRunsCmd.AddCommand(datasetRunsListCmd, datasetRunsGetCmd, datasetRunsDeleteCmd)
	rootCmd.AddCommand(datasetRunsCmd)
}
