package cmd

import (
	"bufio"
	"context"
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
	mcpProject      string
	mcpOrganization string
)

var (
	mcpType            string
	mcpPort            int
	mcpImageType       string
	mcpImageRepository string
	mcpImageName       string
	mcpImageTag        string
	mcpMemory          string
	mcpCPU             string
	mcpEnvVars         []string
	mcpSecretRefs      []string
	mcpEndpointURL     string
	mcpCatalogID       string
	mcpAuthType        string
	mcpCredential      string
	mcpCredentialStdin bool
	mcpAuthHeader      string
	mcpAuthHeaderPfx   string
	mcpHeaders         []string
)

var mcpForce bool

var (
	mcpArgsJSON string
	mcpArgsFile string
)

var (
	mcpListJSON    bool
	mcpListYAML    bool
	mcpGetJSON     bool
	mcpGetYAML     bool
	mcpCatalogJSON bool
	mcpCatalogYAML bool
	mcpVerifyJSON  bool
	mcpVerifyYAML  bool
	mcpToolsJSON   bool
	mcpToolsYAML   bool
)


var mcpsCmd = &cobra.Command{
	Use:     "mcps",
	Aliases: []string{"mcp"},
	Short:   "Deploy and manage MCP servers",
	GroupID: groupInfra,
	Long: `Manage MCP servers for a project — in-cluster workloads ("internal"), custom
external URLs, or catalog-backed providers (external, external URL + auth derived
from the curated catalog).

Attach an mcp to an agent with '--mcp <name>' on 'iai agents create'/'update'.`,
}

var mcpCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Browse the curated MCP catalog",
	Long: `List curated MCP providers available to create an mcp from (see 'iai mcps create
--catalog-id'), showing each entry's id, category, and supported auth methods.`,
	Example: `  iai mcps catalog
  iai mcps catalog --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		data, raw, err := apiClient.ListMcpCatalog(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}

		if mcpCatalogJSON {
			return output.PrintRawJSON(out, raw)
		}
		if mcpCatalogYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintMcpCatalog(out, data.Entries)
	},
}

// runMcpUpsert builds the request body from the mcp* flag vars and submits
// it via submit (deployClient.CreateMcp or .PutMcp), sharing everything
// create and update have in common bar the verb-specific log line and client
// method.
func runMcpUpsert(cmd *cobra.Command, mcpName, verb string, submit func(
	deployClient *clients.DeploymentClient,
	ctx context.Context, orgId, projectId, mcpName string, body clients.CreateMcpBody,
) (string, error),
) error {
	out := cmd.OutOrStdout()

	cred, err := inputs.ResolveCredential(cmd.InOrStdin(), mcpCredential, mcpCredentialStdin)
	if err != nil {
		return err
	}

	reqBody, err := inputs.BuildMcpRequestBody(inputs.McpInput{
		Type:             mcpType,
		Port:             mcpPort,
		ImageType:        mcpImageType,
		ImageRepository:  mcpImageRepository,
		ImageName:        mcpImageName,
		ImageTag:         mcpImageTag,
		Memory:           mcpMemory,
		CPU:              mcpCPU,
		EnvVars:          mcpEnvVars,
		SecretRefs:       mcpSecretRefs,
		EndpointURL:      mcpEndpointURL,
		CatalogID:        mcpCatalogID,
		AuthType:         mcpAuthType,
		Credential:       cred,
		AuthHeader:       mcpAuthHeader,
		AuthHeaderPrefix: mcpAuthHeaderPfx,
		Headers:          mcpHeaders,
	})
	if err != nil {
		return err
	}

	pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Submitting mcp %s request...\n", verb)

	serverMessage, err := submit(
		deployClient,
		cmd.Context(),
		pCtx.orgId,
		pCtx.projectId,
		mcpName,
		reqBody,
	)
	if err != nil {
		return err
	}
	if serverMessage != "" {
		fmt.Fprintln(out, serverMessage)
	}
	return nil
}

var mcpCreateCmd = &cobra.Command{
	Use:   "create <mcp_name>",
	Short: "Create an mcp in a project",
	Long: `Create an mcp — an in-cluster MCP server ("internal"), a custom external
URL, or a catalog-backed provider.

Internal: --image-name, --image-tag, --port (like 'iai services create'); --env
and --secret load env vars from literal values or existing k8s Secrets.
External custom: --external-url — a server not owned by the platform, dialed
directly at that URL.
External catalog: --catalog-id (see 'iai mcps catalog'); external URL and auth are
derived from the catalog entry. Pass an auth type the entry supports; catalog
entries provide their own credential header and prefix.

