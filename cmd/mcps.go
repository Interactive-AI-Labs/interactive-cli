package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

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
	mcpEndpointURL     string
	mcpCatalogID       string
	mcpCredential      string
	mcpCredentialStdin bool
)

var mcpForce bool
var mcpGetVersion string

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
)

var mcpsCmd = &cobra.Command{
	Use:     "mcps",
	Aliases: []string{"mcp"},
	Short:   "Deploy and manage MCP servers",
	GroupID: groupInfra,
	Long: `Manage MCP servers for a project — in-cluster workloads ("internal"), custom
external endpoints, or catalog-backed providers (external, endpoint + auth derived
from the curated catalog).

Attach an mcp to an agent with '--mcp <name>' on 'iai agents create'/'update'.`,
}

var mcpCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Browse the curated MCP catalog",
	Long: `List curated MCP providers available to create an mcp from (see 'iai mcps create
--catalog-id'). Catalog browsing is platform data — not project-scoped infrastructure —
so this reuses the same listing 'iai connectors catalog' uses.`,
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

var mcpCreateCmd = &cobra.Command{
	Use:   "create <mcp_name>",
	Short: "Create an mcp in a project",
	Long: `Create an mcp — an in-cluster MCP server ("internal"), a custom external
endpoint, or a catalog-backed provider.

Internal: --image-name, --image-tag, --port (like 'iai services create').
External custom: --endpoint-url.
External catalog: --catalog-id (see 'iai mcps catalog'); endpoint and auth are
derived from the catalog entry — only entries resolvable to plain
Authorization: Bearer (or no auth) can be used this way.

The mcp is verified against the live server before it's kept: an internal mcp
is verified once its pod is up (checked in the background — see 'iai mcps get');
an external mcp (custom or catalog) is verified immediately, and the create
fails if the server is unreachable or rejects the credential.`,
	Example: `  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080
  iai mcps create acme --endpoint-url https://mcp.acme.com/mcp --credential "$ACME_TOKEN"
  iai mcps create github --catalog-id github --credential "$GITHUB_TOKEN"
  iai mcps create github --catalog-id github --credential-stdin < token.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		cred, err := inputs.ResolveCredential(cmd.InOrStdin(), mcpCredential, mcpCredentialStdin)
		if err != nil {
			return err
		}

		reqBody, err := inputs.BuildMcpRequestBody(inputs.McpInput{
			Type:            mcpType,
			Port:            mcpPort,
			ImageType:       mcpImageType,
			ImageRepository: mcpImageRepository,
			ImageName:       mcpImageName,
			ImageTag:        mcpImageTag,
			Memory:          mcpMemory,
			CPU:             mcpCPU,
			EnvVars:         mcpEnvVars,
			EndpointURL:     mcpEndpointURL,
			CatalogID:       mcpCatalogID,
			Credential:      cred,
		})
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting mcp creation request...")

		serverMessage, err := deployClient.CreateMcp(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName, reqBody,
		)
		if err != nil {
			return err
		}
		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}
		return nil
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
  iai mcps update acme --endpoint-url https://mcp.acme.com/mcp --credential "$NEW_TOKEN"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		cred, err := inputs.ResolveCredential(cmd.InOrStdin(), mcpCredential, mcpCredentialStdin)
		if err != nil {
			return err
		}

		reqBody, err := inputs.BuildMcpRequestBody(inputs.McpInput{
			Type:            mcpType,
			Port:            mcpPort,
			ImageType:       mcpImageType,
			ImageRepository: mcpImageRepository,
			ImageName:       mcpImageName,
			ImageTag:        mcpImageTag,
			Memory:          mcpMemory,
			CPU:             mcpCPU,
			EnvVars:         mcpEnvVars,
			EndpointURL:     mcpEndpointURL,
			CatalogID:       mcpCatalogID,
			Credential:      cred,
		})
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting mcp update request...")

		serverMessage, err := deployClient.PutMcp(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName, reqBody,
		)
		if err != nil {
			return err
		}
		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}
		return nil
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
	Long: `Show the mcp's record (type, endpoint, catalog origin), its last verify result,
and the tools discovered there.

--version reads a past image version's cached tools instead of the latest verify
(internal mcps only — the cache is keyed by image tag, see 'iai mcps update').`,
	Example: `  iai mcps get my-tool
  iai mcps get my-tool --version v1
  iai mcps get my-tool --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, err := deployClient.DescribeMcp(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName, mcpGetVersion,
		)
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

var mcpVerifyCmd = &cobra.Command{
	Use:   "verify <mcp_name>",
	Short: "Re-verify an mcp and refresh its cached tools",
	Long: `Re-dial the mcp (initialize + list tools) and refresh the cached tool list —
also cached per image version for internal mcps. Reports tool count and, on
failure, the error.`,
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
		return output.PrintStructuredJSON(out, res.Result)
	},
}

