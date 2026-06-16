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
	connectorProject      string
	connectorOrganization string
)

var (
	connectorEndpointURL     string
	connectorCatalogID       string
	connectorAuthType        string
	connectorCredential      string
	connectorCredentialStdin bool
	connectorTransport       string
	connectorSlug            string
	connectorDescription     string
	connectorHeaders         []string
)

var connectorForce bool

var (
	connectorArgsJSON string
	connectorArgsFile string
)

var (
	connectorListJSON    bool
	connectorListYAML    bool
	connectorGetJSON     bool
	connectorGetYAML     bool
	connectorCatalogJSON bool
	connectorCatalogYAML bool
	connectorVerifyJSON  bool
	connectorVerifyYAML  bool
)

var connectorsCmd = &cobra.Command{
	Use:     "connectors",
	Aliases: []string{"connector"},
	Short:   "Manage MCP connectors in a project",
	GroupID: groupInfra,
	Long: `A connector stores the endpoint, transport, and credentials for an MCP server.
Pick one from the platform catalog or define a custom endpoint. Once connected,
verify to discover available tools and run them directly.`,
}

var connectorListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List connectors in a project",
	Long:    `Show each connector's type, status, tool count, and endpoint in a table.`,
	Example: `  iai connectors list
  iai connectors list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			connectorOrganization,
			connectorProject,
		)
		if err != nil {
			return err
		}
		data, raw, err := apiClient.ListMcpConnections(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}
		if connectorListJSON {
			return output.PrintRawJSON(out, raw)
		}
		if connectorListYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintMcpConnectionList(out, data.Connections)
	},
}

var connectorGetCmd = &cobra.Command{
	Use:     "get <connector_id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Show a connector and its tools",
	Long: `Print a connector's full configuration and status alongside the cached list of
tools discovered from the MCP server.`,
	Example: `  iai connectors get 3f9c1a2e-...
  iai connectors get 3f9c1a2e-... --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		id := strings.TrimSpace(args[0])
		if id == "" {
			return fmt.Errorf("connector id is required")
		}
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			connectorOrganization,
			connectorProject,
		)
		if err != nil {
			return err
		}
		conn, raw, err := apiClient.GetMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, id)
		if err != nil {
			return err
		}
		if connectorGetJSON {
			return output.PrintRawJSON(out, raw)
		}
		if connectorGetYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintMcpConnectionDetail(out, conn)
	},
}

var connectorCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Browse the connector catalog",
	Long: `List the curated catalog of MCP servers you can connect to with
'iai connectors create --catalog-id', showing each entry's id, category, and
supported auth methods.`,
	Example: `  iai connectors catalog
  iai connectors catalog --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			connectorOrganization,
			connectorProject,
		)
		if err != nil {
			return err
		}
		data, raw, err := apiClient.ListMcpCatalog(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}
		if connectorCatalogJSON {
			return output.PrintRawJSON(out, raw)
		}
		if connectorCatalogYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintMcpCatalog(out, data.Entries)
	},
}

var connectorCreateCmd = &cobra.Command{
	Use:   "create <connector_name>",
	Short: "Create a connector",
	Long: `Register an MCP server as a connector, verified against the live server on save.
If the server cannot be reached or rejects the credential, creation fails and
nothing is stored.

Pass --catalog-id to connect a catalog entry (the endpoint and transport come from
the catalog; see 'iai connectors catalog'). Otherwise the connector is custom and
--endpoint-url is required.`,
	Example: `  iai connectors create github \
    --catalog-id github --auth-type bearer --credential "$GITHUB_TOKEN"
  iai connectors create my-server \
    --endpoint-url https://mcp.example.com/mcp --auth-type none
  iai connectors create github \
    --endpoint-url https://api.githubcopilot.com/mcp \
    --auth-type bearer --credential-stdin < token.txt
  iai connectors create internal \
    --endpoint-url https://mcp.internal/sse --transport sse \
    --auth-type api_key --credential "$KEY" --header "X-Team=platform"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("connector name is required")
		}
		if connectorCatalogID == "" && connectorEndpointURL == "" {
			return fmt.Errorf(
				"--endpoint-url is required for a custom connector (or pass --catalog-id to use a catalog entry)",
			)
		}

		cred, err := inputs.ResolveCredential(
			cmd.InOrStdin(),
			connectorCredential,
			connectorCredentialStdin,
		)
		if err != nil {
			return err
		}

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			connectorOrganization,
			connectorProject,
		)
		if err != nil {
			return err
		}

		body := clients.McpConnectionCreateBody{
			Name:        name,
			Slug:        connectorSlug,
			Description: connectorDescription,
			AuthType:    connectorAuthType,
			Credential:  cred,
		}
		if connectorCatalogID != "" {
			// The backend requires the canonical endpoint_url even for catalog
			// connections (it verifies it matches the entry), so forward it.
			catalogData, _, err := apiClient.ListMcpCatalog(
				cmd.Context(),
				pCtx.orgId,
				pCtx.projectId,
			)
			if err != nil {
				return err
			}
			endpointURL, err := inputs.CatalogEndpointURL(catalogData.Entries, connectorCatalogID)
			if err != nil {
				return err
			}
			body.Type = "platform"
			body.CatalogID = connectorCatalogID
			body.EndpointURL = endpointURL
			fmt.Fprintf(
				out,
				"\nConnecting %q from catalog entry %q and verifying...\n\n",
				name,
				connectorCatalogID,
			)
		} else {
			customHeaders, err := inputs.ParseHeaderFlags(connectorHeaders)
			if err != nil {
				return err
			}
			body.Type = "custom"
			body.EndpointURL = connectorEndpointURL
			body.Transport = connectorTransport
			body.CustomHeaders = customHeaders
			fmt.Fprintf(out, "\nConnecting %q and verifying...\n\n", name)
		}

		conn, err := apiClient.CreateMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, body)
		if err != nil {
			return err
		}
		return output.PrintMcpConnectionDetail(out, conn)
	},
}