The mcp is verified against the live server before it's kept: an internal mcp
is verified once its pod is up (checked in the background — see 'iai mcps get');
an external mcp (custom or catalog) is verified immediately, and the create
fails if the server is unreachable or rejects the credential.`,
	Example: `  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m
  iai mcps create acme --external-url https://mcp.acme.com/mcp --credential "$ACME_TOKEN"
  iai mcps create github --catalog-id github --credential "$GITHUB_TOKEN"
  iai mcps create github --catalog-id github --credential-stdin < token.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mcpName := strings.TrimSpace(args[0])
		return runMcpUpsert(cmd, mcpName, "creation", func(
			deployClient *clients.DeploymentClient,
			ctx context.Context, orgId, projectId, mcpName string, body clients.CreateMcpBody,
		) (string, error) {
			return deployClient.CreateMcp(ctx, orgId, projectId, mcpName, body)
		})
	},
}

var mcpUpdateCmd = &cobra.Command{
	Use:   "update <mcp_name>",
	Short: "Replace an mcp's spec",
	Long: `Full replace — pass the mcp's complete desired spec, same flags as create.
There is no partial-update mechanism for mcps; anything not passed resets to its
default. The type (internal/external) cannot change; delete and recreate instead.

Changing --credential rotates the mcp's Secret and restarts the mcp (if internal)
and every agent currently attached to it, so they pick up the new value.`,
	Example: `  iai mcps update my-tool --image-name my-mcp-server --image-tag v2 --port 8080
  iai mcps update acme --external-url https://mcp.acme.com/mcp --credential "$NEW_TOKEN"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mcpName := strings.TrimSpace(args[0])
		return runMcpUpsert(cmd, mcpName, "update", func(
			deployClient *clients.DeploymentClient,
			ctx context.Context, orgId, projectId, mcpName string, body clients.CreateMcpBody,
		) (string, error) {
			return deployClient.PutMcp(ctx, orgId, projectId, mcpName, body)
		})
	},
}

var mcpListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List mcps in a project",
	Example: `  iai mcps list
  iai mcps list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		mcps, err := deployClient.ListMcps(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}

		if mcpListJSON {
			return output.PrintStructuredJSON(out, mcps)
		}
		if mcpListYAML {
			return output.PrintStructuredYAML(out, mcps)
		}
		return output.PrintMcpList(out, mcps)
	},
}

var mcpGetCmd = &cobra.Command{
	Use:   "get <mcp_name>",
	Short: "Show mcp details, verify state, and cached tools",
	Long: `Show the mcp's record (type, external URL, catalog origin) and its latest
verify result — a tool count, not the tool list itself (see 'iai mcps tools').`,
	Example: `  iai mcps get my-tool
  iai mcps get my-tool --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, err := deployClient.DescribeMcp(cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName)
		if err != nil {
			return err
		}

		if mcpGetJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if mcpGetYAML {
			return output.PrintStructuredYAML(out, res)
		}
		return output.PrintMcpDetail(out, res)
	},
}

var mcpToolsCmd = &cobra.Command{
	Use:   "tools <mcp_name>",
	Short: "List an mcp's cached tools with descriptions",
	Long: `Show the full cached tool list — name and description. 'iai mcps get' only
shows a count; use this to see the tools themselves.`,
	Example: `  iai mcps tools my-tool
  iai mcps tools my-tool --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, err := deployClient.GetMcpTools(cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName)
		if err != nil {
			return err
		}
		if mcpToolsJSON {
			return output.PrintStructuredJSON(out, res.Tools)
		}
		if mcpToolsYAML {
			return output.PrintStructuredYAML(out, res.Tools)
		}
		return output.PrintMcpTools(out, res.Tools)
	},
}

