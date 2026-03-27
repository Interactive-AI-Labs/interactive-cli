package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	queueItemsListQueueID string
	queueItemsListStatus  string
	queueItemsListPage    int
	queueItemsListLimit   int
	queueItemsListColumns []string
	queueItemsListJSON    bool
	queueItemsListOrg     string
	queueItemsListProject string

	queueItemsGetQueueID string
	queueItemsGetJSON    bool
	queueItemsGetOrg     string
	queueItemsGetProject string

	queueItemsCreateQueueID    string
	queueItemsCreateObjectID   string
	queueItemsCreateObjectType string
	queueItemsCreateStatus     string
	queueItemsCreateJSON       bool
	queueItemsCreateOrg        string
	queueItemsCreateProject    string

	queueItemsUpdateQueueID string
	queueItemsUpdateStatus  string
	queueItemsUpdateJSON    bool
	queueItemsUpdateOrg     string
	queueItemsUpdateProject string

	queueItemsDeleteQueueID string
	queueItemsDeleteOrg     string
	queueItemsDeleteProject string
)

var queueItemsCmd = &cobra.Command{
	Use:              "queue-items",
	Aliases:          []string{"queue-item"},
	Short:            "Manage annotation queue items",
	Long:             `Manage items within annotation queues.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var queueItemsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List queue items",
	Long:    `List items in an annotation queue.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		queueID := strings.TrimSpace(queueItemsListQueueID)

		columns := queueItemsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultQueueItemColumns
		}
		if !queueItemsListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllQueueItemColumns); err != nil {
				return err
			}
		}

		opts := clients.QueueItemListOptions{
			Status: queueItemsListStatus,
			Page:   queueItemsListPage,
			Limit:  queueItemsListLimit,
		}
		if err := inputs.ValidateQueueItemListOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(
			cmd.Context(),
			queueItemsListOrg,
			queueItemsListProject,
		)
		if err != nil {
			return err
		}

		items, meta, rawJSON, err := pCtx.apiClient.ListQueueItems(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			queueID,
			opts,
		)
		if err != nil {
			return err
		}

		if queueItemsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintQueueItemList(out, items, meta, columns)
	},
}

var queueItemsGetCmd = &cobra.Command{
	Use:     "get <item-id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a queue item",
	Long:    `Get detailed information about a queue item.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		queueID := strings.TrimSpace(queueItemsGetQueueID)
		itemID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(cmd.Context(), queueItemsGetOrg, queueItemsGetProject)
		if err != nil {
			return err
		}

		item, rawJSON, err := pCtx.apiClient.GetQueueItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			queueID,
			itemID,
		)
		if err != nil {
			return err
		}

		if queueItemsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintQueueItemDetail(out, item)
	},
}

var queueItemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a queue item",
	Long: `Add an item to an annotation queue.

This command requires API key authentication.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		queueID := strings.TrimSpace(queueItemsCreateQueueID)

		body := inputs.BuildQueueItemCreateBody(
			queueItemsCreateObjectID,
			queueItemsCreateObjectType,
			queueItemsCreateStatus,
		)

		pCtx, err := resolveProject(
			cmd.Context(),
			queueItemsCreateOrg,
			queueItemsCreateProject,
		)
		if err != nil {
			return err
		}

		item, rawJSON, err := pCtx.apiClient.CreateQueueItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			queueID,
			body,
		)
		if err != nil {
			return err
		}

		if queueItemsCreateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintQueueItemCreateResult(out, item)
	},
}

var queueItemsUpdateCmd = &cobra.Command{
	Use:   "update <item-id>",
	Short: "Update a queue item",
	Long: `Update the status of a queue item.

This command requires API key authentication.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		queueID := strings.TrimSpace(queueItemsUpdateQueueID)
		status := strings.TrimSpace(queueItemsUpdateStatus)
		itemID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(
			cmd.Context(),
			queueItemsUpdateOrg,
			queueItemsUpdateProject,
		)
		if err != nil {
			return err
		}

		item, rawJSON, err := pCtx.apiClient.UpdateQueueItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			queueID,
			itemID,
			clients.QueueItemUpdateBody{Status: status},
		)
		if err != nil {
			return err
		}

		if queueItemsUpdateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintQueueItemUpdateResult(out, item)
	},
}

var queueItemsDeleteCmd = &cobra.Command{
	Use:     "delete <item-id>",
	Aliases: []string{"rm"},
	Short:   "Delete a queue item",
	Long:    `Delete an item from an annotation queue.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		queueID := strings.TrimSpace(queueItemsDeleteQueueID)
		itemID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(
			cmd.Context(),
			queueItemsDeleteOrg,
			queueItemsDeleteProject,
		)
		if err != nil {
			return err
		}

		message, err := pCtx.apiClient.DeleteQueueItem(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			queueID,
			itemID,
		)
		if err != nil {
			return err
		}

		return output.PrintDeleteSuccess(out, itemID, "queue item", message)
	},
}

