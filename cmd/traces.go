package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	tracesPage          int
	tracesLimit         int
	tracesUserID        string
	tracesName          string
	tracesSessionID     string
	tracesFromTimestamp string
	tracesToTimestamp   string
	tracesOrderBy       string
	tracesOrder         string
	tracesTags          []string
	tracesVersion       string
	tracesRelease       string
	tracesEnvironment   []string
	tracesColumns       []string
	tracesMinCost       float64
	tracesMaxCost       float64
	tracesMinLatency    float64
	tracesMaxLatency    float64
	tracesMinTokens     int
	tracesMaxTokens     int
	tracesModel         string
	tracesHasError      bool
	tracesLevel         string
	tracesSearch        string
	tracesFields        string
	tracesJSON          bool
	tracesGetFields     string
	tracesGetJSON       bool
	tracesListOrg       string
	tracesListProject   string
	tracesGetOrg        string
	tracesGetProject    string
)

var tracesCmd = &cobra.Command{
	Use:     "traces",
	Aliases: []string{"trace"},
	Short:   "Manage traces",
	Long:    `Manage traces. Works with API key (--api-key or INTERACTIVE_API_KEY) or session from 'iai login'.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Cobra doesn't chain PersistentPreRun hooks; call the parent's manually to preserve URL normalization.
		if root := cmd.Root(); root != nil && root.PersistentPreRun != nil {
			root.PersistentPreRun(cmd, args)
		}
	},
}

var tracesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List traces",
	Long: `List traces with optional filters.

Uses the platform API with dual authentication (API key or session).
If --from-timestamp is not provided, defaults to 7 days ago.

Examples:
  iai traces list
  iai traces list --limit 20 --page 2
  iai traces list --name my-trace --user-id user123
  iai traces list --from-timestamp 2025-01-01T00:00:00Z
  iai traces list --order-by timestamp --order desc
  iai traces list --tags tag1 --tags tag2
  iai traces list --model gpt-4 --has-error
  iai traces list --min-cost 0.01 --max-cost 1.0
  iai traces list --level ERROR
  iai traces list --search "my query"
  iai traces list --fields core,io,metrics
  iai traces list --json | jq '.data.traces[].name'
  iai traces list --columns id,name,latency,total_tokens,level`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := tracesColumns
		if len(columns) == 0 {
			columns = inputs.DefaultTraceColumns
		}
		if !tracesJSON {
			if err := inputs.ValidateTraceColumns(columns); err != nil {
				return err
			}
		}

		// Default --from-timestamp to 7 days ago if not set.
		fromTS := tracesFromTimestamp
		if fromTS == "" {
			fromTS = time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
		}

		opts := clients.TraceListOptions{
			Page:          tracesPage,
			Limit:         tracesLimit,
			UserID:        tracesUserID,
			Name:          tracesName,
			SessionID:     tracesSessionID,
			FromTimestamp: fromTS,
			ToTimestamp:   tracesToTimestamp,
			OrderBy:       tracesOrderBy,
			Order:         tracesOrder,
			Tags:          tracesTags,
			Version:       tracesVersion,
			Release:       tracesRelease,
			Environment:   tracesEnvironment,
			Model:         tracesModel,
			Level:         tracesLevel,
			Search:        tracesSearch,
			Fields:        tracesFields,
		}

		if cmd.Flags().Changed("min-cost") {
			opts.MinCost = &tracesMinCost
		}
		if cmd.Flags().Changed("max-cost") {
			opts.MaxCost = &tracesMaxCost
		}
		if cmd.Flags().Changed("min-latency") {
			opts.MinLatency = &tracesMinLatency
		}
		if cmd.Flags().Changed("max-latency") {
			opts.MaxLatency = &tracesMaxLatency
		}
		if cmd.Flags().Changed("min-tokens") {
			opts.MinTokens = &tracesMinTokens
		}
		if cmd.Flags().Changed("max-tokens") {
			opts.MaxTokens = &tracesMaxTokens
		}
		if cmd.Flags().Changed("has-error") {
			opts.HasError = &tracesHasError
		}

		if err := inputs.ValidateTraceListOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), tracesListOrg, tracesListProject)
		if err != nil {
			return err
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		traces, meta, rawJSON, err := apiClient.ListTraces(
			cmd.Context(), pCtx.orgId, pCtx.projectId, opts,
		)
		if err != nil {
			return err
		}

		if tracesJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintTraceList(out, traces, meta, columns)
	},
}

var tracesGetCmd = &cobra.Command{
	Use:   "get <trace-id>",
	Short: "Get a specific trace",
	Long: `Get detailed information about a specific trace.

Uses the platform API with dual authentication (API key or session).

Examples:
  iai traces get abc123
  iai traces get abc123 --fields core,io,metrics
  iai traces get abc123 --json | jq '.data.trace'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		traceID := strings.TrimSpace(args[0])
		if err := inputs.ValidateTraceID(traceID); err != nil {
			return err
		}
		pCtx, err := resolveProject(cmd.Context(), tracesGetOrg, tracesGetProject)
		if err != nil {
			return err
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		trace, rawJSON, err := apiClient.GetTrace(
			cmd.Context(), pCtx.orgId, pCtx.projectId, traceID, tracesGetFields,
		)
		if err != nil {
			return err
		}

		if tracesGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintTraceDetail(out, trace)
	},
}

