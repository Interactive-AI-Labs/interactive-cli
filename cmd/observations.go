package cmd

import (
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	obsTraceID     string
	obsIncludeIO   bool
	obsColumns     []string
	obsListJSON    bool
	obsGetJSON     bool
	obsListOrg     string
	obsListProject string
	obsGetOrg      string
	obsGetProject  string

	obsListFromTimestamp       string
	obsListToTimestamp         string
	obsListCursor              string
	obsListLimit               int
	obsListFields              string
	obsListType                string
	obsListName                string
	obsListLevel               string
	obsListModel               string
	obsListEnvironment         string
	obsListParentObservationID string
	obsListVersion             string
	obsListUserID              string
)

var observationsCmd = &cobra.Command{
	Use:              "observations",
	Aliases:          []string{"obs", "observation"},
	Short:            "Manage observations",
	Long:             `Manage observations within traces. Works with API key (--api-key or INTERACTIVE_API_KEY) or session from 'iai login'.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var obsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List observations",
	Long: `List observations for a specific trace or search across traces with filters.

When --trace-id is provided, lists observations within that trace.
Without --trace-id, searches observations across all traces with optional filters.

Uses the platform API with dual authentication (API key or session).

Examples:
  # List observations for a specific trace
  iai observations list --trace-id abc123
  iai observations list --trace-id abc123 --include-io
  iai observations list --trace-id abc123 --columns id,type,name,model,latency_ms

  # Search observations across traces
  iai observations list --type GENERATION --model gpt-4
  iai observations list --from-timestamp 2025-01-01T00:00:00Z --name my-span
  iai observations list --json | jq '.data'`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		traceID := strings.TrimSpace(obsTraceID)

		// When --trace-id is provided, use the trace-scoped endpoint
		if traceID != "" {
			columns := obsColumns
			if len(columns) == 0 {
				columns = inputs.DefaultObservationColumns
			}
			if !obsListJSON {
				if err := inputs.ValidateColumns(
					columns,
					inputs.AllObservationColumns,
				); err != nil {
					return err
				}
			}

			pCtx, err := resolveProject(cmd.Context(), obsListOrg, obsListProject)
			if err != nil {
				return err
			}

			observations, rawJSON, err := pCtx.apiClient.ListObservations(
				cmd.Context(), pCtx.orgId, pCtx.projectId, traceID, obsIncludeIO,
			)
			if err != nil {
				return err
			}

			if obsListJSON {
				return output.PrintRawJSON(out, rawJSON)
			}

			return output.PrintObservationList(out, observations, columns)
		}

		// Without --trace-id, search across traces
		columns := obsColumns
		if len(columns) == 0 {
			columns = inputs.DefaultStandaloneObservationColumns
		}
		if !obsListJSON {
			if err := inputs.ValidateColumns(
				columns,
				inputs.AllStandaloneObservationColumns,
			); err != nil {
				return err
			}
		}

		fromTS := obsListFromTimestamp
		if fromTS == "" {
			fromTS = time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
		}

		opts := clients.ObservationSearchOptions{
			FromTimestamp:       fromTS,
			ToTimestamp:         obsListToTimestamp,
			Cursor:              obsListCursor,
			Limit:               obsListLimit,
			Fields:              obsListFields,
			Type:                obsListType,
			Name:                obsListName,
			Level:               obsListLevel,
			Model:               obsListModel,
			Environment:         obsListEnvironment,
			ParentObservationID: obsListParentObservationID,
			Version:             obsListVersion,
			UserID:              obsListUserID,
		}
		if err := inputs.ValidateObservationSearchOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), obsListOrg, obsListProject)
		if err != nil {
			return err
		}

		observations, meta, rawJSON, err := pCtx.apiClient.SearchObservations(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if obsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintStandaloneObservationList(out, observations, meta, columns)
	},
}

var obsGetCmd = &cobra.Command{
	Use:   "get <observation-id>",
	Short: "Get a specific observation",
	Long: `Get detailed information about a specific observation.

Uses the platform API with dual authentication (API key or session).

Examples:
  iai observations get obs-abc123
  iai observations get obs-abc123 --json | jq '.data.observation'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		observationID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(cmd.Context(), obsGetOrg, obsGetProject)
		if err != nil {
			return err
		}

		obs, rawJSON, err := pCtx.apiClient.GetObservation(
			cmd.Context(), pCtx.orgId, pCtx.projectId, observationID,
		)
		if err != nil {
			return err
		}

		if obsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintObservationDetail(out, obs)
	},
}

func init() {
	// observations list flags
	obsListCmd.Flags().
		StringVar(&obsTraceID, "trace-id", "", "Trace ID to list observations for (scopes to a single trace)")
	obsListCmd.Flags().
		BoolVar(&obsIncludeIO, "include-io", false, "Include input/output/metadata in response (only with --trace-id)")
	obsListCmd.Flags().BoolVar(&obsListJSON, "json", false, "Output raw API response as JSON")
	obsListCmd.Flags().
		StringSliceVar(&obsColumns, "columns", nil, "Columns to display (comma-separated)\nWith --trace-id default: id,type,name,model,latency_ms,total_cost,total_tokens\nWithout --trace-id default: id,trace_id,type,name,model,latency_ms,total_cost,total_tokens")
	obsListCmd.Flags().
		StringVarP(&obsListOrg, "organization", "o", "", "Organization name that owns the project")
	obsListCmd.Flags().
		StringVarP(&obsListProject, "project", "p", "", "Project name")

	// Search filters (used when --trace-id is not provided)
	obsListCmd.Flags().StringVar(
		&obsListFromTimestamp,
		"from-timestamp",
		"",
		"Filter observations from this timestamp (ISO 8601, default: 7 days ago)",
	)
	obsListCmd.Flags().
		StringVar(&obsListToTimestamp, "to-timestamp", "", "Filter observations to this timestamp (ISO 8601)")
	obsListCmd.Flags().StringVar(&obsListCursor, "cursor", "", "Cursor for pagination")
	obsListCmd.Flags().IntVar(&obsListLimit, "limit", 0, "Items per page")
	obsListCmd.Flags().
		StringVar(&obsListFields, "fields", "", "Field groups to include (comma-separated)")
	obsListCmd.Flags().StringVar(&obsListType, "type", "", "Filter by observation type")
	obsListCmd.Flags().StringVar(&obsListName, "name", "", "Filter by observation name")
	obsListCmd.Flags().StringVar(&obsListLevel, "level", "", "Filter by level")
	obsListCmd.Flags().StringVar(&obsListModel, "model", "", "Filter by model")
	obsListCmd.Flags().
		StringVar(&obsListEnvironment, "environment", "", "Filter by environment")
	obsListCmd.Flags().
		StringVar(&obsListParentObservationID, "parent-observation-id", "", "Filter by parent observation ID")
	obsListCmd.Flags().StringVar(&obsListVersion, "version", "", "Filter by version")
	obsListCmd.Flags().StringVar(&obsListUserID, "user-id", "", "Filter by user ID")

	// observations get flags
	obsGetCmd.Flags().BoolVar(&obsGetJSON, "json", false, "Output raw API response as JSON")
	obsGetCmd.Flags().
		StringVarP(&obsGetOrg, "organization", "o", "", "Organization name that owns the project")
	obsGetCmd.Flags().
		StringVarP(&obsGetProject, "project", "p", "", "Project name")

	observationsCmd.AddCommand(obsListCmd, obsGetCmd)
	rootCmd.AddCommand(observationsCmd)
}