func init() {
	queueItemsListCmd.Flags().
		StringVar(&queueItemsListQueueID, "queue-id", "", "Queue ID (required)")
	_ = queueItemsListCmd.MarkFlagRequired("queue-id")
	queueItemsListCmd.Flags().
		StringVar(&queueItemsListStatus, "status", "", "Filter by status (PENDING/COMPLETED)")
	queueItemsListCmd.Flags().
		IntVar(&queueItemsListPage, "page", 1, "Page number (starts at 1)")
	queueItemsListCmd.Flags().IntVar(&queueItemsListLimit, "limit", 50, "Items per page")
	queueItemsListCmd.Flags().
		StringSliceVar(&queueItemsListColumns, "columns", nil, "Columns to display (comma-separated)")
	queueItemsListCmd.Flags().
		BoolVar(&queueItemsListJSON, "json", false, "Output raw API response as JSON")
	queueItemsListCmd.Flags().
		StringVarP(&queueItemsListOrg, "organization", "o", "", "Organization name that owns the project")
	queueItemsListCmd.Flags().
		StringVarP(&queueItemsListProject, "project", "p", "", "Project name")

	queueItemsGetCmd.Flags().
		StringVar(&queueItemsGetQueueID, "queue-id", "", "Queue ID (required)")
	_ = queueItemsGetCmd.MarkFlagRequired("queue-id")
	queueItemsGetCmd.Flags().
		BoolVar(&queueItemsGetJSON, "json", false, "Output raw API response as JSON")
	queueItemsGetCmd.Flags().
		StringVarP(&queueItemsGetOrg, "organization", "o", "", "Organization name that owns the project")
	queueItemsGetCmd.Flags().
		StringVarP(&queueItemsGetProject, "project", "p", "", "Project name")

	queueItemsCreateCmd.Flags().
		StringVar(&queueItemsCreateQueueID, "queue-id", "", "Queue ID (required)")
	queueItemsCreateCmd.Flags().
		StringVar(&queueItemsCreateObjectID, "object-id", "", "Object ID (required)")
	queueItemsCreateCmd.Flags().
		StringVar(
			&queueItemsCreateObjectType, "object-type", "",
			"Object type: TRACE or OBSERVATION (required)",
		)
	_ = queueItemsCreateCmd.MarkFlagRequired("queue-id")
	_ = queueItemsCreateCmd.MarkFlagRequired("object-id")
	_ = queueItemsCreateCmd.MarkFlagRequired("object-type")
	queueItemsCreateCmd.Flags().
		StringVar(&queueItemsCreateStatus, "status", "", "Status (PENDING/COMPLETED)")
	queueItemsCreateCmd.Flags().
		BoolVar(&queueItemsCreateJSON, "json", false, "Output raw API response as JSON")
	queueItemsCreateCmd.Flags().
		StringVarP(&queueItemsCreateOrg, "organization", "o", "", "Organization name that owns the project")
	queueItemsCreateCmd.Flags().
		StringVarP(&queueItemsCreateProject, "project", "p", "", "Project name")

	queueItemsUpdateCmd.Flags().
		StringVar(&queueItemsUpdateQueueID, "queue-id", "", "Queue ID (required)")
	queueItemsUpdateCmd.Flags().
		StringVar(&queueItemsUpdateStatus, "status", "", "New status (required)")
	_ = queueItemsUpdateCmd.MarkFlagRequired("queue-id")
	_ = queueItemsUpdateCmd.MarkFlagRequired("status")
	queueItemsUpdateCmd.Flags().
		BoolVar(&queueItemsUpdateJSON, "json", false, "Output raw API response as JSON")
	queueItemsUpdateCmd.Flags().
		StringVarP(&queueItemsUpdateOrg, "organization", "o", "", "Organization name that owns the project")
	queueItemsUpdateCmd.Flags().
		StringVarP(&queueItemsUpdateProject, "project", "p", "", "Project name")

	queueItemsDeleteCmd.Flags().
		StringVar(&queueItemsDeleteQueueID, "queue-id", "", "Queue ID (required)")
	_ = queueItemsDeleteCmd.MarkFlagRequired("queue-id")
	queueItemsDeleteCmd.Flags().
		StringVarP(&queueItemsDeleteOrg, "organization", "o", "", "Organization name that owns the project")
	queueItemsDeleteCmd.Flags().
		StringVarP(&queueItemsDeleteProject, "project", "p", "", "Project name")

	queueItemsCmd.AddCommand(
		queueItemsListCmd,
		queueItemsGetCmd,
		queueItemsCreateCmd,
		queueItemsUpdateCmd,
		queueItemsDeleteCmd,
	)
	rootCmd.AddCommand(queueItemsCmd)
}
