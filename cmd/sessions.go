package cmd

import (
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
	"github.com/spf13/cobra"
)

var (
	sessionsFromTimestamp string
	sessionsToTimestamp   string
	sessionsPage          int
	sessionsLimit         int
	sessionsEnvironment   string
	sessionsColumns       []string
	sessionsListJSON      bool
	sessionsListYAML      bool
	sessionsGetFields     string
	sessionsGetJSON       bool
	sessionsGetYAML       bool
	sessionsListOrg       string
	sessionsListProject   string
	sessionsGetOrg        string
	sessionsGetProject    string
	sessionsGetSummary    bool
)

var sessionsCmd = &cobra.Command{
	Use:              "sessions",
	Aliases:          []string{"session"},
	Short:            "Browse trace-derived conversation sessions",
	GroupID:          groupObserve,
	Long:             `Manage trace-derived sessions. Works with API key (--api-key or INTERACTIVE_API_KEY) or session from 'iai login'.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var sessionsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List sessions",
	Long: `List sessions with optional filters.

Uses the platform API with dual authentication (API key or session).
If --from-timestamp is not provided, defaults to 7 days ago.`,
	Example: `  iai sessions list
  iai sessions list --from-timestamp 2026-06-01T00:00:00Z --to-timestamp 2026-06-08T00:00:00Z
  iai sessions list --environment production --limit 50 --page 2
  iai sessions list --columns id,created_at,total_cost --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := sessionsColumns
		if len(columns) == 0 {
			columns = inputs.DefaultSessionColumns
		}
		if err := validateTableOnlyColumns(cmd, sessionsListJSON, sessionsListYAML); err != nil {
			return err
		}
		if !sessionsListJSON && !sessionsListYAML {
			if err := inputs.ValidateColumns(columns, inputs.AllSessionColumns); err != nil {
				return err
			}
		}

		fromTS := sessionsFromTimestamp
		if fromTS == "" {
			fromTS = time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
		}

		opts := clients.SessionListOptions{
			FromTimestamp: fromTS,
			ToTimestamp:   sessionsToTimestamp,
			Page:          sessionsPage,
			Limit:         sessionsLimit,
			Environment:   sessionsEnvironment,
		}
		if err := inputs.ValidateSessionListOptions(opts); err != nil {
			return err
		}

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			sessionsListOrg,
			sessionsListProject,
		)
		if err != nil {
			return err
		}

		sessions, meta, rawJSON, err := apiClient.ListSessions(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if sessionsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		if sessionsListYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintSessionList(out, sessions, meta, columns)
	},
}

var sessionsGetCmd = &cobra.Command{
	Use:     "get <session-id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a specific session",
	Long: `Get detailed information about a specific session.

Uses the platform API with dual authentication (API key or session).`,
	Example: `  iai sessions get <session-id>
  iai sessions get <session-id> --fields core,traces
  iai sessions get <session-id> --json
  iai sessions get <session-id> --summary`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		sessionID := strings.TrimSpace(args[0])

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), sessionsGetOrg, sessionsGetProject)
		if err != nil {
			return err
		}

		if sessionsGetSummary {
			var all []clients.TraceInfo
			page := 1
			for {
				traces, meta, _, err := apiClient.ListTraces(
					cmd.Context(), pCtx.orgId, pCtx.projectId,
					clients.TraceListOptions{
						SessionID: sessionID,
						// The traces endpoint requires from_timestamp; the
						// session_id filter is the real scope, so use an
						// all-time lower bound to capture every turn.
						FromTimestamp: "1970-01-01T00:00:00Z",
						Fields:        "core,io",
						Order:         "asc",
						OrderBy:       "timestamp",
						Limit:         100,
						Page:          page,
					},
				)
				if err != nil {
					return err
				}
				all = append(all, traces...)
				if meta.TotalPages <= page || len(traces) == 0 {
					break
				}
				page++
			}
			return output.PrintSessionSummary(out, summary.SessionSummary(sessionID, all))
		}

		session, rawJSON, err := apiClient.GetSession(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			sessionID,
			sessionsGetFields,
		)
		if err != nil {
			return err
		}

		if sessionsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		if sessionsGetYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintSessionDetail(out, session)
	},
}

func init() {
	sessionsListCmd.Flags().StringVar(
		&sessionsFromTimestamp,
		"from-timestamp",
		"",
		"Filter sessions from this timestamp (ISO 8601, default: 7 days ago)",
	)
	sessionsListCmd.Flags().
		StringVar(&sessionsToTimestamp, "to-timestamp", "", "Filter sessions to this timestamp (ISO 8601)")
	sessionsListCmd.Flags().IntVar(&sessionsPage, "page", 1, "Page number (starts at 1)")
	sessionsListCmd.Flags().IntVar(&sessionsLimit, "limit", 0, "Items per page (max 100)")
	sessionsListCmd.Flags().
		StringVar(&sessionsEnvironment, "environment", "", "Filter by environment")
	sessionsListCmd.Flags().
		StringSliceVar(&sessionsColumns, "columns", nil, "Columns to display for table output only (comma-separated, default: id,created_at,environment,trace_count,duration_seconds,total_cost,total_tokens). Cannot be used with --json or --yaml.\nAvailable: id,created_at,updated_at,environment,user_id,trace_count,duration_seconds,total_cost,input_tokens,output_tokens,total_tokens")
	sessionsListCmd.Flags().
		BoolVar(&sessionsListJSON, "json", false, "Output raw API response as JSON")
	sessionsListCmd.Flags().
		BoolVar(&sessionsListYAML, "yaml", false, "Output raw API response as YAML")
	sessionsListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	sessionsListCmd.Flags().
		StringVarP(&sessionsListOrg, "organization", "o", "", "Organization name that owns the project")
	sessionsListCmd.Flags().
		StringVarP(&sessionsListProject, "project", "p", "", "Project name")

	sessionsGetCmd.Flags().
		StringVar(&sessionsGetFields, "fields", "core,traces", "Field groups to include (comma-separated)")
	sessionsGetCmd.Flags().
		BoolVar(&sessionsGetJSON, "json", false, "Output raw API response as JSON")
	sessionsGetCmd.Flags().
		BoolVar(&sessionsGetYAML, "yaml", false, "Output raw API response as YAML")
	sessionsGetCmd.Flags().BoolVar(&sessionsGetSummary, "summary", false,
		"Render a compact, LLM-readable overview of the conversation (transcript + event tags)")
	sessionsGetCmd.MarkFlagsMutuallyExclusive("summary", "json", "yaml", "fields")
	sessionsGetCmd.Flags().
		StringVarP(&sessionsGetOrg, "organization", "o", "", "Organization name that owns the project")
	sessionsGetCmd.Flags().
		StringVarP(&sessionsGetProject, "project", "p", "", "Project name")

	sessionsCmd.AddCommand(sessionsListCmd, sessionsGetCmd)
	rootCmd.AddCommand(sessionsCmd)
}
