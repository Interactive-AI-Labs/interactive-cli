package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	chunkFile          string
	chunkLimit         int
	chunkCursor        string
	chunkPrefix        string
	chunkIncludeVector bool
	chunkIDs           []string
	chunkFilter        string
	chunkAll           bool
	chunkYes           bool
)

var chunksCmd = &cobra.Command{
	Use:   "chunks",
	Short: "Manage the chunks (rows) in a collection",
	Long:  `Upsert, inspect, and delete the chunks stored in a collection.`,
}

var chunksUpsertCmd = &cobra.Command{
	Use:   "upsert <collection>",
	Short: "Upsert chunks from a file",
	Long: `Upsert a batch of chunks from a YAML or JSON file (--file).

Chunks with text and no client vector are embedded server-side (set
defer_embedding=true with client-supplied vectors to skip embedding).`,
	Example: `  iai collections chunks upsert docs -d my-db --file chunks.json`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])

		body, err := inputs.ReadCollectionBodyFile(chunkFile)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		result, err := deployClient.UpsertChunks(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			body,
			collDryRun,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, result)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, result)
		}
		return output.PrintChunkUpsertResult(out, result)
	},
}

var chunksListCmd = &cobra.Command{
	Use:     "list <collection>",
	Aliases: []string{"ls"},
	Short:   "List chunks (keyset-paginated)",
	Example: `  iai collections chunks list docs -d my-db --limit 20
  iai collections chunks list docs -d my-db --cursor <token>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		list, err := deployClient.ListChunks(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			clients.ListChunksOpts{
				Limit:  chunkLimit,
				Cursor: chunkCursor,
				Prefix: chunkPrefix,
				Filter: chunkFilter,
			},
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, list)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, list)
		}
		return output.PrintChunkList(out, list)
	},
}

var chunksGetCmd = &cobra.Command{
	Use:   "get <collection> <id>",
	Short: "Get a single chunk",
	Example: `  iai collections chunks get docs chunk-1 -d my-db
  iai collections chunks get docs chunk-1 -d my-db --include-vector`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		id := strings.TrimSpace(args[1])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		chunk, err := deployClient.GetChunk(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			id,
			chunkIncludeVector,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, chunk)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, chunk)
		}
		return output.PrintChunk(out, chunk)
	},
}

var chunksCountCmd = &cobra.Command{
	Use:     "count <collection>",
	Short:   "Count chunks, optionally scoped by a metadata filter or id prefix",
	Example: `  iai collections chunks count docs -d my-db --filter '{"lang":"en"}'`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])

		body, err := inputs.BuildChunkCountBody(chunkFilter, chunkPrefix)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		count, err := deployClient.CountChunks(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			body,
		)
		if err != nil {
			return err
		}

		fmt.Fprintf(out, "%d\n", count)
		return nil
	},
}

var chunksPatchCmd = &cobra.Command{
	Use:     "patch <collection> <id>",
	Short:   "Update a chunk's metadata and/or text from a file",
	Example: `  iai collections chunks patch docs chunk-1 -d my-db --file patch.json`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		id := strings.TrimSpace(args[1])

		body, err := inputs.ReadCollectionBodyFile(chunkFile)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		chunk, err := deployClient.PatchChunk(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			id,
			body,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, chunk)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, chunk)
		}
		return output.PrintChunk(out, chunk)
	},
}

var chunksDeleteCmd = &cobra.Command{
	Use:     "delete <collection> <id>",
	Aliases: []string{"rm"},
	Short:   "Delete a single chunk by id",
	Example: `  iai collections chunks delete docs chunk-1 -d my-db`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		id := strings.TrimSpace(args[1])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		msg, err := deployClient.DeleteChunk(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			id,
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

var chunksBulkDeleteCmd = &cobra.Command{
	Use:   "bulk-delete <collection>",
	Short: "Delete many chunks by ids, metadata filter, or all",
	Long: `Delete chunks by exactly one selector: --ids, --filter, or --all.

--all deletes every chunk and requires confirmation.`,
	Example: `  iai collections chunks bulk-delete docs -d my-db --ids a,b,c
  iai collections chunks bulk-delete docs -d my-db --filter '{"lang":"en"}'
  iai collections chunks bulk-delete docs -d my-db --all`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])

		body, err := inputs.BuildBulkDeleteBody(chunkIDs, chunkFilter, chunkAll)
		if err != nil {
			return err
		}

		if chunkAll && !chunkYes {
			fmt.Fprintf(out, "Delete ALL chunks in %q? This cannot be undone. [y/N]: ", collection)
			line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}
			if ans := strings.ToLower(strings.TrimSpace(line)); ans != "y" && ans != "yes" {
				fmt.Fprintln(out, "Aborted.")
				return nil
			}
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		result, err := deployClient.BulkDeleteChunks(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			body,
			chunkAll,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, result)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, result)
		}
		return output.PrintBulkDeleteResult(out, result)
	},
}

func init() {
	chunkSubcommands := []*cobra.Command{
		chunksUpsertCmd, chunksListCmd, chunksGetCmd, chunksCountCmd,
		chunksPatchCmd, chunksDeleteCmd, chunksBulkDeleteCmd,
	}
	for _, c := range chunkSubcommands {
		c.Flags().StringVarP(&collOrganization, "organization", "o", "", "Organization name")
		c.Flags().StringVarP(&collProject, "project", "p", "", "Project name")
		c.Flags().
			StringVarP(&collDatabase, "database", "d", "", "Database that holds the collection (required)")
		_ = c.MarkFlagRequired("database")
	}

	for _, c := range []*cobra.Command{chunksUpsertCmd, chunksListCmd, chunksGetCmd, chunksPatchCmd, chunksBulkDeleteCmd} {
		c.Flags().BoolVar(&collJSON, "json", false, "Output raw API response as JSON")
		c.Flags().BoolVar(&collYAML, "yaml", false, "Output raw API response as YAML")
	}

	chunksUpsertCmd.Flags().StringVar(&chunkFile, "file", "", "Path to a YAML/JSON chunks file")
	chunksUpsertCmd.Flags().
		BoolVar(&collDryRun, "dry-run", false, "Validate the batch without embedding or storing")
	_ = chunksUpsertCmd.MarkFlagRequired("file")

	chunksListCmd.Flags().IntVar(&chunkLimit, "limit", 0, "Page size (1-1000, default 100)")
	chunksListCmd.Flags().
		StringVar(&chunkCursor, "cursor", "", "Opaque cursor from a previous page")
	chunksListCmd.Flags().
		StringVar(&chunkPrefix, "prefix", "", "Only chunks whose id has this prefix")
	chunksListCmd.Flags().
		StringVar(&chunkFilter, "filter", "", "Metadata filter as a JSON object")

	chunksGetCmd.Flags().
		BoolVar(&chunkIncludeVector, "include-vector", false, "Include the stored vector(s)")

	chunksCountCmd.Flags().StringVar(&chunkFilter, "filter", "", "Metadata filter as a JSON object")
	chunksCountCmd.Flags().
		StringVar(&chunkPrefix, "prefix", "", "Only count chunks with this id prefix")

	chunksPatchCmd.Flags().StringVar(&chunkFile, "file", "", "Path to a YAML/JSON patch file")
	_ = chunksPatchCmd.MarkFlagRequired("file")

	chunksBulkDeleteCmd.Flags().
		StringSliceVar(&chunkIDs, "ids", nil, "Comma-separated chunk ids to delete")
	chunksBulkDeleteCmd.Flags().
		StringVar(&chunkFilter, "filter", "", "Metadata filter as a JSON object")
	chunksBulkDeleteCmd.Flags().
		BoolVar(&chunkAll, "all", false, "Delete every chunk (requires confirm)")
	chunksBulkDeleteCmd.Flags().
		BoolVar(&chunkYes, "yes", false, "Skip the --all confirmation prompt")

	chunksCmd.AddCommand(chunkSubcommands...)
	collectionsCmd.AddCommand(chunksCmd)
}
