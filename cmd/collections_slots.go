package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	slotType      string
	slotDimension int
	slotDistance  string
	slotFile      string
)

var slotsCmd = &cobra.Command{
	Use:     "slots",
	Aliases: []string{"slot", "vectors"},
	Short:   "Manage a collection's vector slots and their indexes",
	Long:    `Add, reindex, vacuum, inspect, and remove the vector slots of a collection.`,
}

var slotsAddCmd = &cobra.Command{
	Use:   "add <collection> <slot>",
	Short: "Add a vector slot",
	Long: `Add a vector slot. Provide a raw vector slot via flags (--type, --dimension,
--distance) or a full slot config via --file (e.g. for an embedding-backed slot
or custom index tuning). --file takes precedence.`,
	Example: `  iai collections slots add docs title -d my-db --dimension 1536
  iai collections slots add docs title -d my-db --file slot.yaml`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		slot := strings.TrimSpace(args[1])

		body, err := slotAddBody()
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		res, err := deployClient.AddSlot(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			slot,
			body,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, res)
		}
		return output.PrintSlotAddResult(out, res)
	},
}

var slotsReindexCmd = &cobra.Command{
	Use:   "reindex <collection> <slot>",
	Short: "Rebuild a slot's index (online)",
	Long: `Rebuild a slot's index. With no --file it rebuilds with the current config;
--file (YAML/JSON) can change index params or quantization.`,
	Example: `  iai collections slots reindex docs title -d my-db
  iai collections slots reindex docs title -d my-db --file reindex.yaml`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		slot := strings.TrimSpace(args[1])

		body := []byte("{}")
		if slotFile != "" {
			b, err := inputs.ReadCollectionBodyFile(slotFile)
			if err != nil {
				return err
			}
			body = b
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		res, err := deployClient.ReindexSlot(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			slot,
			body,
		)
		if err != nil {
			return err
		}
		return output.PrintSlotOpResult(out, res)
	},
}

var slotsVacuumCmd = &cobra.Command{
	Use:     "vacuum <collection> <slot>",
	Short:   "Vacuum a slot (reclaim space, refresh stats)",
	Example: `  iai collections slots vacuum docs title -d my-db`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		slot := strings.TrimSpace(args[1])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		res, err := deployClient.VacuumSlot(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			slot,
		)
		if err != nil {
			return err
		}
		return output.PrintSlotOpResult(out, res)
	},
}

var slotsProgressCmd = &cobra.Command{
	Use:     "progress <collection> <slot>",
	Short:   "Show a slot's index build progress",
	Example: `  iai collections slots progress docs title -d my-db`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		slot := strings.TrimSpace(args[1])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		res, err := deployClient.SlotIndexProgressStatus(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			slot,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, res)
		}
		return output.PrintSlotIndexProgress(out, res)
	},
}

var slotsDeleteCmd = &cobra.Command{
	Use:     "delete <collection> <slot>",
	Aliases: []string{"rm"},
	Short:   "Delete a vector slot",
	Example: `  iai collections slots delete docs title -d my-db`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		slot := strings.TrimSpace(args[1])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		msg, err := deployClient.DeleteSlot(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			slot,
		)
		if err != nil {
			return err
		}
		if msg != "" {
			fmt.Fprintln(out, msg)
		}
		return nil
	},
}

// slotAddBody builds the add-slot body from --file (if set) or the flags.
func slotAddBody() ([]byte, error) {
	if slotFile != "" {
		return inputs.ReadCollectionBodyFile(slotFile)
	}
	return inputs.BuildAddSlotBody(slotType, slotDimension, slotDistance)
}

func init() {
	slotSubs := []*cobra.Command{
		slotsAddCmd, slotsReindexCmd, slotsVacuumCmd, slotsProgressCmd, slotsDeleteCmd,
	}
	for _, c := range slotSubs {
		c.Flags().StringVarP(&collOrganization, "organization", "o", "", "Organization name")
		c.Flags().StringVarP(&collProject, "project", "p", "", "Project name")
		c.Flags().
			StringVarP(&collDatabase, "database", "d", "", "Database that holds the collection (required)")
		_ = c.MarkFlagRequired("database")
	}

	slotsAddCmd.Flags().StringVar(&slotType, "type", "float32", "Vector slot type")
	slotsAddCmd.Flags().IntVar(&slotDimension, "dimension", 0, "Vector dimension (with flag form)")
	slotsAddCmd.Flags().
		StringVar(&slotDistance, "distance", "", "Distance metric (default: cosine)")
	slotsAddCmd.Flags().StringVar(&slotFile, "file", "", "Path to a YAML/JSON slot config")
	slotsAddCmd.Flags().BoolVar(&collJSON, "json", false, "Output raw API response as JSON")
	slotsAddCmd.Flags().BoolVar(&collYAML, "yaml", false, "Output raw API response as YAML")

	slotsReindexCmd.Flags().StringVar(&slotFile, "file", "", "Path to a YAML/JSON reindex config")

	slotsProgressCmd.Flags().BoolVar(&collJSON, "json", false, "Output raw API response as JSON")
	slotsProgressCmd.Flags().BoolVar(&collYAML, "yaml", false, "Output raw API response as YAML")

	slotsCmd.AddCommand(slotSubs...)
	collectionsCmd.AddCommand(slotsCmd)
}
