package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var validMcpAuthTypes = []string{"api_key", "bearer", "none"}
var validMcpTransports = []string{"streamable_http", "sse"}

func init() {
	parentCmd := &cobra.Command{
		Use:     "integrations",
		Aliases: []string{"integration", "mcp"},
		Short:   "MCP integration connections for a project",
		GroupID: groupInfra,
		Long: `Manage Model Context Protocol (MCP) integration connections in an InteractiveAI project.

Connections let agents reach external tools exposed by an MCP server — either a
curated catalog entry (a vendor-hosted server) or a custom endpoint you define.
Create a connection, verify it to discover its tools, then run a tool directly.`,
	}

	parentCmd.AddCommand(
		makeIntegrationsListCmd(),
		makeIntegrationsGetCmd(),
		makeIntegrationsCatalogCmd(),
		makeIntegrationsCreateCustomCmd(),
		makeIntegrationsCreateFromCatalogCmd(),
		makeIntegrationsDeleteCmd(),
		makeIntegrationsVerifyCmd(),
		makeIntegrationsToolsCmd(),
	)

	rootCmd.AddCommand(parentCmd)
}

func makeIntegrationsListCmd() *cobra.Command {
	var project, org string
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List integration connections in a project",
		Long: `List the MCP integration connections in a project, showing each connection's
type, status, tool count, and endpoint.

Examples:
  iai integrations list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
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
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connections")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}

func makeIntegrationsGetCmd() *cobra.Command {
	var project, org string
	cmd := &cobra.Command{
		Use:     "get <connection-id>",
		Aliases: []string{"describe", "desc"},
		Short:   "Show an integration connection and its tools",
		Long: `Show detailed information about a single integration connection, including the
cached list of tools discovered from the MCP server.

Examples:
  iai integrations get 3f9c1a2e-...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id := strings.TrimSpace(args[0])
			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
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
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}

