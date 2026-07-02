package cmd

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	searchQuery       string
	searchVector      string
	searchUsing       string
	searchLimit       int
	searchFilter      string
	searchExact       bool
	searchFile        string
	searchID          string
	searchExcludeSelf bool
)

var searchCmd = &cobra.Command{
	Use:   "search <collection>",
	Short: "Search a collection (single-lane vector search)",
	Long: `Run a single-lane search: --query (text, embedded server-side) or --vector
(comma-separated floats). --exact runs an exhaustive scan instead of the index.

Sub-commands cover the other modes: batch, by-id, hybrid.`,
	Example: `  iai collections search docs -d my-db --query "reset my password"
  iai collections search docs -d my-db --query "..." --exact --limit 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collection := strings.TrimSpace(args[0])

		var vector []float64
		if searchVector != "" {
			v, err := inputs.ParseVector(searchVector)
			if err != nil {
				return err
			}
			vector = v
		}
		body, err := inputs.BuildSearchBody(
			searchQuery,
			vector,
			searchUsing,
			searchLimit,
			searchFilter,
		)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		res, err := deployClient.Search(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			collDatabase,
			collection,
			body,
			searchExact,
		)
		if err != nil {
			return err
		}
		return printSearch(cmd, res)
	},
}

var searchBatchCmd = &cobra.Command{
	Use:     "batch <collection>",
	Short:   "Run several searches in one request (from a file)",
	Example: `  iai collections search batch docs -d my-db --file searches.json`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		collection := strings.TrimSpace(args[0])

		body, err := inputs.ReadCollectionBodyFile(searchFile)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		res, err := deployClient.SearchBatch(
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

		if collJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if collYAML {
			return output.PrintStructuredYAML(out, res)
		}
		return output.PrintBatchSearchResults(out, res)
	},
}

var searchByIDCmd = &cobra.Command{
	Use:     "by-id <collection>",
	Short:   "Find neighbors of an existing chunk by its stored vector",
	Example: `  iai collections search by-id docs -d my-db --id chunk-1 --exclude-self`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collection := strings.TrimSpace(args[0])

		body, err := inputs.BuildQueryByIDBody(
			searchID,
			searchUsing,
			searchLimit,
			searchExcludeSelf,
			searchFilter,
		)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		res, err := deployClient.QueryByID(
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
		return printSearch(cmd, res)
	},
}

var searchHybridCmd = &cobra.Command{
	Use:   "hybrid <collection>",
	Short: "Run a multi-lane hybrid search (RRF) from a file",
	Long: `Run a hybrid search from a YAML/JSON file. The command dispatches to the
hybrid path automatically (it sets "mode":"hybrid" for you).

The body holds a "queries" array and an optional "fusion" config. Each lane
supplies exactly one of: query (dense, embedded server-side), vector
(pre-computed dense), sparse_vector, or full_text (keyword search; requires
full-text enabled on the collection). Note: "using" selects the vector slot —
it does NOT select the full-text lane; set "full_text" for that. Lanes are
fused with RRF.

Schema:
  {
    "queries": [
      {"query": "text", "using": "default", "candidate_limit": 50},
      {"full_text": "keyword query", "candidate_limit": 30}
    ],
    "fusion": {"method": "rrf", "k": 60},
    "limit": 10
  }`,
	Example: `  iai collections search hybrid docs -d my-db --file hybrid.json`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collection := strings.TrimSpace(args[0])

		body, err := inputs.ReadCollectionBodyFile(searchFile)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), collOrganization, collProject)
		if err != nil {
			return err
		}

		res, err := deployClient.HybridSearch(
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
		return printSearch(cmd, res)
	},
}

// printSearch renders a SearchResponse honoring the --json/--yaml flags.
func printSearch(cmd *cobra.Command, res *clients.SearchResponse) error {
	out := cmd.OutOrStdout()
	if collJSON {
		return output.PrintStructuredJSON(out, res)
	}
	if collYAML {
		return output.PrintStructuredYAML(out, res)
	}
	return output.PrintSearchResults(out, res)
}

func init() {
	searchSubs := []*cobra.Command{searchCmd, searchBatchCmd, searchByIDCmd, searchHybridCmd}
	for _, c := range searchSubs {
		c.Flags().StringVarP(&collOrganization, "organization", "o", "", "Organization name")
		c.Flags().StringVarP(&collProject, "project", "p", "", "Project name")
		c.Flags().
			StringVarP(&collDatabase, "database", "d", "", "Database that holds the collection (required)")
		_ = c.MarkFlagRequired("database")
		c.Flags().BoolVar(&collJSON, "json", false, "Output raw API response as JSON")
		c.Flags().BoolVar(&collYAML, "yaml", false, "Output raw API response as YAML")
	}

	searchCmd.Flags().StringVar(&searchQuery, "query", "", "Query text (embedded server-side)")
	searchCmd.Flags().
		StringVar(&searchVector, "vector", "", "Query vector as comma-separated floats")
	searchCmd.Flags().
		StringVar(&searchUsing, "using", "", `Vector slot to search (omit for the server default, "default")`)
	searchCmd.Flags().IntVar(&searchLimit, "limit", 0, "Max results")
	searchCmd.Flags().StringVar(&searchFilter, "filter", "", "Metadata filter as a JSON object")
	searchCmd.Flags().BoolVar(&searchExact, "exact", false, "Exhaustive scan instead of the index")

	searchBatchCmd.Flags().
		StringVar(&searchFile, "file", "", "Path to a YAML/JSON batch-search file")
	_ = searchBatchCmd.MarkFlagRequired("file")

	searchByIDCmd.Flags().StringVar(&searchID, "id", "", "Seed chunk id (required)")
	_ = searchByIDCmd.MarkFlagRequired("id")
	searchByIDCmd.Flags().StringVar(&searchUsing, "using", "", "Vector slot to search")
	searchByIDCmd.Flags().IntVar(&searchLimit, "limit", 0, "Max results")
	searchByIDCmd.Flags().
		BoolVar(&searchExcludeSelf, "exclude-self", false, "Exclude the seed chunk")
	searchByIDCmd.Flags().StringVar(&searchFilter, "filter", "", "Metadata filter as a JSON object")

	searchHybridCmd.Flags().
		StringVar(&searchFile, "file", "", "Path to a YAML/JSON hybrid-search file")
	_ = searchHybridCmd.MarkFlagRequired("file")

	searchCmd.AddCommand(searchBatchCmd, searchByIDCmd, searchHybridCmd)
	collectionsCmd.AddCommand(searchCmd)
}
