package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	queuesListPage    int
	queuesListLimit   int
	queuesListColumns []string
	queuesListJSON    bool
	queuesListYAML    bool
	queuesListOrg     string
	queuesListProject string

	queuesGetJSON    bool
	queuesGetYAML    bool
	queuesGetOrg     string
	queuesGetProject string

	queuesCreateDescription  string
	queuesCreateScoreConfigs []string
	queuesCreateJSON         bool
	queuesCreateYAML         bool
	queuesCreateOrg          string
	queuesCreateProject      string

	queuesAssignUserID  string
	queuesAssignOrg     string
	queuesAssignProject string

	queuesUnassignUserID  string
	queuesUnassignOrg     string
	queuesUnassignProject string
)

var queuesCmd = &cobra.Command{
	Use:              "queues",
	Aliases:          []string{"queue"},
	Short:            "Annotation queues for human review workflows",
	GroupID:          groupEvaluation,
	Long:             `Manage annotation queues for review workflows.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var queuesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List annotation queues",
	Long:    `List annotation queues with pagination.`,
	Example: `  iai queues list
  iai queues list --page 2 --limit 50
  iai queues list --columns id,name,description
  iai queues list -o my-org -p my-project --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := queuesListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultQueueColumns
		}
		if err := validateTableOnlyColumns(cmd, queuesListJSON, queuesListYAML); err != nil {
			return err
		}
		if !queuesListJSON && !queuesListYAML {
			if err := inputs.ValidateColumns(columns, inputs.AllQueueColumns); err != nil {
				return err
			}
		}

		opts := clients.AnnotationQueueListOptions{
			Page:  queuesListPage,
			Limit: queuesListLimit,
		}
		if err := inputs.ValidateQueueListOptions(opts); err != nil {
			return err
		}

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), queuesListOrg, queuesListProject)
		if err != nil {
			return err
		}

		queues, meta, rawJSON, err := apiClient.ListAnnotationQueues(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if queuesListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		if queuesListYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintQueueList(out, queues, meta, columns)
	},
}

var queuesGetCmd = &cobra.Command{
	Use:     "get <id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get an annotation queue",
	Long:    `Get detailed information about an annotation queue.`,
	Example: `  iai queues get queue-123
  iai queues get queue-123 --json
  iai queues get queue-123 -o my-org -p my-project --yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		queueID := strings.TrimSpace(args[0])

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), queuesGetOrg, queuesGetProject)
		if err != nil {
			return err
		}

		queue, rawJSON, err := apiClient.GetAnnotationQueue(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			queueID,
		)
		if err != nil {
			return err
		}

		if queuesGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		if queuesGetYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintQueueDetail(out, queue)
	},
}

var queuesCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an annotation queue",
	Long: `Create a new annotation queue.

This command requires API key authentication.`,
	Example: `  iai queues create my-queue
  iai queues create my-queue --description "Review of chat outputs"
  iai queues create my-queue --score-config-ids sc-1,sc-2
  iai queues create my-queue --description "QA queue" --score-config-ids sc-1 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		body := inputs.BuildQueueCreateBody(
			args[0],
			queuesCreateDescription,
			queuesCreateScoreConfigs,
		)

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			queuesCreateOrg,
			queuesCreateProject,
		)
		if err != nil {
			return err
		}

		queue, rawJSON, err := apiClient.CreateAnnotationQueue(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			body,
		)
		if err != nil {
			return err
		}

		if queuesCreateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		if queuesCreateYAML {
			return output.PrintRawYAML(out, rawJSON)
		}

		return output.PrintQueueCreateResult(out, queue)
	},
}

var queuesAssignCmd = &cobra.Command{
	Use:   "assign <queue-id>",
	Short: "Assign a user to a queue",
	Long: `Assign a user to an annotation queue.

This command requires API key authentication.`,
	Example: `  iai queues assign queue-123 --user-id user-456
  iai queues assign queue-123 --user-id user-456 -o my-org -p my-project`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		queueID := strings.TrimSpace(args[0])
		userID := strings.TrimSpace(queuesAssignUserID)

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			queuesAssignOrg,
			queuesAssignProject,
		)
		if err != nil {
			return err
		}

		message, err := apiClient.AssignQueue(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			queueID,
			userID,
		)
		if err != nil {
			return err
		}

		return output.PrintActionSuccess(out, message)
	},
}