func init() {
	// traces list flags
	tracesListCmd.Flags().IntVar(&tracesPage, "page", 1, "Page number (starts at 1)")
	tracesListCmd.Flags().IntVar(&tracesLimit, "limit", 0, "Items per page")
	tracesListCmd.Flags().StringVar(&tracesUserID, "user-id", "", "Filter by user ID")
	tracesListCmd.Flags().StringVar(&tracesName, "name", "", "Filter by trace name")
	tracesListCmd.Flags().StringVar(&tracesSessionID, "session-id", "", "Filter by session ID")
	tracesListCmd.Flags().
		StringVar(&tracesFromTimestamp, "from-timestamp", "", "Filter traces from this timestamp (ISO 8601, default: 7 days ago)")
	tracesListCmd.Flags().
		StringVar(&tracesToTimestamp, "to-timestamp", "", "Filter traces to this timestamp (ISO 8601)")
	tracesListCmd.Flags().
		StringVar(&tracesOrderBy, "order-by", "", "Order by field: timestamp, latency, cost, name")
	tracesListCmd.Flags().
		StringVar(&tracesOrder, "order", "desc", "Sort direction: asc or desc (default: desc)")
	tracesListCmd.Flags().StringArrayVar(&tracesTags, "tags", nil, "Filter by tags (repeatable)")
	tracesListCmd.Flags().StringVar(&tracesVersion, "version", "", "Filter by version")
	tracesListCmd.Flags().StringVar(&tracesRelease, "release", "", "Filter by release")
	tracesListCmd.Flags().
		StringArrayVar(&tracesEnvironment, "environment", nil, "Filter by environment (repeatable)")

	// New filtering flags
	tracesListCmd.Flags().Float64Var(&tracesMinCost, "min-cost", 0, "Minimum total cost filter")
	tracesListCmd.Flags().Float64Var(&tracesMaxCost, "max-cost", 0, "Maximum total cost filter")
	tracesListCmd.Flags().
		Float64Var(&tracesMinLatency, "min-latency", 0, "Minimum latency filter (seconds)")
	tracesListCmd.Flags().
		Float64Var(&tracesMaxLatency, "max-latency", 0, "Maximum latency filter (seconds)")
	tracesListCmd.Flags().IntVar(&tracesMinTokens, "min-tokens", 0, "Minimum total tokens filter")
	tracesListCmd.Flags().IntVar(&tracesMaxTokens, "max-tokens", 0, "Maximum total tokens filter")
	tracesListCmd.Flags().StringVar(&tracesModel, "model", "", "Filter by model name")
	tracesListCmd.Flags().BoolVar(&tracesHasError, "has-error", false, "Filter traces with errors")
	tracesListCmd.Flags().
		StringVar(&tracesLevel, "level", "", "Filter by aggregated level: DEBUG, DEFAULT, WARNING, ERROR")
	tracesListCmd.Flags().
		StringVar(&tracesSearch, "search", "", "Search in trace name (max 200 characters)")
	tracesListCmd.Flags().
		StringVar(&tracesFields, "fields", "core,metrics", "Field groups to include: core, io, metrics (comma-separated)")

	// Output flags
	tracesListCmd.Flags().BoolVar(&tracesJSON, "json", false, "Output raw API response as JSON")
	// StringSliceVar (not StringArrayVar) so users can pass --columns id,name,cost as a comma-separated list.
	// --tags and --environment use StringArrayVar to avoid splitting values that may contain commas.
	tracesListCmd.Flags().
		StringSliceVar(&tracesColumns, "columns", nil, "Columns to display (comma-separated, default: id,name,timestamp,latency,cost,tags)\nAvailable: id,name,timestamp,user_id,session_id,release,version,environment,public,latency,cost,tags,observation_count,input_tokens,output_tokens,total_tokens,level")

	// Org/project flags
	tracesListCmd.Flags().
		StringVarP(&tracesListOrg, "organization", "o", "", "Organization name that owns the project")
	tracesListCmd.Flags().
		StringVarP(&tracesListProject, "project", "p", "", "Project name")

	// traces get flags
	tracesGetCmd.Flags().
		StringVar(&tracesGetFields, "fields", "core,io,metrics", "Field groups to include: core, io, metrics (comma-separated)")
	tracesGetCmd.Flags().BoolVar(&tracesGetJSON, "json", false, "Output raw API response as JSON")
	tracesGetCmd.Flags().
		StringVarP(&tracesGetOrg, "organization", "o", "", "Organization name that owns the project")
	tracesGetCmd.Flags().
		StringVarP(&tracesGetProject, "project", "p", "", "Project name")

	tracesCmd.AddCommand(tracesListCmd, tracesGetCmd)
	rootCmd.AddCommand(tracesCmd)
}
