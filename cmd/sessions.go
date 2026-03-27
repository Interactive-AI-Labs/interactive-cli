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
	sessionsFromTimestamp string
	sessionsToTimestamp   string
	sessionsPage          int
	sessionsLimit         int
	sessionsEnvironment   string
	sessionsColumns       []string
	sessionsListJSON      bool
	sessionsGetFields     string
	sessionsGetJSON       bool
	sessionsListOrg       string
	sessionsListProject   string
	sessionsGetOrg        string
	sessionsGetProject    string
)

var sessionsCmd = &cobra.Command{
	Use:              "sessions",
	Aliases:          []string{"session"},
	Short:            "Manage sessions",
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := sessionsColumns
		if len(columns) == 0 {
			columns = inputs.DefaultSessionColumns
		}
		if !sessionsListJSON {
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

		pCtx, err := resolveProject(cmd.Context(), sessionsListOrg, sessionsListProject)
		if err != nil {
			return err
		}

		sessions, meta, rawJSON, err := pCtx.apiClient.ListSessions(
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

		return output.PrintSessionList(out, sessions, meta, columns)
	},
}

var sessionsGetCmd = &cobra.Command{
	Use:     "get <session-id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a specific session",
	Long: `Get detailed information about a specific session.

Uses the platform API with dual authentication (API key or session).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		sessionID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(cmd.Context(), sessionsGetOrg, sessionsGetProject)
		if err != nil {
			return err
		}

		session, rawJSON, err := pCtx.apiClient.GetSession(
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
	sessionsListCmd.Flags().IntVar(&sessionsLimit, "limit", 0, "Items per page")
	sessionsListCmd.Flags().
		StringVar(&sessionsEnvironment, "environment", "", "Filter by environment")
	sessionsListCmd.Flags().
		StringSliceVar(&sessionsColumns, "columns", nil, "Columns to display (comma-separated, default: id,created_at,environment,trace_count,duration_seconds,total_cost,total_tokens)\nAvailable: id,created_at,updated_at,environment,user_id,trace_count,duration_seconds,total_cost,input_tokens,output_tokens,total_tokens")
	sessionsListCmd.Flags().
		BoolVar(&sessionsListJSON, "json", false, "Output raw API response as JSON")
	sessionsListCmd.Flags().
		StringVarP(&sessionsListOrg, "organization", "o", "", "Organization name that owns the project")
	sessionsListCmd.Flags().
		StringVarP(&sessionsListProject, "project", "p", "", "Project name")

	sessionsGetCmd.Flags().
		StringVar(&sessionsGetFields, "fields", "core,traces", "Field groups to include (comma-separated)")
	sessionsGetCmd.Flags().
		BoolVar(&sessionsGetJSON, "json", false, "Output raw API response as JSON")
	sessionsGetCmd.Flags().
		StringVarP(&sessionsGetOrg, "organization", "o", "", "Organization name that owns the project")
	sessionsGetCmd.Flags().
		StringVarP(&sessionsGetProject, "project", "p", "", "Project name")

	sessionsCmd.AddCommand(sessionsListCmd, sessionsGetCmd)
	rootCmd.AddCommand(sessionsCmd)
}
