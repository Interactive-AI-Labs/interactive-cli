package cmd

import (
	"fmt"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	metricsListFromTimestamp string
	metricsListToTimestamp   string
	metricsListPage          int
	metricsListLimit         int
	metricsListTraceName     string
	metricsListUserID        string
	metricsListTags          []string
	metricsListEnvironment   string
	metricsListColumns       []string
	metricsListShowModels    bool
	metricsListJSON          bool
	metricsListOrg           string
	metricsListProject       string
	metricsListDaily         bool
)

var metricsCmd = &cobra.Command{
	Use:              "metrics",
	Aliases:          []string{"metric"},
	Short:            "Manage observability metrics",
	Long:             `Access observability metrics. Works with API key (--api-key or INTERACTIVE_API_KEY) or session from 'iai login'.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var metricsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List observability metrics",
	Long: `List observability metrics with optional filters.

Uses the platform API with dual authentication (API key or session).
If --from-timestamp is not provided, defaults to 7 days ago.
Use --daily to get metrics aggregated by day (default).

Examples:
  iai metrics list --daily
  iai metrics list --daily --from-timestamp 2025-01-01T00:00:00Z
  iai metrics list --daily --trace-name my-trace --show-models
  iai metrics list --daily --json | jq '.data'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if !metricsListDaily {
			return fmt.Errorf("a granularity flag is required (e.g. --daily)")
		}

		columns := metricsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultMetricsDailyColumns
		}
		if !metricsListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllMetricsDailyColumns); err != nil {
				return err
			}
		}

		fromTS := metricsListFromTimestamp
		if fromTS == "" {
			fromTS = time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
		}

		opts := clients.MetricsDailyOptions{
			FromTimestamp: fromTS,
			ToTimestamp:   metricsListToTimestamp,
			Page:          metricsListPage,
			Limit:         metricsListLimit,
			TraceName:     metricsListTraceName,
			UserID:        metricsListUserID,
			Tags:          metricsListTags,
			Environment:   metricsListEnvironment,
		}
		if err := inputs.ValidateMetricsDailyOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), metricsListOrg, metricsListProject)
		if err != nil {
			return err
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(
			hostname, defaultHTTPTimeout, apiKey, cookies,
		)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		metrics, meta, rawJSON, err := apiClient.ListMetricsDaily(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if metricsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintMetricsDaily(out, metrics, meta, columns, metricsListShowModels)
	},
}

func init() {
	metricsListCmd.Flags().
		BoolVar(&metricsListDaily, "daily", true, "Aggregate metrics by day")
	metricsListCmd.Flags().StringVar(
		&metricsListFromTimestamp,
		"from-timestamp",
		"",
		"Filter metrics from this timestamp (ISO 8601, default: 7 days ago)",
	)
	metricsListCmd.Flags().
		StringVar(&metricsListToTimestamp, "to-timestamp", "", "Filter metrics to this timestamp (ISO 8601)")
	metricsListCmd.Flags().IntVar(&metricsListPage, "page", 1, "Page number (starts at 1)")
	metricsListCmd.Flags().IntVar(&metricsListLimit, "limit", 0, "Items per page")
	metricsListCmd.Flags().
		StringVar(&metricsListTraceName, "trace-name", "", "Filter by trace name")
	metricsListCmd.Flags().StringVar(&metricsListUserID, "user-id", "", "Filter by user ID")
	metricsListCmd.Flags().
		StringArrayVar(&metricsListTags, "tags", nil, "Filter by tags (repeatable)")
	metricsListCmd.Flags().
		StringVar(&metricsListEnvironment, "environment", "", "Filter by environment")
	metricsListCmd.Flags().
		StringSliceVar(&metricsListColumns, "columns", nil, "Columns to display (comma-separated, default: date,count_traces,count_observations,total_cost)\nAvailable: date,count_traces,count_observations,total_cost,total_tokens")
	metricsListCmd.Flags().
		BoolVar(&metricsListShowModels, "show-models", false, "Show per-model breakdown")
	metricsListCmd.Flags().
		BoolVar(&metricsListJSON, "json", false, "Output raw API response as JSON")
	metricsListCmd.Flags().
		StringVarP(&metricsListOrg, "organization", "o", "", "Organization name that owns the project")
	metricsListCmd.Flags().
		StringVarP(&metricsListProject, "project", "p", "", "Project name")

	metricsCmd.AddCommand(metricsListCmd)
	rootCmd.AddCommand(metricsCmd)
}
