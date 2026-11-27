package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	collProject      string
	collOrganization string
	collDatabase     string
	collListAgent    string
	collFile         string
	collJSON         bool
	collYAML         bool
	collDryRun       bool
)

var collectionsCmd = &cobra.Command{
	Use:     "collections",
	Aliases: []string{"collection", "coll"},
	Short:   "Knowledge bases (searchable tables of chunks) inside a pgvector database",
	GroupID: groupContext,
	Long: `Manage collections within a database.

A collection is a table of chunks (rows) — each chunk is text plus its vector
embedding(s) — that you search by meaning or keyword; it's what backs a
knowledge base. It lives inside an existing pgvector database, so every command
requires --database ('list' takes --agent instead). Use 'iai databases create'
first to provision the database.

Run 'iai collections schema' to see the body format for every --file command.`,
}

var collListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List collections in a database, or the one an agent searches",
	Long: `List every collection in a database (-d), or only the collection an agent's
manifest searches (--agent, which defaults to the selected agent). With both, the
list is empty unless the agent's collection lives in that database.`,
	Example: `  iai collections list -d my-db
  iai collections list -d my-db --json
  iai collections list --agent my-agent`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		out := cmd.OutOrStdout()
		ctx := cmd.Context()

		name := ""
		if collDatabase == "" || cmd.Flags().Changed("agent") {
			if name, err = agentFlag(
				cmd,
				collListAgent,
				collOrganization,
				collProject,
			); err != nil {
				return err
			}
			if name == "" {
				return errors.New("pass -d or --agent, or select an agent")
			}
		}

		pCtx, apiClient, deployClient, err := resolveProject(ctx, collOrganization, collProject)
		if err != nil {
			return err
		}

		database, collection := collDatabase, ""
		if name != "" {
			agent, err := describeAgentOrDraft(ctx, apiClient, deployClient, pCtx, name)
			if err != nil {
				return err
			}
			database, collection = agentCollection(agent.AgentConfig, collDatabase)
		}
		collections := []deployment.CollectionSummary{}
		if database != "" {
			collections, err = deployClient.ListCollections(
				ctx,
				pCtx.orgId,
				pCtx.projectId,
				database,
			)
			if err != nil {
				return err
			}
		}
		if collection != "" {
			collections = slices.DeleteFunc(collections, func(c deployment.CollectionSummary) bool {
				return c.Name != collection
			})
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

// agentCollection is the database and collection the agent's manifest searches, or "" for
// both when it searches no collection or one outside database.
func agentCollection(agentConfig any, database string) (string, string) {
	config, _ := agentConfig.(map[string]any)
	search, _ := config["search"].(map[string]any)
	db, _ := search["database"].(string)
	name, _ := search["collection"].(string)
	if search["type"] != "collections" || db == "" || name == "" ||
		database != "" && database != db {
		return "", ""
	}
	return db, name
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
distance}) — and optional full-text search.

Slot type, dimension, distance, and the embedding model are IMMUTABLE after
creation; fixing a wrong value means deleting and recreating the collection.

Run 'iai collections schema' for the config file format.`,
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
			collDryRun,
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

		path := deployment.CollectionsPath(pCtx.orgId, pCtx.projectId, collDatabase) +
			"/" + url.PathEscape(name)
		msg, err := deployClient.SendCollectionBody(
			cmd.Context(), http.MethodPatch, path, body, "update collection", "",
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
		if c == collListCmd {
			continue
		}
		c.Flags().
			StringVarP(&collDatabase, "database", "d", "", "Database that holds the collection (required)")
		_ = c.MarkFlagRequired("database")
	}
	collListCmd.Flags().
		StringVarP(&collDatabase, "database", "d", "", "Database to list; required without --agent or a selected agent")
	collListCmd.Flags().StringVar(&collListAgent, "agent", "",
		"Filter by agent name; defaults to the selected agent")

	for _, c := range []*cobra.Command{collListCmd, collDescribeCmd, collStatsCmd} {
		c.Flags().BoolVar(&collJSON, "json", false, "Output raw API response as JSON")
		c.Flags().BoolVar(&collYAML, "yaml", false, "Output raw API response as YAML")
	}

	collCreateCmd.Flags().StringVar(&collFile, "file", "", "Path to a YAML/JSON collection config")
	collCreateCmd.Flags().
		BoolVar(&collDryRun, "dry-run", false, "Validate the config without creating the collection")
	_ = collCreateCmd.MarkFlagRequired("file")
	collPatchCmd.Flags().StringVar(&collFile, "file", "", "Path to a YAML/JSON patch config")
	_ = collPatchCmd.MarkFlagRequired("file")

	collectionsCmd.AddCommand(
		collListCmd, collDescribeCmd, collStatsCmd, collCreateCmd, collPatchCmd, collDeleteCmd,
		collSchemaCmd,
	)
	collSchemaCmd.Flags().BoolVar(&collJSON, "json", false, "Output the schemas as JSON")
	collSchemaCmd.Flags().BoolVar(&collYAML, "yaml", false, "Output the schemas as YAML")
	rootCmd.AddCommand(collectionsCmd)
}

var collSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Show the file schemas for the --file-based collection commands",
	Long: `Print the expected shape of every --file body: collection create/patch,
