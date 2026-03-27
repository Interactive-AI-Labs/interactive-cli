package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	scoreConfigsListPage    int
	scoreConfigsListLimit   int
	scoreConfigsListColumns []string
	scoreConfigsListJSON    bool
	scoreConfigsListOrg     string
	scoreConfigsListProject string

	scoreConfigsGetJSON    bool
	scoreConfigsGetOrg     string
	scoreConfigsGetProject string

	scoreConfigsCreateName        string
	scoreConfigsCreateDataType    string
	scoreConfigsCreateMinValue    float64
	scoreConfigsCreateMaxValue    float64
	scoreConfigsCreateCategories  string
	scoreConfigsCreateDescription string
	scoreConfigsCreateJSON        bool
	scoreConfigsCreateOrg         string
	scoreConfigsCreateProject     string

	scoreConfigsUpdateDescription string
	scoreConfigsUpdateIsArchived  bool
	scoreConfigsUpdateMinValue    float64
	scoreConfigsUpdateMaxValue    float64
	scoreConfigsUpdateCategories  string
	scoreConfigsUpdateJSON        bool
	scoreConfigsUpdateOrg         string
	scoreConfigsUpdateProject     string
)

var scoreConfigsCmd = &cobra.Command{
	Use:              "score-configs",
	Aliases:          []string{"score-config"},
	Short:            "Manage score configs",
	Long:             `Manage scoring configuration schemas for annotation workflows.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var scoreConfigsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List score configs",
	Long:    `List scoring configurations with pagination.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		columns := scoreConfigsListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultScoreConfigColumns
		}
		if !scoreConfigsListJSON {
			if err := inputs.ValidateColumns(columns, inputs.AllScoreConfigColumns); err != nil {
				return err
			}
		}

		opts := clients.ScoreConfigListOptions{
			Page:  scoreConfigsListPage,
			Limit: scoreConfigsListLimit,
		}
		if err := inputs.ValidateScoreConfigListOptions(opts); err != nil {
			return err
		}

		pCtx, err := resolveProject(
			cmd.Context(),
			scoreConfigsListOrg,
			scoreConfigsListProject,
		)
		if err != nil {
			return err
		}

		configs, meta, rawJSON, err := pCtx.apiClient.ListScoreConfigs(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if scoreConfigsListJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintScoreConfigList(out, configs, meta, columns)
	},
}

var scoreConfigsGetCmd = &cobra.Command{
	Use:     "get <id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a score config",
	Long:    `Get detailed information about a score configuration.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		configID := strings.TrimSpace(args[0])

		pCtx, err := resolveProject(
			cmd.Context(),
			scoreConfigsGetOrg,
			scoreConfigsGetProject,
		)
		if err != nil {
			return err
		}

		config, rawJSON, err := pCtx.apiClient.GetScoreConfig(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			configID,
		)
		if err != nil {
			return err
		}

		if scoreConfigsGetJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintScoreConfigDetail(out, config)
	},
}

var scoreConfigsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a score config",
	Long: `Create a new scoring configuration.

This command requires API key authentication.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		createInput := inputs.ScoreConfigCreateInput{
			Name:           scoreConfigsCreateName,
			DataType:       scoreConfigsCreateDataType,
			CategoriesJSON: scoreConfigsCreateCategories,
			Description:    scoreConfigsCreateDescription,
		}
		if cmd.Flags().Changed("min-value") {
			createInput.MinValue = &scoreConfigsCreateMinValue
		}
		if cmd.Flags().Changed("max-value") {
			createInput.MaxValue = &scoreConfigsCreateMaxValue
		}

		body, err := inputs.BuildScoreConfigCreateBody(createInput)
		if err != nil {
			return err
		}

		pCtx, err := resolveProject(
			cmd.Context(),
			scoreConfigsCreateOrg,
			scoreConfigsCreateProject,
		)
		if err != nil {
			return err
		}

		config, rawJSON, err := pCtx.apiClient.CreateScoreConfig(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			body,
		)
		if err != nil {
			return err
		}

		if scoreConfigsCreateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintScoreConfigCreateResult(out, config)
	},
}

var scoreConfigsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a score config",
	Long: `Update an existing scoring configuration.

This command requires API key authentication.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		configID := strings.TrimSpace(args[0])

		if !cmd.Flags().Changed("description") &&
			!cmd.Flags().Changed("is-archived") &&
			!cmd.Flags().Changed("min-value") &&
			!cmd.Flags().Changed("max-value") &&
			!cmd.Flags().Changed("categories") {
			return fmt.Errorf("at least one flag must be provided to update")
		}

		updateInput := inputs.ScoreConfigUpdateInput{
			CategoriesJSON: scoreConfigsUpdateCategories,
		}
		if cmd.Flags().Changed("description") {
			updateInput.Description = &scoreConfigsUpdateDescription
		}
		if cmd.Flags().Changed("is-archived") {
			updateInput.IsArchived = &scoreConfigsUpdateIsArchived
		}
		if cmd.Flags().Changed("min-value") {
			updateInput.MinValue = &scoreConfigsUpdateMinValue
		}
		if cmd.Flags().Changed("max-value") {
			updateInput.MaxValue = &scoreConfigsUpdateMaxValue
		}

		body, err := inputs.BuildScoreConfigUpdateBody(updateInput)
		if err != nil {
			return err
		}

		pCtx, err := resolveProject(
			cmd.Context(),
			scoreConfigsUpdateOrg,
			scoreConfigsUpdateProject,
		)
		if err != nil {
			return err
		}

		config, rawJSON, err := pCtx.apiClient.UpdateScoreConfig(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			configID,
			body,
		)
		if err != nil {
			return err
		}

		if scoreConfigsUpdateJSON {
			return output.PrintRawJSON(out, rawJSON)
		}

		return output.PrintScoreConfigUpdateResult(out, config)
	},
}

func init() {
	scoreConfigsListCmd.Flags().
		IntVar(&scoreConfigsListPage, "page", 1, "Page number (starts at 1)")
	scoreConfigsListCmd.Flags().
		IntVar(&scoreConfigsListLimit, "limit", 50, "Items per page")
	scoreConfigsListCmd.Flags().
		StringSliceVar(&scoreConfigsListColumns, "columns", nil, "Columns to display (comma-separated)")
	scoreConfigsListCmd.Flags().
		BoolVar(&scoreConfigsListJSON, "json", false, "Output raw API response as JSON")
	scoreConfigsListCmd.Flags().
		StringVarP(&scoreConfigsListOrg, "organization", "o", "", "Organization name that owns the project")
	scoreConfigsListCmd.Flags().
		StringVarP(&scoreConfigsListProject, "project", "p", "", "Project name")

	scoreConfigsGetCmd.Flags().
		BoolVar(&scoreConfigsGetJSON, "json", false, "Output raw API response as JSON")
	scoreConfigsGetCmd.Flags().
		StringVarP(&scoreConfigsGetOrg, "organization", "o", "", "Organization name that owns the project")
	scoreConfigsGetCmd.Flags().
		StringVarP(&scoreConfigsGetProject, "project", "p", "", "Project name")

	scoreConfigsCreateCmd.Flags().
		StringVar(&scoreConfigsCreateName, "name", "", "Config name (required)")
	scoreConfigsCreateCmd.Flags().
		StringVar(
			&scoreConfigsCreateDataType, "data-type", "",
			"Data type: NUMERIC, CATEGORICAL, or BOOLEAN (required)",
		)
	scoreConfigsCreateCmd.Flags().
		Float64Var(&scoreConfigsCreateMinValue, "min-value", 0, "Minimum value")
	scoreConfigsCreateCmd.Flags().
		Float64Var(&scoreConfigsCreateMaxValue, "max-value", 0, "Maximum value")
	scoreConfigsCreateCmd.Flags().
		StringVar(&scoreConfigsCreateCategories, "categories", "", "Categories as JSON array")
	scoreConfigsCreateCmd.Flags().
		StringVar(&scoreConfigsCreateDescription, "description", "", "Config description")
	scoreConfigsCreateCmd.Flags().
		BoolVar(&scoreConfigsCreateJSON, "json", false, "Output raw API response as JSON")
	scoreConfigsCreateCmd.Flags().
		StringVarP(&scoreConfigsCreateOrg, "organization", "o", "", "Organization name that owns the project")
	scoreConfigsCreateCmd.Flags().
		StringVarP(&scoreConfigsCreateProject, "project", "p", "", "Project name")
	_ = scoreConfigsCreateCmd.MarkFlagRequired("name")
	_ = scoreConfigsCreateCmd.MarkFlagRequired("data-type")

	scoreConfigsUpdateCmd.Flags().
		StringVar(&scoreConfigsUpdateDescription, "description", "", "New description")
	scoreConfigsUpdateCmd.Flags().
		BoolVar(&scoreConfigsUpdateIsArchived, "is-archived", false, "Set archived status")
	scoreConfigsUpdateCmd.Flags().
		Float64Var(&scoreConfigsUpdateMinValue, "min-value", 0, "New minimum value")
	scoreConfigsUpdateCmd.Flags().
		Float64Var(&scoreConfigsUpdateMaxValue, "max-value", 0, "New maximum value")
	scoreConfigsUpdateCmd.Flags().
		StringVar(&scoreConfigsUpdateCategories, "categories", "", "New categories as JSON array")
	scoreConfigsUpdateCmd.Flags().
		BoolVar(&scoreConfigsUpdateJSON, "json", false, "Output raw API response as JSON")
	scoreConfigsUpdateCmd.Flags().
		StringVarP(&scoreConfigsUpdateOrg, "organization", "o", "", "Organization name that owns the project")
	scoreConfigsUpdateCmd.Flags().
		StringVarP(&scoreConfigsUpdateProject, "project", "p", "", "Project name")

	scoreConfigsCmd.AddCommand(
		scoreConfigsListCmd,
		scoreConfigsGetCmd,
		scoreConfigsCreateCmd,
		scoreConfigsUpdateCmd,
	)
	rootCmd.AddCommand(scoreConfigsCmd)
}