var mcpVerifyCmd = &cobra.Command{
	Use:   "verify <mcp_name>",
	Short: "Re-verify an external mcp and refresh its cached tools",
	Long: `Re-dial the mcp (initialize + list tools) and refresh the cached tool list.
External mcps only — internal mcps verify automatically once their pod is up
(background reconciler; see 'iai mcps get') and reject a manual verify.`,
	Example: `  iai mcps verify my-tool
  iai mcps verify my-tool --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, err := deployClient.VerifyMcp(cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName)
		if err != nil {
			return err
		}

		if mcpVerifyJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if mcpVerifyYAML {
			return output.PrintStructuredYAML(out, res)
		}
		fmt.Fprintf(out, "Verified — %d tool(s) discovered", res.ToolCount)
		if res.ProtocolVersion != "" {
			fmt.Fprintf(out, " (protocol %s)", res.ProtocolVersion)
		}
		fmt.Fprintln(out)
		if res.Truncated {
			fmt.Fprintln(out, "Warning: tool list truncated to fit the cache size limit.")
		}
		return nil
	},
}

var mcpRunToolCmd = &cobra.Command{
	Use:   "run-tool <mcp_name> <tool_name>",
	Short: "Run a tool on an mcp",
	Long: `Call one of an mcp's tools and print the result.

Pass arguments as a JSON object with --args or --args-file (mutually exclusive);
omit both to send an empty object. Works for external mcps from anywhere;
internal mcps need the in-cluster operator.`,
	Example: `  iai mcps run-tool github search_repositories --args '{"query":"interactiveai"}'
  iai mcps run-tool github search_repositories --args-file ./args.json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])
		tool := strings.TrimSpace(args[1])

		toolArgs, err := inputs.ResolveToolArgs(mcpArgsJSON, mcpArgsFile)
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, err := deployClient.RunMcpTool(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName, tool, toolArgs,
		)
		if err != nil {
			return err
		}
		return output.PrintRawJSON(out, res.Result)
	},
}

// confirmDeletion stays here rather than in inputs: it is interactive I/O, not
// input parsing. Tolerating io.EOF honors input without a trailing newline
// (echo -n y); a bare EOF with no input declines rather than erroring.
func confirmDeletion(in io.Reader, out io.Writer, name string) (bool, error) {
	fmt.Fprintf(out, "This will delete mcp %q. Continue? [y/N] ", name)
	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	return strings.ToLower(strings.TrimSpace(answer)) == "y", nil
}

var mcpDeleteCmd = &cobra.Command{
	Use:     "delete <mcp_name>",
	Aliases: []string{"rm"},
	Short:   "Delete an mcp",
	Long: `Remove the mcp's release from the project namespace — its workload (if
internal), credential Secret, and cached tools. Rejected if agents are still
attached, unless -f is also set, in which case the delete proceeds and those
agents keep a dangling reference until it's removed. -f also skips the
confirmation prompt.

Detach it from any attached agent first with 'iai agents update <agent> --detach-mcp <mcp_name>'.`,
	Example: `  iai mcps delete my-tool
  iai mcps delete my-tool -f`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		if !mcpForce {
			confirmed, err := confirmDeletion(cmd.InOrStdin(), out, mcpName)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(out, "Aborted.")
				return nil
			}
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		serverMessage, err := deployClient.DeleteMcp(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			mcpName,
			mcpForce,
		)
		if err != nil {
			return err
		}
		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		} else {
			fmt.Fprintf(out, "Successfully deleted mcp %q.\n", mcpName)
		}
		return nil
	},
}