var queuesUnassignCmd = &cobra.Command{
	Use:   "unassign <queue-id>",
	Short: "Unassign a user from a queue",
	Long: `Remove a user from an annotation queue.

This command requires API key authentication.`,
	Example: `  iai queues unassign queue-123 --user-id user-456
  iai queues unassign queue-123 --user-id user-456 -o my-org -p my-project`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		queueID := strings.TrimSpace(args[0])
		userID := strings.TrimSpace(queuesUnassignUserID)

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			queuesUnassignOrg,
			queuesUnassignProject,
		)
		if err != nil {
			return err
		}

		message, err := apiClient.UnassignQueue(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			queueID,
			userID,
		)
		if err != nil {
			return err
		}

		return output.PrintActionSuccess(out, message)
	},
}

func init() {
	queuesListCmd.Flags().IntVar(&queuesListPage, "page", 1, "Page number (starts at 1)")
	queuesListCmd.Flags().IntVar(&queuesListLimit, "limit", 0, "Items per page (max 100)")
	queuesListCmd.Flags().
		StringSliceVar(&queuesListColumns, "columns", nil, "Columns to display for table output only (comma-separated). Cannot be used with --json or --yaml")
	queuesListCmd.Flags().
		BoolVar(&queuesListJSON, "json", false, "Output raw API response as JSON")
	queuesListCmd.Flags().
		BoolVar(&queuesListYAML, "yaml", false, "Output raw API response as YAML")
	queuesListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	queuesListCmd.Flags().
		StringVarP(&queuesListOrg, "organization", "o", "", "Organization name that owns the project")
	queuesListCmd.Flags().
		StringVarP(&queuesListProject, "project", "p", "", "Project name")

	queuesGetCmd.Flags().
		BoolVar(&queuesGetJSON, "json", false, "Output raw API response as JSON")
	queuesGetCmd.Flags().
		BoolVar(&queuesGetYAML, "yaml", false, "Output raw API response as YAML")
	queuesGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	queuesGetCmd.Flags().
		StringVarP(&queuesGetOrg, "organization", "o", "", "Organization name that owns the project")
	queuesGetCmd.Flags().
		StringVarP(&queuesGetProject, "project", "p", "", "Project name")

	queuesCreateCmd.Flags().
		StringVar(&queuesCreateDescription, "description", "", "Queue description")
	queuesCreateCmd.Flags().
		StringSliceVar(
			&queuesCreateScoreConfigs, "score-config-ids", nil,
			"Score config IDs (comma-separated)",
		)
	queuesCreateCmd.Flags().
		BoolVar(&queuesCreateJSON, "json", false, "Output raw API response as JSON")
	queuesCreateCmd.Flags().
		BoolVar(&queuesCreateYAML, "yaml", false, "Output raw API response as YAML")
	queuesCreateCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	queuesCreateCmd.Flags().
		StringVarP(&queuesCreateOrg, "organization", "o", "", "Organization name that owns the project")
	queuesCreateCmd.Flags().
		StringVarP(&queuesCreateProject, "project", "p", "", "Project name")

	queuesAssignCmd.Flags().
		StringVar(&queuesAssignUserID, "user-id", "", "User ID to assign (required)")
	_ = queuesAssignCmd.MarkFlagRequired("user-id")
	queuesAssignCmd.Flags().
		StringVarP(&queuesAssignOrg, "organization", "o", "", "Organization name that owns the project")
	queuesAssignCmd.Flags().
		StringVarP(&queuesAssignProject, "project", "p", "", "Project name")

	queuesUnassignCmd.Flags().
		StringVar(&queuesUnassignUserID, "user-id", "", "User ID to unassign (required)")
	_ = queuesUnassignCmd.MarkFlagRequired("user-id")
	queuesUnassignCmd.Flags().
		StringVarP(&queuesUnassignOrg, "organization", "o", "", "Organization name that owns the project")
	queuesUnassignCmd.Flags().
		StringVarP(&queuesUnassignProject, "project", "p", "", "Project name")

	queuesCmd.AddCommand(
		queuesListCmd,
		queuesGetCmd,
		queuesCreateCmd,
		queuesAssignCmd,
		queuesUnassignCmd,
	)
	rootCmd.AddCommand(queuesCmd)
}
