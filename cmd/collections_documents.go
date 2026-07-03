package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	docLimit         int
	docCursor        string
	docFilter        string
	docIncludeVector bool
)

var documentsCmd = &cobra.Command{
	Use:     "documents",
	Aliases: []string{"document", "docs"},
	Short:   "Inspect documents (chunks grouped by documentId)",
	Long:    `A document groups chunks by documentId; these commands read or delete them.`,
}

var documentsListCmd = &cobra.Command{
	Use:     "list <collection>",
	Aliases: []string{"ls"},
	Short:   "List documents in a collection",
	Example: `  iai collections documents list docs -d my-db`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		list, err := deployClient.ListDocuments(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			docLimit,
			docCursor,
			docFilter,
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
		return output.PrintDocumentList(out, list)
	},
}

var documentsGetCmd = &cobra.Command{
	Use:     "get <collection> <documentId>",
	Short:   "Get a document's chunks",
	Example: `  iai collections documents get docs support-faq -d my-db`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		documentID := strings.TrimSpace(args[1])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		doc, err := deployClient.GetDocument(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			documentID,
			docLimit,
			docCursor,
			docIncludeVector,
		)
		if err != nil {
			return err
		}

		if collJSON {
			return output.PrintStructuredJSON(out, doc)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, doc)
		}
		return output.PrintDocumentChunks(out, doc)
	},
}

var documentsDeleteCmd = &cobra.Command{
	Use:     "delete <collection> <documentId>",
	Aliases: []string{"rm"},
	Short:   "Delete a document (all chunks sharing the documentId)",
	Example: `  iai collections documents delete docs support-faq -d my-db`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])
		documentID := strings.TrimSpace(args[1])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		result, err := deployClient.DeleteDocument(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			documentID,
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
		fmt.Fprintf(out, "Deleted document %q (%d chunk(s))\n", result.DocumentID, result.DeletedCount)
		return nil
	},
}

func init() {
	docSubcommands := []*cobra.Command{documentsListCmd, documentsGetCmd, documentsDeleteCmd}
	for _, c := range docSubcommands {
		c.Flags().StringVarP(&collOrganization, "organization", "o", "", "Organization name")
		c.Flags().StringVarP(&collProject, "project", "p", "", "Project name")
		c.Flags().
			StringVarP(&collDatabase, "database", "d", "", "Database that holds the collection (required)")
		_ = c.MarkFlagRequired("database")
	}

	for _, c := range docSubcommands {
		c.Flags().BoolVar(&collJSON, "json", false, "Output raw API response as JSON")
		c.Flags().BoolVar(&collYAML, "yaml", false, "Output raw API response as YAML")
	}
	for _, c := range []*cobra.Command{documentsListCmd, documentsGetCmd} {
		c.Flags().IntVar(&docLimit, "limit", 0, "Page size (1-1000, default 100)")
		c.Flags().StringVar(&docCursor, "cursor", "", "Opaque cursor from a previous page")
	}
	documentsListCmd.Flags().
		StringVar(&docFilter, "filter", "", "Metadata filter as a JSON object")

	documentsGetCmd.Flags().
		BoolVar(&docIncludeVector, "include-vector", false, "Include the stored vector(s)")

	documentsCmd.AddCommand(docSubcommands...)
	collectionsCmd.AddCommand(documentsCmd)
}