func makeIntegrationsCatalogCmd() *cobra.Command {
	var project, org string
	cmd := &cobra.Command{
		Use:   "catalog",
		Short: "Browse the MCP integrations catalog",
		Long: `List the curated catalog of MCP servers you can connect to with
'iai integrations create-from-catalog', showing each entry's id, category, and
supported auth methods.

Examples:
  iai integrations catalog`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
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
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name to browse the catalog for")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}

func validateMcpAuth(authType, credential string) error {
	if !slices.Contains(validMcpAuthTypes, authType) {
		return fmt.Errorf("invalid --auth-type %q: must be one of %s", authType, strings.Join(validMcpAuthTypes, ", "))
	}
	if authType == "none" && credential != "" {
		return fmt.Errorf("--credential must not be set when --auth-type is 'none'")
	}
	if authType != "none" && credential == "" {
		return fmt.Errorf("--credential is required when --auth-type is %q", authType)
	}
	return nil
}

func validateMcpTransport(transport string) error {
	if !slices.Contains(validMcpTransports, transport) {
		return fmt.Errorf("invalid --transport %q: must be one of %s", transport, strings.Join(validMcpTransports, ", "))
	}
	return nil
}

func parseHeaderFlags(pairs []string) (map[string]string, error) {
	headers := make(map[string]string, len(pairs))
	for _, p := range pairs {
		key, value, found := strings.Cut(p, "=")
		if !found || key == "" {
			return nil, fmt.Errorf("invalid --header %q: expected KEY=VALUE", p)
		}
		headers[key] = value
	}
	return headers, nil
}

// resolveCredential returns the credential to send. With --credential-stdin it
// reads the credential from in (trimming a single trailing newline) so the
// secret never appears in the process list or shell history; otherwise it
// returns the --credential flag value unchanged.
func resolveCredential(in io.Reader, credential string, fromStdin bool) (string, error) {
	if !fromStdin {
		return credential, nil
	}
	data, err := io.ReadAll(in)
	if err != nil {
		return "", fmt.Errorf("failed to read credential from stdin: %w", err)
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}

func makeIntegrationsCreateCustomCmd() *cobra.Command {
	var (
		endpointURL     string
		authType        string
		credential      string
		credentialStdin bool
		transport       string
		slug            string
		description     string
		headers         []string
		project         string
		org             string
	)
	cmd := &cobra.Command{
		Use:   "create-custom <name>",
		Short: "Connect a custom MCP endpoint",
		Long: `Create an integration connection to a custom (user-defined) MCP endpoint.

The connection is verified against the live server on save; if the server cannot
be reached or rejects the credential, creation fails and nothing is stored.

Examples:
  iai integrations create-custom my-server \
    --endpoint-url https://mcp.example.com/mcp --auth-type none
  iai integrations create-custom github \
    --endpoint-url https://api.githubcopilot.com/mcp \
    --auth-type bearer --credential "$GITHUB_TOKEN"
  iai integrations create-custom github \
    --endpoint-url https://api.githubcopilot.com/mcp \
    --auth-type bearer --credential-stdin < token.txt
  iai integrations create-custom internal \
    --endpoint-url https://mcp.internal/sse --transport sse \
    --auth-type api_key --credential "$KEY" --header "X-Team=platform"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			credential, err := resolveCredential(cmd.InOrStdin(), credential, credentialStdin)
			if err != nil {
				return err
			}
			if err := validateMcpAuth(authType, credential); err != nil {
				return err
			}
			if err := validateMcpTransport(transport); err != nil {
				return err
			}
			customHeaders, err := parseHeaderFlags(headers)
			if err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			body := clients.McpConnectionCreateBody{
				Type:          "custom",
				Name:          name,
				Slug:          slug,
				Description:   description,
				EndpointURL:   endpointURL,
				Transport:     transport,
				AuthType:      authType,
				Credential:    credential,
				CustomHeaders: customHeaders,
			}

			fmt.Fprintf(out, "\nConnecting %q and verifying...\n\n", name)
			conn, err := apiClient.CreateMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, body)
			if err != nil {
				return err
			}
			return output.PrintMcpConnectionDetail(out, conn)
		},
	}
	cmd.Flags().StringVar(&endpointURL, "endpoint-url", "", "MCP server endpoint URL (required)")
	cmd.Flags().StringVar(&authType, "auth-type", "", "Auth type: api_key, bearer, or none (required)")
	cmd.Flags().StringVar(&credential, "credential", "", "API key or bearer token (required unless auth-type=none)")
	cmd.Flags().BoolVar(&credentialStdin, "credential-stdin", false, "Read the credential from stdin instead of --credential")
	cmd.Flags().StringVar(&transport, "transport", "streamable_http", "Transport: streamable_http (default) or sse")
	cmd.Flags().StringVar(&slug, "slug", "", "Tool prefix used as <slug>:<tool> (auto-derived from name if omitted)")
	cmd.Flags().StringVar(&description, "description", "", "Human-readable description")
	cmd.Flags().StringArrayVar(&headers, "header", nil, "Extra header as KEY=VALUE (repeatable)")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	cmd.MarkFlagsMutuallyExclusive("credential", "credential-stdin")
	_ = cmd.MarkFlagRequired("endpoint-url")
	_ = cmd.MarkFlagRequired("auth-type")
	return cmd
}

func makeIntegrationsCreateFromCatalogCmd() *cobra.Command {
	var (
		catalogID       string
		authType        string
		credential      string
		credentialStdin bool
		slug            string
		description     string
		project         string
		org             string
	)
	cmd := &cobra.Command{
		Use:   "create-from-catalog <name>",
		Short: "Connect an MCP server from the catalog",
		Long: `Create an integration connection from a curated catalog entry. The endpoint and
transport come from the catalog entry; you supply a name and (unless the entry
needs no auth) a credential.

Use 'iai integrations catalog' to find the --catalog-id.

The connection is verified against the live server on save.

Examples:
  iai integrations create-from-catalog github \
    --catalog-id github --auth-type bearer --credential "$GITHUB_TOKEN"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			name := strings.TrimSpace(args[0])

			credential, err := resolveCredential(cmd.InOrStdin(), credential, credentialStdin)
			if err != nil {
				return err
			}
			if err := validateMcpAuth(authType, credential); err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}

			body := clients.McpConnectionCreateBody{
				Type:        "platform",
				CatalogID:   catalogID,
				Name:        name,
				Slug:        slug,
				Description: description,
				AuthType:    authType,
				Credential:  credential,
			}

			fmt.Fprintf(out, "\nConnecting %q from catalog entry %q and verifying...\n\n", name, catalogID)
			conn, err := apiClient.CreateMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, body)
			if err != nil {
				return err
			}
			return output.PrintMcpConnectionDetail(out, conn)
		},
	}
	cmd.Flags().StringVar(&catalogID, "catalog-id", "", "Catalog entry id (required; see 'iai integrations catalog')")
	cmd.Flags().StringVar(&authType, "auth-type", "", "Auth type: api_key, bearer, or none (required)")
	cmd.Flags().StringVar(&credential, "credential", "", "API key or bearer token (required unless auth-type=none)")
	cmd.Flags().BoolVar(&credentialStdin, "credential-stdin", false, "Read the credential from stdin instead of --credential")
	cmd.Flags().StringVar(&slug, "slug", "", "Tool prefix used as <slug>:<tool> (auto-derived from name if omitted)")
	cmd.Flags().StringVar(&description, "description", "", "Human-readable description")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	cmd.MarkFlagsMutuallyExclusive("credential", "credential-stdin")
	_ = cmd.MarkFlagRequired("catalog-id")
	_ = cmd.MarkFlagRequired("auth-type")
	return cmd
}