func init() {
	mcpsCmd.PersistentFlags().
		StringVarP(&mcpProject, "project", "p", "", "Project name that owns the mcps")
	mcpsCmd.PersistentFlags().
		StringVarP(&mcpOrganization, "organization", "o", "", "Organization name that owns the project")

	for _, c := range []*cobra.Command{mcpCreateCmd, mcpUpdateCmd} {
		c.Flags().
			StringVar(&mcpType, "type", "", `Mcp type: "internal" or "external" (inferred from other flags if omitted)`)
		c.Flags().IntVar(&mcpPort, "port", 0, "Port the mcp server listens on (internal)")
		c.Flags().
			StringVar(&mcpImageType, "image-type", "internal", `Image source: "internal" or "external" (internal)`)
		c.Flags().
			StringVar(&mcpImageRepository, "image-repository", "", "Image repository (required for external images)")
		c.Flags().StringVar(&mcpImageName, "image-name", "", "Container image name (internal)")
		c.Flags().StringVar(&mcpImageTag, "image-tag", "", "Container image tag (internal)")
		c.Flags().
			StringVar(&mcpMemory, "memory", "", "Memory request/limit, e.g. 512M (required for internal)")
		c.Flags().
			StringVar(&mcpCPU, "cpu", "", "CPU request/limit, e.g. 250m (required for internal)")
		c.Flags().
			StringArrayVar(&mcpEnvVars, "env", nil, "Environment variable (NAME=VALUE) for the mcp server; can be repeated (internal)")
		c.Flags().
			StringArrayVar(&mcpSecretRefs, "secret", nil, "Existing k8s Secret to load as env vars; can be repeated (internal)")
		c.Flags().
			StringVar(&mcpEndpointURL, "external-url", "", "External MCP server URL — not platform-owned, dialed directly (custom external mcp)")
		c.Flags().
			StringVar(&mcpCatalogID, "catalog-id", "", "Catalog entry id (see 'iai mcps catalog'); derives endpoint + auth (catalog external mcp)")
		c.Flags().
			StringVar(&mcpAuthType, "auth-type", "", `How the credential is sent: "bearer", "api_key", or "none" (inferred from --credential if omitted)`)
		c.Flags().
			StringVar(&mcpCredential, "credential", "", "Credential the mcp server requires (bearer token, API key)")
		c.Flags().
			BoolVar(&mcpCredentialStdin, "credential-stdin", false, "Read the credential from stdin instead of --credential")
		c.Flags().
			StringVar(&mcpAuthHeader, "auth-header", "", "Override the header the credential is sent in (default Authorization for bearer, X-API-Key for api_key)")
		c.Flags().
			StringVar(&mcpAuthHeaderPfx, "auth-header-prefix", "", `Override the credential value prefix (default "Bearer " for bearer)`)
		c.Flags().
			StringArrayVar(&mcpHeaders, "header", nil, "Extra non-secret request header (NAME=VALUE); can be repeated")
		c.MarkFlagsMutuallyExclusive("credential", "credential-stdin")
		c.MarkFlagsMutuallyExclusive("catalog-id", "external-url")
		c.MarkFlagsMutuallyExclusive("catalog-id", "image-name")
		c.MarkFlagsMutuallyExclusive("external-url", "image-name")
		// Catalog entries carry their own auth routing — these only apply to custom endpoints.
		c.MarkFlagsMutuallyExclusive("catalog-id", "auth-header")
		c.MarkFlagsMutuallyExclusive("catalog-id", "auth-header-prefix")
		c.MarkFlagsMutuallyExclusive("catalog-id", "header")
	}

	mcpRunToolCmd.Flags().
		StringVar(&mcpArgsJSON, "args", "", "Tool arguments as an inline JSON object")
	mcpRunToolCmd.Flags().
		StringVar(&mcpArgsFile, "args-file", "", "Path to a file containing the tool arguments as a JSON object")
	mcpRunToolCmd.MarkFlagsMutuallyExclusive("args", "args-file")

	mcpDeleteCmd.Flags().BoolVarP(&mcpForce, "force", "f", false, "Skip confirmation prompt")

	mcpListCmd.Flags().BoolVar(&mcpListJSON, "json", false, "Output raw API response as JSON")
	mcpListCmd.Flags().BoolVar(&mcpListYAML, "yaml", false, "Output raw API response as YAML")
	mcpListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	mcpGetCmd.Flags().BoolVar(&mcpGetJSON, "json", false, "Output raw API response as JSON")
	mcpGetCmd.Flags().BoolVar(&mcpGetYAML, "yaml", false, "Output raw API response as YAML")
	mcpGetCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	mcpCatalogCmd.Flags().BoolVar(&mcpCatalogJSON, "json", false, "Output raw API response as JSON")
	mcpCatalogCmd.Flags().BoolVar(&mcpCatalogYAML, "yaml", false, "Output raw API response as YAML")
	mcpCatalogCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	mcpVerifyCmd.Flags().BoolVar(&mcpVerifyJSON, "json", false, "Output raw API response as JSON")
	mcpVerifyCmd.Flags().BoolVar(&mcpVerifyYAML, "yaml", false, "Output raw API response as YAML")
	mcpVerifyCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	mcpToolsCmd.Flags().BoolVar(&mcpToolsJSON, "json", false, "Output raw API response as JSON")
	mcpToolsCmd.Flags().BoolVar(&mcpToolsYAML, "yaml", false, "Output raw API response as YAML")
	mcpToolsCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	rootCmd.AddCommand(mcpsCmd)
	mcpsCmd.AddCommand(
		mcpCatalogCmd,
		mcpCreateCmd,
		mcpUpdateCmd,
		mcpListCmd,
		mcpGetCmd,
		mcpToolsCmd,
		mcpVerifyCmd,
		mcpRunToolCmd,
		mcpDeleteCmd,
	)
}
