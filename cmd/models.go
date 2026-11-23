package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	modelsListPage    int
	modelsListLimit   int
	modelsListSearch  string
	modelsListRegion  string
	modelsListInput   string
	modelsListOutput  string
	modelsListJSON    bool
	modelsListYAML    bool
	modelsListOrg     string
	modelsListProject string

	modelsGetJSON    bool
	modelsGetYAML    bool
	modelsGetOrg     string
	modelsGetProject string
)

var (
	routerModelInputModalities  = []string{"audio", "file", "image", "text", "video"}
	routerModelOutputModalities = []string{
		"audio",
		"decisions",
		"embeddings",
		"image",
		"rerank",
		"speech",
		"text",
		"transcription",
		"video",
	}
)

var modelsCmd = &cobra.Command{
	Use:              "models",
	Aliases:          []string{"model"},
	Short:            "List and inspect models",
	Long:             `List and inspect router models available to a project.`,
	PersistentPreRun: chainRootPersistentPreRun,
}

var modelsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List router models",
	Long:    `List router models for a project.`,
	Example: `  iai router models list
  iai router models list -o my-org -p my-project
  iai router models list --page 1 --limit 10
  iai router models list --search claude
  iai router models list --region global
  iai router models list --input-modality image
  iai router models list --input-modality audio --output-modality text
  iai router models list --json
  iai router models list --yaml`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		opts, err := inputs.PrepareRouterModelListOptions(platform.RouterModelListOptions{
			Page:           modelsListPage,
			Limit:          modelsListLimit,
			Search:         modelsListSearch,
			Region:         modelsListRegion,
			InputModality:  modelsListInput,
			OutputModality: modelsListOutput,
		})
		if err != nil {
			return err
		}

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), modelsListOrg, modelsListProject)
		if err != nil {
			return err
		}

		models, meta, raw, err := apiClient.ListRouterModels(
			cmd.Context(),
			pCtx.projectId,
			opts,
		)
		if err != nil {
			return err
		}

		if modelsListJSON {
			return output.PrintRawJSON(out, raw)
		}
		if modelsListYAML {
			return output.PrintRawYAML(out, raw)
		}

		return output.PrintRouterModelList(out, models, meta)
	},
}

var modelsGetCmd = &cobra.Command{
	Use:     "get <id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Get a router model",
	Long:    `Get detailed information about a router model by its ID.`,
	Example: `  iai router models get d34313ff-92ce-47ed-a1ae-fbb37f8a9411
  iai router models get d34313ff-92ce-47ed-a1ae-fbb37f8a9411 -o my-org -p my-project
  iai router models get d34313ff-92ce-47ed-a1ae-fbb37f8a9411 --json
  iai router models get d34313ff-92ce-47ed-a1ae-fbb37f8a9411 --yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		modelID := strings.TrimSpace(args[0])
		if modelID == "" {
			return fmt.Errorf("model id must not be empty")
		}

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), modelsGetOrg, modelsGetProject)
		if err != nil {
			return err
		}

		model, raw, err := apiClient.GetRouterModelByID(
			cmd.Context(),
			pCtx.projectId,
			modelID,
		)
		if err != nil {
			return err
		}

		if modelsGetJSON {
			return output.PrintRawJSON(out, raw)
		}
		if modelsGetYAML {
			return output.PrintRawYAML(out, raw)
		}

		return output.PrintRouterModelDetail(out, model)
	},
}

func init() {
	modelsListCmd.Flags().IntVar(&modelsListPage, "page", 0, "Page number (0-indexed)")
	modelsListCmd.Flags().IntVar(&modelsListLimit, "limit", 50, "Items per page (max 100)")
	indexFlag(modelsListCmd, "limit", "max 100")
	modelsListCmd.Flags().StringVar(&modelsListSearch, "search", "", "Search filter")
	modelsListCmd.Flags().StringVar(&modelsListRegion, "region", "", "Filter by region (global|eu)")
	addChoiceFlag(modelsListCmd, &modelsListInput, "input-modality", "",
		"Only models that accept this input", routerModelInputModalities)
	addChoiceFlag(modelsListCmd, &modelsListOutput, "output-modality", "",
		"Only models that produce this output", routerModelOutputModalities)
	modelsListCmd.Flags().
		BoolVar(&modelsListJSON, "json", false, "Output response as JSON")
	modelsListCmd.Flags().
		BoolVar(&modelsListYAML, "yaml", false, "Output response as YAML")
	modelsListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	modelsListCmd.Flags().
		StringVarP(&modelsListOrg, "organization", "o", "", "Organization name that owns the project")
	modelsListCmd.Flags().StringVarP(&modelsListProject, "project", "p", "", "Project name")

	modelsGetCmd.Flags().BoolVar(&modelsGetJSON, "json", false, "Output response as JSON")
	modelsGetCmd.Flags().BoolVar(&modelsGetYAML, "yaml", false, "Output response as YAML")
	modelsGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	modelsGetCmd.Flags().
		StringVarP(&modelsGetOrg, "organization", "o", "", "Organization name that owns the project")
	modelsGetCmd.Flags().StringVarP(&modelsGetProject, "project", "p", "", "Project name")

	modelsCmd.AddCommand(modelsListCmd, modelsGetCmd)
	routerCmd.AddCommand(modelsCmd)
}
