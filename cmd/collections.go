package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	collProject      string
	collOrganization string
	collDatabase     string
	collFile         string
	collJSON         bool
	collYAML         bool
)

var collectionsCmd = &cobra.Command{
	Use:     "collections",
	Aliases: []string{"collection", "coll"},
	Short:   "Vector collections (knowledge bases) inside a pgvector database",
	GroupID: groupInfra,
	Long: `Manage vector collections within a database.

A collection is a vector store (knowledge base) that lives inside an existing
pgvector database, so every command requires --database. Use 'iai databases
create' first to provision the database.`,
}

var collListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List collections in a database",
	Example: `  iai collections list -d my-db
  iai collections list -d my-db --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		collections, err := deployClient.ListCollections(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, collections)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, collections)
		}
		return output.PrintCollectionList(out, collections)
	},
}

var collDescribeCmd = &cobra.Command{
	Use:     "describe <collection>",
	Aliases: []string{"desc"},
	Short:   "Describe a collection's configuration",
	Example: `  iai collections describe docs -d my-db
  iai collections describe docs -d my-db --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		coll, err := deployClient.DescribeCollection(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			name,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, coll)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, coll)
		}
		return output.PrintCollectionDescribe(out, coll)
	},
}

var collStatsCmd = &cobra.Command{
	Use:   "stats <collection>",
	Short: "Show a collection's chunk count, size, and index status",
	Example: `  iai collections stats docs -d my-db
  iai collections stats docs -d my-db --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		stats, err := deployClient.GetCollectionStats(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			name,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, stats)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, stats)
		}
		return output.PrintCollectionStats(out, stats)
	},
}

var collCreateCmd = &cobra.Command{
	Use:   "create <collection>",
	Short: "Create a collection from a config file",
	Long: `Create a vector collection from a YAML or JSON config file (--file).

The config declares the vector slot(s) — either an embedding-backed slot
("embedding": {model, dimension}) or a raw vector slot ({type, dimension,
distance}) — and optional full-text search.`,
	Example: `  iai collections create docs -d my-db --file collection.yaml`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		body, err := inputs.ReadCollectionBodyFile(collFile)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		msg, err := deployClient.CreateCollection(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			name,
			body,
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

var collPatchCmd = &cobra.Command{
	Use:   "patch <collection>",
	Short: "Update a collection's mutable config from a file",
	Long: `Update a collection's mutable configuration from a YAML or JSON file (--file):
full-text settings and per-slot ef_search_default. Slot type/dimension/distance
and the embedding model are immutable.`,
	Example: `  iai collections patch docs -d my-db --file patch.yaml`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		body, err := inputs.ReadCollectionBodyFile(collFile)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		msg, err := deployClient.PatchCollection(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			name,
			body,
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

var collDeleteCmd = &cobra.Command{
	Use:     "delete <collection>",
	Aliases: []string{"rm"},
	Short:   "Delete a collection and all its data",
	Example: `  iai collections delete docs -d my-db`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		msg, err := deployClient.DeleteCollection(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			name,
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

func init() {
	// Shared scope flags on every subcommand: -o org, -p project, -d database.
	for _, c := range []*cobra.Command{
		collListCmd, collDescribeCmd, collStatsCmd, collCreateCmd, collPatchCmd, collDeleteCmd,
	} {
		c.Flags().StringVarP(&collOrganization, "organization", "o", "", "Organization name")
		c.Flags().StringVarP(&collProject, "project", "p", "", "Project name")
		c.Flags().
			StringVarP(&collDatabase, "database", "d", "", "Database that holds the collection (required)")
		_ = c.MarkFlagRequired("database")
	}

	for _, c := range []*cobra.Command{collListCmd, collDescribeCmd, collStatsCmd} {
		c.Flags().BoolVar(&collJSON, "json", false, "Output raw API response as JSON")
		c.Flags().BoolVar(&collYAML, "yaml", false, "Output raw API response as YAML")
	}

	collCreateCmd.Flags().StringVar(&collFile, "file", "", "Path to a YAML/JSON collection config")
	_ = collCreateCmd.MarkFlagRequired("file")
	collPatchCmd.Flags().StringVar(&collFile, "file", "", "Path to a YAML/JSON patch config")
	_ = collPatchCmd.MarkFlagRequired("file")

	collectionsCmd.AddCommand(
		collListCmd, collDescribeCmd, collStatsCmd, collCreateCmd, collPatchCmd, collDeleteCmd,
	)
	rootCmd.AddCommand(collectionsCmd)
}
