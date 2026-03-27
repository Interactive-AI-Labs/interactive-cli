package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	commentsListObjectType   string
	commentsListObjectID     string
	commentsListAuthorUserID string
	commentsListPage         int
	commentsListLimit        int
	commentsListColumns      []string
	commentsListJSON         bool
	commentsListOrg          string
	commentsListProject      string

	commentsGetJSON    bool
	commentsGetOrg     string
	commentsGetProject string

	commentsCreateObjectType   string
	commentsCreateObjectID     string
	commentsCreateContent      string
	commentsCreateAuthorUserID string
	commentsCreateJSON         bool
	commentsCreateOrg          string
	commentsCreateProject      string
)

var commentsCmd = &cobra.Command{
	Use:              "comments",
	Aliases:          []string{"comment"},
	Short:            "Manage comments",
	Long:             `Manage comments on traces, observations, sessions, and prompts.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var commentsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List comments",
	Long:    `List comments with optional filters.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := commentsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultCommentColumns
		}
		if !commentsListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllCommentColumns); err != nil {
				return err
			}
		}

		opts := clients.CommentListOptions{
			ObjectType:   commentsListObjectType,
			ObjectID:     commentsListObjectID,
			AuthorUserID: commentsListAuthorUserID,
			Page:         commentsListPage,
			Limit:        commentsListLimit,
		}
		if err := inputs.ValidateCommentListOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(cmd.Context(), commentsListOrg, commentsListProject)
		if err != nil {
			return err
		}

		comments, meta, rawJSON, err := pCtx.apiClient.ListComments(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if commentsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintCommentList(out, comments, meta, columns)
	},
}

var commentsGetCmd = &cobra.Command{
	Use:     "get <id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a comment",
	Long:    `Get full details of a comment.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		commentID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(cmd.Context(), commentsGetOrg, commentsGetProject)
		if err != nil {
			return err
		}

		comment, rawJSON, err := pCtx.apiClient.GetComment(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			commentID,
		)
		if err != nil {
			return err
		}

		if commentsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintCommentDetail(out, comment)
	},
}

var commentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a comment",
	Long: `Create a new comment on a trace, observation, session, or prompt.

This command requires API key authentication.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		body := inputs.BuildCommentCreateBody(
			commentsCreateObjectType,
			commentsCreateObjectID,
			commentsCreateContent,
			commentsCreateAuthorUserID,
		)

		pCtx, err := resolveProject(cmd.Context(), commentsCreateOrg, commentsCreateProject)
		if err != nil {
			return err
		}

		comment, rawJSON, err := pCtx.apiClient.CreateComment(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			body,
		)
		if err != nil {
			return err
		}

		if commentsCreateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintCommentCreateResult(out, comment)
	},
}

func init() {
	commentsListCmd.Flags().
		StringVar(
			&commentsListObjectType, "object-type", "",
			"Filter by object type (TRACE/OBSERVATION/SESSION/PROMPT)",
		)
	commentsListCmd.Flags().
		StringVar(&commentsListObjectID, "object-id", "", "Filter by object ID")
	commentsListCmd.Flags().
		StringVar(&commentsListAuthorUserID, "author-user-id", "", "Filter by author user ID")
	commentsListCmd.Flags().IntVar(&commentsListPage, "page", 1, "Page number (starts at 1)")
	commentsListCmd.Flags().IntVar(&commentsListLimit, "limit", 0, "Items per page")
	commentsListCmd.Flags().
		StringSliceVar(&commentsListColumns, "columns", nil, "Columns to display (comma-separated)")
	commentsListCmd.Flags().
		BoolVar(&commentsListJSON, "json", false, "Output raw API response as JSON")
	commentsListCmd.Flags().
		StringVarP(&commentsListOrg, "organization", "o", "", "Organization name that owns the project")
	commentsListCmd.Flags().
		StringVarP(&commentsListProject, "project", "p", "", "Project name")

	commentsGetCmd.Flags().
		BoolVar(&commentsGetJSON, "json", false, "Output raw API response as JSON")
	commentsGetCmd.Flags().
		StringVarP(&commentsGetOrg, "organization", "o", "", "Organization name that owns the project")
	commentsGetCmd.Flags().
		StringVarP(&commentsGetProject, "project", "p", "", "Project name")

	commentsCreateCmd.Flags().
		StringVar(
			&commentsCreateObjectType, "object-type", "",
			"Object type: TRACE, OBSERVATION, SESSION, or PROMPT (required)",
		)
	commentsCreateCmd.Flags().
		StringVar(&commentsCreateObjectID, "object-id", "", "Object ID (required)")
	commentsCreateCmd.Flags().
		StringVar(&commentsCreateContent, "content", "", "Comment content (required)")
	_ = commentsCreateCmd.MarkFlagRequired("object-type")
	_ = commentsCreateCmd.MarkFlagRequired("object-id")
	_ = commentsCreateCmd.MarkFlagRequired("content")
	commentsCreateCmd.Flags().
		StringVar(&commentsCreateAuthorUserID, "author-user-id", "", "Author user ID")
	commentsCreateCmd.Flags().
		BoolVar(&commentsCreateJSON, "json", false, "Output raw API response as JSON")
	commentsCreateCmd.Flags().
		StringVarP(&commentsCreateOrg, "organization", "o", "", "Organization name that owns the project")
	commentsCreateCmd.Flags().
		StringVarP(&commentsCreateProject, "project", "p", "", "Project name")

	commentsCmd.AddCommand(commentsListCmd, commentsGetCmd, commentsCreateCmd)
	rootCmd.AddCommand(commentsCmd)
}