// confirmDeletion stays here rather than in inputs: it is interactive I/O, not
// input parsing. Tolerating io.EOF honors input without a trailing newline
// (echo -n y); a bare EOF with no input declines rather than erroring.
func confirmDeletion(in io.Reader, out io.Writer, id string) (bool, error) {
	fmt.Fprintf(out, "This will delete connector %q. Continue? [y/N] ", id)
	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	return strings.ToLower(strings.TrimSpace(answer)) == "y", nil
}

var connectorDeleteCmd = &cobra.Command{
	Use:     "delete <connector_id>",
	Aliases: []string{"rm"},
	Short:   "Delete a connector",
	Long: `Remove a connector and its cached tools from the project. The remote MCP server
is not affected. Use -f to skip the confirmation prompt.`,
	Example: `  iai connectors delete 3f9c1a2e-...
  iai connectors delete 3f9c1a2e-... -f`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		id := strings.TrimSpace(args[0])
		if id == "" {
			return fmt.Errorf("connector id is required")
		}

		if !connectorForce {
			confirmed, err := confirmDeletion(cmd.InOrStdin(), out, id)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(out, "Aborted.")
				return nil
			}
		}

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			connectorOrganization,
			connectorProject,
		)
		if err != nil {
			return err
		}
		if err := apiClient.DeleteMcpConnection(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			id,
		); err != nil {
			return err
		}
		fmt.Fprintf(out, "Successfully deleted connector %q.\n", id)
		return nil
	},
}

var connectorVerifyCmd = &cobra.Command{
	Use:   "verify <connector_id>",
	Short: "Re-verify a connector and refresh its tools",
	Long: `Re-dial the MCP server for a connector (initialize + list tools) and refresh the
cached tool list. Reports the status and, on failure, the error class and message.`,
	Example: `  iai connectors verify 3f9c1a2e-...
  iai connectors verify 3f9c1a2e-... --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		id := strings.TrimSpace(args[0])
		if id == "" {
			return fmt.Errorf("connector id is required")
		}
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			connectorOrganization,
			connectorProject,
		)
		if err != nil {
			return err
		}
		res, raw, err := apiClient.VerifyMcpConnection(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			id,
		)
		if err != nil {
			return err
		}
		if connectorVerifyJSON {
			return output.PrintRawJSON(out, raw)
		}
		if connectorVerifyYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintMcpVerifyResult(out, res)
	},
}

var connectorRunToolCmd = &cobra.Command{
	Use:   "run-tool <connector_id> <tool_name>",
	Short: "Run a tool on a connector",
	Long: `Call one of a connector's enabled tools and print the result it returns.

Pass arguments as a JSON object with --args or --args-file (mutually exclusive);
omit both to send an empty object.`,
	Example: `  iai connectors run-tool 3f9c1a2e-... search --args '{"query":"langfuse"}'
  iai connectors run-tool 3f9c1a2e-... search --args-file ./args.json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		id := strings.TrimSpace(args[0])
		tool := strings.TrimSpace(args[1])
		if id == "" {
			return fmt.Errorf("connector id is required")
		}
		if tool == "" {
			return fmt.Errorf("tool name is required")
		}

		toolArgs, err := inputs.ResolveToolArgs(connectorArgsJSON, connectorArgsFile)
		if err != nil {
			return err
		}

		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			connectorOrganization,
			connectorProject,
		)
		if err != nil {
			return err
		}
		res, err := apiClient.RunMcpTool(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			id,
			tool,
			toolArgs,
		)
		if err != nil {
			return err
		}
		return emitToolResult(out, res)
	},
}

