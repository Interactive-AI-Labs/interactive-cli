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
	scoresFromTimestamp string
	scoresToTimestamp   string
	scoresCursor        string
	scoresLimit         int
	scoresFields        string
	scoresName          string
	scoresTraceID       string
	scoresObservationID string
	scoresSessionID     string
	scoresSource        string
	scoresDataType      string
	scoresEnvironment   string
	scoresConfigID      string
	scoresMinValue      float64
	scoresMaxValue      float64
	scoresScoreIDs      []string
	scoresUserID        string
	scoresTraceTags     []string
	scoresValue         string
	scoresOperator      string
	scoresColumns       []string
	scoresListJSON      bool
	scoresListOrg       string
	scoresListProject   string

	scoreCreateID           string
	scoreCreateName         string
	scoreCreateTraceID      string
	scoreCreateObservation  string
	scoreCreateSessionID    string
	scoreCreateValue        string
	scoreCreateDataType     string
	scoreCreateComment      string
	scoreCreateMetadataJSON string
	scoreCreateEnvironment  string
	scoreCreateConfigID     string
	scoreCreateQueueID      string
	scoreCreateOrg          string
	scoreCreateProject      string

	scoreDeleteOrg     string
	scoreDeleteProject string
)

var scoresCmd = &cobra.Command{
	Use:              "scores",
	Aliases:          []string{"score"},
	Short:            "Manage scores",
	Long:             `Manage observability scores. Read commands work with API key or session login; write commands currently require API key authentication.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var scoresListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List scores",
	Long: `List scores with optional filters.

Uses the platform API with dual authentication (API key or session).
If --from-timestamp is not provided, defaults to 7 days ago.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := scoresColumns
		if len(columns) == 0 {
			columns = inputs.DefaultScoreColumns
		}
		if !scoresListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllScoreColumns); err != nil {
				return err
			}
		}

		fromTS := scoresFromTimestamp
		if fromTS == "" {
			fromTS = time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
		}

		opts := clients.ScoreListOptions{
			FromTimestamp: fromTS,
			ToTimestamp:   scoresToTimestamp,
			Cursor:        scoresCursor,
			Limit:         scoresLimit,
			Fields:        scoresFields,
			Name:          scoresName,
			TraceID:       scoresTraceID,
			ObservationID: scoresObservationID,
			SessionID:     scoresSessionID,
			Source:        scoresSource,
			DataType:      scoresDataType,
			Environment:   scoresEnvironment,
			ConfigID:      scoresConfigID,
			ScoreIDs:      scoresScoreIDs,
			UserID:        scoresUserID,
			TraceTags:     scoresTraceTags,
			Value:         scoresValue,
			Operator:      scoresOperator,
		}
		if cmd.Flags().Changed("min-value") {
			opts.MinValue = &scoresMinValue
		}
		if cmd.Flags().Changed("max-value") {
			opts.MaxValue = &scoresMaxValue
		}

		opts, err := inputs.PrepareScoreListOptions(opts)
		if err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), scoresListOrg, scoresListProject)
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

		scores, meta, rawJSON, err := apiClient.ListScores(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if scoresListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintScoreList(out, scores, meta, columns)
	},
}

var scoresCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a score",
	Long: `Create a score on exactly one target resource.

This command currently requires API key authentication.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		body, err := inputs.BuildScoreCreateBody(inputs.ScoreCreateInput{
			ID:            scoreCreateID,
			Name:          scoreCreateName,
			TraceID:       scoreCreateTraceID,
			ObservationID: scoreCreateObservation,
			SessionID:     scoreCreateSessionID,
			DataType:      scoreCreateDataType,
			Value:         scoreCreateValue,
			Comment:       scoreCreateComment,
			MetadataJSON:  scoreCreateMetadataJSON,
			Environment:   scoreCreateEnvironment,
			ConfigID:      scoreCreateConfigID,
			QueueID:       scoreCreateQueueID,
		})
		if err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), scoreCreateOrg, scoreCreateProject)
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

		score, err := apiClient.CreateScore(cmd.Context(), pCtx.orgId, pCtx.projectId, body)
		if err != nil {
			return err
		}

		return output.PrintScoreCreateResult(out, score)
	},
}

var scoresDeleteCmd = &cobra.Command{
	Use:     "delete <score-id>",
	Aliases: []string{"rm"},
	Short:   "Delete a score",
	Long: `Delete a score by ID.

This command currently requires API key authentication.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		scoreID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(cmd.Context(), scoreDeleteOrg, scoreDeleteProject)
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

		message, err := apiClient.DeleteScore(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			scoreID,
		)
		if err != nil {
			return err
		}

		return output.PrintDeleteSuccess(out, scoreID, "score", message)
	},
}