func confirmMcpDeletion(in io.Reader, out io.Writer, name string) (bool, error) {
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
internal), credential Secret, and cached tools. Agents still attached to it will
fail to resolve on their next deploy. Use -f to skip the confirmation prompt.`,
	Example: `  iai mcps delete my-tool
  iai mcps delete my-tool -f`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		if !mcpForce {
			confirmed, err := confirmMcpDeletion(cmd.InOrStdin(), out, mcpName)
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

		serverMessage, err := deployClient.DeleteMcp(cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName)
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
		c.Flags().StringVar(&mcpType, "type", "", `Mcp type: "internal" or "external" (inferred from other flags if omitted)`)
		c.Flags().IntVar(&mcpPort, "port", 0, "Port the mcp server listens on (internal)")
		c.Flags().StringVar(&mcpImageType, "image-type", "internal", `Image source: "internal", "external", or "platform" (internal)`)
		c.Flags().StringVar(&mcpImageRepository, "image-repository", "", "Image repository (required for external/platform images)")
		c.Flags().StringVar(&mcpImageName, "image-name", "", "Container image name (internal)")
		c.Flags().StringVar(&mcpImageTag, "image-tag", "", "Container image tag (internal)")
		c.Flags().StringVar(&mcpMemory, "memory", "", "Memory request/limit, e.g. 512M (internal, default 512M)")
		c.Flags().StringVar(&mcpCPU, "cpu", "", "CPU request/limit, e.g. 250m (internal, default 250m)")
		c.Flags().StringArrayVar(&mcpEnvVars, "env", nil, "Environment variable (NAME=VALUE) for the mcp server; can be repeated (internal)")
		c.Flags().StringVar(&mcpEndpointURL, "endpoint-url", "", "MCP server endpoint URL (custom external mcp)")
		c.Flags().StringVar(&mcpCatalogID, "catalog-id", "", "Catalog entry id (see 'iai mcps catalog'); derives endpoint + auth (catalog external mcp)")
		c.Flags().StringVar(&mcpCredential, "credential", "", "Credential the mcp server requires (bearer token, API key)")
		c.Flags().BoolVar(&mcpCredentialStdin, "credential-stdin", false, "Read the credential from stdin instead of --credential")
		c.MarkFlagsMutuallyExclusive("credential", "credential-stdin")
		c.MarkFlagsMutuallyExclusive("catalog-id", "endpoint-url")
		c.MarkFlagsMutuallyExclusive("catalog-id", "image-name")
		c.MarkFlagsMutuallyExclusive("endpoint-url", "image-name")
	}

	mcpGetCmd.Flags().StringVar(&mcpGetVersion, "version", "", "Read this image tag's cached tools instead of the latest verify (internal only)")

	mcpRunToolCmd.Flags().StringVar(&mcpArgsJSON, "args", "", "Tool arguments as an inline JSON object")
	mcpRunToolCmd.Flags().StringVar(&mcpArgsFile, "args-file", "", "Path to a file containing the tool arguments as a JSON object")
	mcpRunToolCmd.MarkFlagsMutuallyExclusive("args", "args-file")

	mcpDeleteCmd.Flags().BoolVarP(&mcpForce, "force", "f", false, "Skip confirmation prompt")

	mcpListCmd.Flags().BoolVar(&mcpListJSON, "json", false, "Output raw API response as JSON")
	mcpListCmd.Flags().BoolVar(&mcpListYAML, "yaml", false, "Output raw API response as YAML")
	mcpGetCmd.Flags().BoolVar(&mcpGetJSON, "json", false, "Output raw API response as JSON")
	mcpGetCmd.Flags().BoolVar(&mcpGetYAML, "yaml", false, "Output raw API response as YAML")
	mcpCatalogCmd.Flags().BoolVar(&mcpCatalogJSON, "json", false, "Output raw API response as JSON")
	mcpCatalogCmd.Flags().BoolVar(&mcpCatalogYAML, "yaml", false, "Output raw API response as YAML")
	mcpVerifyCmd.Flags().BoolVar(&mcpVerifyJSON, "json", false, "Output raw API response as JSON")
	mcpVerifyCmd.Flags().BoolVar(&mcpVerifyYAML, "yaml", false, "Output raw API response as YAML")

	rootCmd.AddCommand(mcpsCmd)
	mcpsCmd.AddCommand(
		mcpCatalogCmd,
		mcpCreateCmd,
		mcpUpdateCmd,
		mcpListCmd,
		mcpGetCmd,
		mcpVerifyCmd,
		mcpRunToolCmd,
		mcpDeleteCmd,
	)
}