// confirmDeletion prompts on out and reads a y/N answer from in. It returns true
// only when the user typed "y". Input that ends without a trailing newline (e.g.
// `echo -n y | ...`, where ReadString yields the bytes alongside io.EOF) is still
// honored; a bare EOF with no input is treated as a decline, not an error.
func confirmDeletion(in io.Reader, out io.Writer, id string) (bool, error) {
	fmt.Fprintf(out, "This will delete integration connection %q. Continue? [y/N] ", id)
	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	return strings.ToLower(strings.TrimSpace(answer)) == "y", nil
}

func makeIntegrationsDeleteCmd() *cobra.Command {
	var (
		force   bool
		project string
		org     string
	)
	cmd := &cobra.Command{
		Use:     "delete <connection-id>",
		Aliases: []string{"rm"},
		Short:   "Delete an integration connection",
		Long: `Delete an integration connection and its cached tools. This does not affect the
remote MCP server. Use -f to skip the confirmation prompt.

Examples:
  iai integrations delete 3f9c1a2e-...
  iai integrations delete 3f9c1a2e-... -f`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id := strings.TrimSpace(args[0])

			if !force {
				confirmed, err := confirmDeletion(cmd.InOrStdin(), out, id)
				if err != nil {
					return err
				}
				if !confirmed {
					fmt.Fprintln(out, "Aborted.")
					return nil
				}
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}
			if err := apiClient.DeleteMcpConnection(cmd.Context(), pCtx.orgId, pCtx.projectId, id); err != nil {
				return err
			}
			fmt.Fprintf(out, "Successfully deleted integration connection %q.\n", id)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}

func makeIntegrationsVerifyCmd() *cobra.Command {
	var project, org string
	cmd := &cobra.Command{
		Use:   "verify <connection-id>",
		Short: "Re-verify a connection and refresh its tools",
		Long: `Re-dial the MCP server for a connection (initialize + list tools) and refresh the
cached tool list. Reports the connection status and, on failure, the error class
and message.

Examples:
  iai integrations verify 3f9c1a2e-...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id := strings.TrimSpace(args[0])
			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
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
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	return cmd
}

func resolveToolArgs(inline, file string) (map[string]any, error) {
	raw := inline
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read --args-file %q: %w", file, err)
		}
		raw = string(data)
	}
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return nil, fmt.Errorf("invalid tool arguments: must be a JSON object: %w", err)
	}
	// JSON null unmarshals to a nil map without error; reject it like any other
	// non-object so it can't masquerade as an empty argument set.
	if args == nil {
		return nil, fmt.Errorf("invalid tool arguments: must be a JSON object, got null")
	}
	return args, nil
}

func makeIntegrationsToolsCmd() *cobra.Command {
	toolsCmd := &cobra.Command{
		Use:   "tools",
		Short: "Inspect and run tools on a connection",
		Long:  `Subcommands for working with the tools exposed by an integration connection.`,
	}
	toolsCmd.AddCommand(makeIntegrationsToolsRunCmd())
	return toolsCmd
}

func makeIntegrationsToolsRunCmd() *cobra.Command {
	var (
		argsJSON string
		argsFile string
		project  string
		org      string
	)
	cmd := &cobra.Command{
		Use:   "run <connection-id> <tool-name>",
		Short: "Run a tool on a connection",
		Long: `Invoke a tool exposed by a connection's MCP server and print the result.

Only enabled, server-advertised tools can run. Arguments are a JSON object passed
inline with --args or from a file with --args-file (mutually exclusive). When
omitted, an empty argument object is sent.

Examples:
  iai integrations tools run 3f9c1a2e-... search --args '{"query":"langfuse"}'
  iai integrations tools run 3f9c1a2e-... search --args-file ./args.json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			id := strings.TrimSpace(args[0])
			tool := strings.TrimSpace(args[1])

			toolArgs, err := resolveToolArgs(argsJSON, argsFile)
			if err != nil {
				return err
			}

			pCtx, apiClient, _, err := resolveProject(cmd.Context(), org, project)
			if err != nil {
				return err
			}
			res, err := apiClient.RunMcpTool(cmd.Context(), pCtx.orgId, pCtx.projectId, id, tool, toolArgs)
			if err != nil {
				return err
			}
			return output.PrintMcpToolResult(out, res)
		},
	}
	cmd.Flags().StringVar(&argsJSON, "args", "", "Tool arguments as an inline JSON object")
	cmd.Flags().StringVar(&argsFile, "args-file", "", "Path to a file containing the tool arguments as a JSON object")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Project name that owns the connection")
	cmd.Flags().StringVarP(&org, "organization", "o", "", "Organization name that owns the project")
	cmd.MarkFlagsMutuallyExclusive("args", "args-file")
	return cmd
}
