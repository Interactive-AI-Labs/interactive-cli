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

var connectorsCmd = &cobra.Command{
	Use:     "connectors",
	Aliases: []string{"connector"},
	Short:   "Manage MCP connectors in a project",
	GroupID: groupInfra,
	Long: `Manage MCP connectors in a project.

A connector stores the endpoint, transport, and credentials for an MCP server.
Pick one from the platform catalog or define a custom endpoint. Once connected,
verify to discover available tools and run them directly.`,
}

var connectorListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List connectors in a project",
	Long: `List MCP connectors in a project, showing type, status, tool count, and endpoint.

Examples:
  iai connectors list`,
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
		data, err := apiClient.ListMcpConnections(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}
		return output.PrintMcpConnectionList(out, data.Connections)
	},
}

var connectorGetCmd = &cobra.Command{
	Use:     "get <connector_id>",
	Aliases: []string{"describe", "desc"},
	Short:   "Show a connector and its tools",
	Long: `Show a connector in detail, including the cached list of tools discovered from
the MCP server.

Examples:
  iai connectors get 3f9c1a2e-...`,
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
		conn, err := apiClient.GetMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, id)
		if err != nil {
			return err
		}
		return output.PrintMcpConnectionDetail(out, conn)
	},
}

var connectorCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Browse the connector catalog",
	Long: `List the curated catalog of MCP servers you can connect to with
'iai connectors create --catalog-id', showing each entry's id, category, and
supported auth methods.

Examples:
  iai connectors catalog`,
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
		data, err := apiClient.ListMcpCatalog(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}
		return output.PrintMcpCatalog(out, data.Entries)
	},
}

var connectorCreateCmd = &cobra.Command{
	Use:   "create <connector_name>",
	Short: "Create a connector",
	Long: `Create an MCP connector, verified against the live server on save. If the server
cannot be reached or rejects the credential, creation fails and nothing is stored.

Pass --catalog-id to connect a catalog entry (the endpoint and transport come from
the catalog; see 'iai connectors catalog'). Otherwise the connector is custom and
--endpoint-url is required.

Examples:
  iai connectors create github \
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
		if err := inputs.ValidateMcpAuth(connectorAuthType, cred); err != nil {
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
			catalogData, err := apiClient.ListMcpCatalog(cmd.Context(), pCtx.orgId, pCtx.projectId)
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
			if err := inputs.ValidateMcpTransport(connectorTransport); err != nil {
				return err
			}
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
	Long: `Delete a connector and its cached tools. Does not affect the remote MCP server.
Use -f to skip confirmation.

Examples:
  iai connectors delete 3f9c1a2e-...
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
cached tool list. Reports the status and, on failure, the error class and message.

Examples:
  iai connectors verify 3f9c1a2e-...`,
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
		res, err := apiClient.VerifyMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, id)
		if err != nil {
			return err
		}
		return output.PrintMcpVerifyResult(out, res)
	},
}

var connectorRunToolCmd = &cobra.Command{
	Use:   "run-tool <connector_id> <tool_name>",
	Short: "Run a tool on a connector",
	Long: `Invoke a tool on a connector and print the result. Only enabled tools can run.

Pass arguments as a JSON object with --args or --args-file (mutually exclusive);
omit both to send an empty object.

Examples:
  iai connectors run-tool 3f9c1a2e-... search --args '{"query":"langfuse"}'
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

// emitToolResult prints a successful tool call and, for any non-ok status,
// returns a single formatted error instead of printing it. Returning the error
// gives a non-zero exit code so a failed call can't be silently chained with
// '&&', and routes one coherent message to stderr rather than a stdout block
// plus a generic stderr error.
func emitToolResult(out io.Writer, res *clients.McpToolCallData) error {
	if res.Status != "ok" {
		return output.McpToolCallError(res)
	}
	return output.PrintMcpToolResult(out, res)
}

func init() {
	connectorListCmd.Flags().
		StringVarP(&connectorProject, "project", "p", "", "Project name that owns the connectors")
	connectorListCmd.Flags().
		StringVarP(&connectorOrganization, "organization", "o", "", "Organization name that owns the project")

	connectorGetCmd.Flags().
		StringVarP(&connectorProject, "project", "p", "", "Project name that owns the connector")
	connectorGetCmd.Flags().
		StringVarP(&connectorOrganization, "organization", "o", "", "Organization name that owns the project")

	connectorCatalogCmd.Flags().
		StringVarP(&connectorProject, "project", "p", "", "Project name to browse the catalog for")
	connectorCatalogCmd.Flags().
		StringVarP(&connectorOrganization, "organization", "o", "", "Organization name that owns the project")

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
	connectorCreateCmd.Flags().
		StringVarP(&connectorProject, "project", "p", "", "Project name that owns the connector")
	connectorCreateCmd.Flags().
		StringVarP(&connectorOrganization, "organization", "o", "", "Organization name that owns the project")
	connectorCreateCmd.MarkFlagsMutuallyExclusive("credential", "credential-stdin")
	connectorCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "endpoint-url")
	connectorCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "transport")
	connectorCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "header")
	_ = connectorCreateCmd.MarkFlagRequired("auth-type")

	connectorDeleteCmd.Flags().
		BoolVarP(&connectorForce, "force", "f", false, "Skip confirmation prompt")
	connectorDeleteCmd.Flags().
		StringVarP(&connectorProject, "project", "p", "", "Project name that owns the connector")
	connectorDeleteCmd.Flags().
		StringVarP(&connectorOrganization, "organization", "o", "", "Organization name that owns the project")

	connectorVerifyCmd.Flags().
		StringVarP(&connectorProject, "project", "p", "", "Project name that owns the connector")
	connectorVerifyCmd.Flags().
		StringVarP(&connectorOrganization, "organization", "o", "", "Organization name that owns the project")

	connectorRunToolCmd.Flags().
		StringVar(&connectorArgsJSON, "args", "", "Tool arguments as an inline JSON object")
	connectorRunToolCmd.Flags().
		StringVar(&connectorArgsFile, "args-file", "", "Path to a file containing the tool arguments as a JSON object")
	connectorRunToolCmd.Flags().
		StringVarP(&connectorProject, "project", "p", "", "Project name that owns the connector")
	connectorRunToolCmd.Flags().
		StringVarP(&connectorOrganization, "organization", "o", "", "Organization name that owns the project")
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