func init() {
	scoresListCmd.Flags().StringVar(
		&scoresFromTimestamp,
		"from-timestamp",
		"",
		"Filter scores from this timestamp (ISO 8601, default: 7 days ago)",
	)
	scoresListCmd.Flags().
		StringVar(&scoresToTimestamp, "to-timestamp", "", "Filter scores to this timestamp (ISO 8601)")
	scoresListCmd.Flags().StringVar(&scoresCursor, "cursor", "", "Cursor for pagination")
	scoresListCmd.Flags().IntVar(&scoresLimit, "limit", 0, "Items per page")
	scoresListCmd.Flags().
		StringVar(&scoresFields, "fields", "", "Field groups to include (comma-separated)")
	scoresListCmd.Flags().StringVar(&scoresName, "name", "", "Filter by score name")
	scoresListCmd.Flags().StringVar(&scoresTraceID, "trace-id", "", "Filter by trace ID")
	scoresListCmd.Flags().
		StringVar(&scoresObservationID, "observation-id", "", "Filter by observation ID")
	scoresListCmd.Flags().
		StringVar(&scoresSessionID, "session-id", "", "Filter by session ID")
	scoresListCmd.Flags().StringVar(&scoresSource, "source", "", "Filter by source")
	scoresListCmd.Flags().StringVar(&scoresDataType, "data-type", "", "Filter by data type")
	scoresListCmd.Flags().
		StringVar(&scoresEnvironment, "environment", "", "Filter by environment")
	scoresListCmd.Flags().StringVar(&scoresConfigID, "config-id", "", "Filter by config ID")
	scoresListCmd.Flags().Float64Var(&scoresMinValue, "min-value", 0, "Minimum score value")
	scoresListCmd.Flags().Float64Var(&scoresMaxValue, "max-value", 0, "Maximum score value")
	scoresListCmd.Flags().
		StringArrayVar(&scoresScoreIDs, "score-id", nil, "Filter by score ID (repeatable)")
	scoresListCmd.Flags().StringVar(&scoresUserID, "user-id", "", "Filter by user ID")
	scoresListCmd.Flags().
		StringArrayVar(&scoresTraceTags, "trace-tag", nil, "Filter by trace tag (repeatable)")
	scoresListCmd.Flags().StringVar(&scoresValue, "value", "", "Exact value filter")
	scoresListCmd.Flags().StringVar(&scoresOperator, "operator", "", "Operator for --value")
	scoresListCmd.Flags().
		StringSliceVar(&scoresColumns, "columns", nil, "Columns to display (comma-separated, default: id,name,data_type,value,source,timestamp,trace_id)\nAvailable: id,name,data_type,value,source,timestamp,trace_id,observation_id,session_id,environment,config_id,user_id,comment")
	scoresListCmd.Flags().
		BoolVar(&scoresListJSON, "json", false, "Output raw API response as JSON")
	scoresListCmd.Flags().
		StringVarP(&scoresListOrg, "organization", "o", "", "Organization name that owns the project")
	scoresListCmd.Flags().
		StringVarP(&scoresListProject, "project", "p", "", "Project name")

	scoresCreateCmd.Flags().StringVar(&scoreCreateID, "id", "", "Explicit score ID")
	scoresCreateCmd.Flags().StringVar(&scoreCreateName, "name", "", "Score name")
	scoresCreateCmd.Flags().StringVar(&scoreCreateTraceID, "trace-id", "", "Target trace ID")
	scoresCreateCmd.Flags().
		StringVar(&scoreCreateObservation, "observation-id", "", "Target observation ID")
	scoresCreateCmd.Flags().
		StringVar(&scoreCreateSessionID, "session-id", "", "Target session ID")
	scoresCreateCmd.Flags().StringVar(&scoreCreateValue, "value", "", "Score value")
	scoresCreateCmd.Flags().
		StringVar(&scoreCreateDataType, "data-type", "NUMERIC", "Score data type")
	scoresCreateCmd.Flags().StringVar(&scoreCreateComment, "comment", "", "Score comment")
	scoresCreateCmd.Flags().
		StringVar(&scoreCreateMetadataJSON, "metadata-json", "", "Metadata as JSON object")
	scoresCreateCmd.Flags().
		StringVar(&scoreCreateEnvironment, "environment", "", "Target environment")
	scoresCreateCmd.Flags().StringVar(&scoreCreateConfigID, "config-id", "", "Related config ID")
	scoresCreateCmd.Flags().StringVar(&scoreCreateQueueID, "queue-id", "", "Related queue ID")
	scoresCreateCmd.Flags().
		StringVarP(&scoreCreateOrg, "organization", "o", "", "Organization name that owns the project")
	scoresCreateCmd.Flags().
		StringVarP(&scoreCreateProject, "project", "p", "", "Project name")
	_ = scoresCreateCmd.MarkFlagRequired("name")
	_ = scoresCreateCmd.MarkFlagRequired("value")

	scoresDeleteCmd.Flags().
		StringVarP(&scoreDeleteOrg, "organization", "o", "", "Organization name that owns the project")
	scoresDeleteCmd.Flags().
		StringVarP(&scoreDeleteProject, "project", "p", "", "Project name")

	scoresCmd.AddCommand(scoresListCmd, scoresCreateCmd, scoresDeleteCmd)
	rootCmd.AddCommand(scoresCmd)
}
