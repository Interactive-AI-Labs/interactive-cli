package cmd

import (
	"fmt"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
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
	tracesTags          []string
	tracesVersion       string
	tracesRelease       string
	tracesEnvironment   []string
	tracesColumns       []string
)

var tracesCmd = &cobra.Command{
	Use:     "traces",
	Aliases: []string{"trace"},
	Short:   "Manage traces",
	Long:    `Manage traces. Requires an API key (--api-key or INTERACTIVE_API_KEY).`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if apiKey == "" {
			return fmt.Errorf("traces commands require API key authentication; use --api-key or set INTERACTIVE_API_KEY")
		}
		return nil
	},
}

var tracesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List traces",
	Long: `List traces with optional filters.

Examples:
  iai traces list
  iai traces list --limit 20 --page 2
  iai traces list --name my-trace --user-id user123
  iai traces list --from-timestamp 2025-01-01T00:00:00Z
  iai traces list --order-by timestamp.desc
  iai traces list --tags tag1 --tags tag2`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := tracesColumns
		if len(columns) == 0 {
			columns = inputs.DefaultTraceColumns
		}
		if err := inputs.ValidateTraceColumns(columns); err != nil {
			return err
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, nil)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		traces, meta, err := apiClient.ListTraces(
			cmd.Context(),
			clients.TraceListOptions{
				Page:          tracesPage,
				Limit:         tracesLimit,
				UserID:        tracesUserID,
				Name:          tracesName,
				SessionID:     tracesSessionID,
				FromTimestamp: tracesFromTimestamp,
				ToTimestamp:   tracesToTimestamp,
				OrderBy:       tracesOrderBy,
				Tags:          tracesTags,
				Version:       tracesVersion,
				Release:       tracesRelease,
				Environment:   tracesEnvironment,
			},
		)
		if err != nil {
			return err
		}

		return output.PrintTraceList(out, traces, meta, columns)
	},
}

var tracesGetCmd = &cobra.Command{
	Use:   "get <trace-id>",
	Short: "Get a specific trace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		traceID := strings.TrimSpace(args[0])
		if err := inputs.ValidateTraceID(traceID); err != nil {
			return err
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, nil)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		trace, err := apiClient.GetTrace(cmd.Context(), traceID)
		if err != nil {
			return fmt.Errorf("failed to get trace %q: %w", traceID, err)
		}

		return output.PrintTraceDetail(out, trace)
	},
}

var tracesDeleteCmd = &cobra.Command{
	Use:     "delete <trace-id>",
	Aliases: []string{"rm"},
	Short:   "Delete a specific trace",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		traceID := strings.TrimSpace(args[0])
		if err := inputs.ValidateTraceID(traceID); err != nil {
			return err
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, nil)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting trace delete request...")

		serverMessage, err := apiClient.DeleteTrace(cmd.Context(), traceID)
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
	// traces list filters
	tracesListCmd.Flags().IntVar(&tracesPage, "page", 0, "Page number (starts at 1)")
	tracesListCmd.Flags().IntVar(&tracesLimit, "limit", 0, "Items per page")
	tracesListCmd.Flags().StringVar(&tracesUserID, "user-id", "", "Filter by user ID")
	tracesListCmd.Flags().StringVar(&tracesName, "name", "", "Filter by trace name")
	tracesListCmd.Flags().StringVar(&tracesSessionID, "session-id", "", "Filter by session ID")
	tracesListCmd.Flags().StringVar(&tracesFromTimestamp, "from-timestamp", "", "Filter traces from this timestamp (ISO 8601)")
	tracesListCmd.Flags().StringVar(&tracesToTimestamp, "to-timestamp", "", "Filter traces to this timestamp (ISO 8601)")
	tracesListCmd.Flags().StringVar(&tracesOrderBy, "order-by", "", "Order by field.direction (e.g. timestamp.desc)")
	tracesListCmd.Flags().StringArrayVar(&tracesTags, "tags", nil, "Filter by tags (repeatable)")
	tracesListCmd.Flags().StringVar(&tracesVersion, "version", "", "Filter by version")
	tracesListCmd.Flags().StringVar(&tracesRelease, "release", "", "Filter by release")
	tracesListCmd.Flags().StringArrayVar(&tracesEnvironment, "environment", nil, "Filter by environment (repeatable)")
	tracesListCmd.Flags().StringSliceVar(&tracesColumns, "columns", nil, "Columns to display (default: id,name,timestamp,latency,cost,tags)\nAvailable: id,name,timestamp,user_id,session_id,release,version,environment,public,latency,cost,tags")

	tracesCmd.AddCommand(tracesListCmd, tracesGetCmd, tracesDeleteCmd)
	rootCmd.AddCommand(tracesCmd)
}