chunks upsert/patch, slots add/reindex, and search batch/hybrid. Use --json or
--yaml for structured output.`,
	Example: `  iai collections schema
  iai collections schema --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		if collJSON {
			return output.PrintStructuredJSON(out, collectionSchemas())
		}
		if collYAML {
			return output.PrintStructuredYAML(out, collectionSchemas())
		}
		fmt.Fprint(out, collectionSchemaText)
		return nil
	},
}

const collectionSchemaText = `Collection file schemas (bodies for --file). Keys are snake_case
(full_text, ef_search_default) except chunk record fields, which are camelCase (documentId).

create (--file):
  {
    "vectors": {
      "<slot>": { "type": "float32", "dimension": 768, "distance": "cosine" },
      "<slot>": { "embedding": { "model": "<model-id>" }, "dimension": 768 }
    },
    "full_text": { "enabled": true, "language": "english" }
  }
  NOTE: slot type, dimension, distance, and embedding model are IMMUTABLE.
  Provide either a raw slot (type+dimension) or an embedding slot (embedding+dimension).

patch (--file):
  { "full_text": { "enabled": false },
    "vectors": { "<slot>": { "index": { "ef_search_default": 200 } } } }

chunks upsert (--file):
  { "chunks": [
      { "id": "c1", "documentId": "doc1", "text": "content", "metadata": {"k":"v"} }
  ] }
  Fields: id, documentId, text, metadata, vector/vectors. Use documentId (NOT document_id).

chunks patch (--file):
  { "text": "new content", "metadata": {"k":"v"} }

slots add / reindex (--file):
  { "type": "float32", "dimension": 768, "distance": "cosine",
    "index": { "type": "hnsw", "m": 16, "ef_construct": 200 } }

search batch (--file):
  { "searches": [ {"query":"text","using":"default","limit":10} ] }

search hybrid (--file):
  { "queries": [
      {"query":"text","using":"default","candidate_limit":50},
      {"full_text":"keywords","candidate_limit":30}
    ],
    "fusion": {"method":"rrf","k":60}, "limit": 10 }
  NOTE: "using" selects a vector slot; use "full_text" for the keyword lane.
`

// collectionSchemas mirrors collectionSchemaText as a structured map for --json/--yaml.
func collectionSchemas() map[string]any {
	return map[string]any{
		"create": map[string]any{
			"vectors": map[string]any{
				"<slot>": map[string]any{
					"type": "float32", "dimension": 768, "distance": "cosine",
					"embedding": map[string]string{"model": "<model-id>"},
				},
			},
			"full_text": map[string]any{"enabled": true, "language": "english"},
			"immutable": []string{"type", "dimension", "distance", "embedding.model"},
		},
		"patch": map[string]any{
			"full_text": map[string]any{"enabled": false},
			"vectors": map[string]any{
				"<slot>": map[string]any{"index": map[string]int{"ef_search_default": 200}},
			},
		},
		"chunks_upsert": map[string]any{
			"chunks": []map[string]any{
				{
					"id":         "c1",
					"documentId": "doc1",
					"text":       "content",
					"metadata":   map[string]string{"k": "v"},
				},
			},
		},
		"chunks_patch": map[string]any{
			"text":     "new content",
			"metadata": map[string]string{"k": "v"},
		},
		"slots_add": map[string]any{
			"type": "float32", "dimension": 768, "distance": "cosine",
			"index": map[string]any{"type": "hnsw", "m": 16, "ef_construct": 200},
		},
		"search_batch": map[string]any{
			"searches": []map[string]any{{"query": "text", "using": "default", "limit": 10}},
		},
		"search_hybrid": map[string]any{
			"queries": []map[string]any{
				{"query": "text", "using": "default", "candidate_limit": 50},
				{"full_text": "keywords", "candidate_limit": 30},
			},
			"fusion": map[string]any{"method": "rrf", "k": 60}, "limit": 10,
		},
	}
}