// emitToolResult returns failed tool calls as errors so the command exits non-zero.
func emitToolResult(out io.Writer, res *clients.McpToolCallData) error {
	if res.Status != "ok" {
		return output.McpToolCallError(res)
	}
	return output.PrintMcpToolResult(out, res)
}

func init() {
	connectorsCmd.PersistentFlags().
		StringVarP(&connectorProject, "project", "p", "", "Project name that owns the connectors")
	connectorsCmd.PersistentFlags().
		StringVarP(&connectorOrganization, "organization", "o", "", "Organization name that owns the project")

	connectorListCmd.Flags().
		BoolVar(&connectorListJSON, "json", false, "Output raw API response as JSON")
	connectorListCmd.Flags().
		BoolVar(&connectorListYAML, "yaml", false, "Output raw API response as YAML")
	connectorGetCmd.Flags().
		BoolVar(&connectorGetJSON, "json", false, "Output raw API response as JSON")
	connectorGetCmd.Flags().
		BoolVar(&connectorGetYAML, "yaml", false, "Output raw API response as YAML")
	connectorCatalogCmd.Flags().
		BoolVar(&connectorCatalogJSON, "json", false, "Output raw API response as JSON")
	connectorCatalogCmd.Flags().
		BoolVar(&connectorCatalogYAML, "yaml", false, "Output raw API response as YAML")
	connectorVerifyCmd.Flags().
		BoolVar(&connectorVerifyJSON, "json", false, "Output raw API response as JSON")
	connectorVerifyCmd.Flags().
		BoolVar(&connectorVerifyYAML, "yaml", false, "Output raw API response as YAML")

	connectorCreateCmd.Flags().
		StringVar(&connectorCatalogID, "catalog-id", "", "Catalog entry id for a catalog connector (see 'iai connectors catalog')")
	connectorCreateCmd.Flags().
		StringVar(&connectorEndpointURL, "endpoint-url", "", "MCP server endpoint URL (required for a custom connector)")
	connectorCreateCmd.Flags().
		StringVar(&connectorAuthType, "auth-type", "", "Auth type: api_key, bearer, or none (required)")
	connectorCreateCmd.Flags().
		StringVar(&connectorCredential, "credential", "", "API key or bearer token (required unless auth-type=none)")
	connectorCreateCmd.Flags().
		BoolVar(&connectorCredentialStdin, "credential-stdin", false, "Read the credential from stdin instead of --credential")
	connectorCreateCmd.Flags().
		StringVar(&connectorTransport, "transport", "streamable_http", "Transport for a custom connector: streamable_http (default) or sse")
	connectorCreateCmd.Flags().
		StringVar(&connectorSlug, "slug", "", "Tool prefix used as <slug>:<tool> (auto-derived from name if omitted)")
	connectorCreateCmd.Flags().
		StringVar(&connectorDescription, "description", "", "Human-readable description")
	connectorCreateCmd.Flags().
		StringArrayVar(&connectorHeaders, "header", nil, "Extra header as KEY=VALUE for a custom connector (repeatable)")
	connectorCreateCmd.MarkFlagsMutuallyExclusive("credential", "credential-stdin")
	connectorCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "endpoint-url")
	connectorCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "transport")
	connectorCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "header")
	_ = connectorCreateCmd.MarkFlagRequired("auth-type")

	connectorDeleteCmd.Flags().
		BoolVarP(&connectorForce, "force", "f", false, "Skip confirmation prompt")

	connectorRunToolCmd.Flags().
		StringVar(&connectorArgsJSON, "args", "", "Tool arguments as an inline JSON object")
	connectorRunToolCmd.Flags().
		StringVar(&connectorArgsFile, "args-file", "", "Path to a file containing the tool arguments as a JSON object")
	connectorRunToolCmd.MarkFlagsMutuallyExclusive("args", "args-file")

	rootCmd.AddCommand(connectorsCmd)
	connectorsCmd.AddCommand(connectorListCmd)
	connectorsCmd.AddCommand(connectorGetCmd)
	connectorsCmd.AddCommand(connectorCatalogCmd)
	connectorsCmd.AddCommand(connectorCreateCmd)
	connectorsCmd.AddCommand(connectorDeleteCmd)
	connectorsCmd.AddCommand(connectorVerifyCmd)
	connectorsCmd.AddCommand(connectorRunToolCmd)
}
