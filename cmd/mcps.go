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
	mcpPath            string
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

var (
	mcpClearEnv     bool
	mcpClearSecret  bool
	mcpClearHeaders bool
)

var mcpForce bool

var (
	mcpArgsJSON string
	mcpArgsFile string
)

var (
	mcpListJSON     bool
	mcpListYAML     bool
	mcpDescribeJSON bool
	mcpDescribeYAML bool
	mcpCatalogJSON  bool
	mcpCatalogYAML  bool
	mcpVerifyJSON   bool
	mcpVerifyYAML   bool
	mcpToolsJSON    bool
	mcpToolsYAML    bool
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

var mcpCreateCmd = &cobra.Command{
	Use:   "create <mcp_name>",
	Short: "Create an mcp in a project",
	Long: `Create an mcp — an in-cluster MCP server ("internal"), a custom external
URL, or a catalog-backed provider.

Internal: --image-name, --image-tag, --port; --env and --secret load env vars
from literal values or existing secrets. --path is the endpoint path the mcp's
own server exposes (default "/mcp" — set to whatever the mcp owner actually
configured, don't assume).
External custom: --external-url — a server not owned by the platform, dialed
directly at that URL, path included.
External catalog: --catalog-id (see 'iai mcps catalog'); external URL and auth are
derived from the catalog entry. Pass an auth type the entry supports; catalog
entries provide their own credential header and prefix.

The mcp is verified against the live server before it's kept: an internal mcp
is verified once its status is healthy (checked in the background — see 'iai
mcps describe'); an external mcp (custom or catalog) is verified immediately,
and the create fails if the server is unreachable or rejects the credential.`,
	Example: `  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m
  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m --path /api/mcp
  iai mcps create acme --external-url https://mcp.acme.com/mcp --credential "$ACME_TOKEN"
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
			Type:             mcpType,
			Port:             mcpPort,
			Path:             mcpPath,
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
		fmt.Fprintln(out, "Submitting mcp creation request...")

		serverMessage, err := deployClient.CreateMcp(
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
	},
}

var mcpUpdateCmd = &cobra.Command{
	Use:   "update <mcp_name>",
	Short: "Update an mcp's spec",
	Long: `Partial update — only the fields whose flags you pass are changed; everything
else keeps its current value. port/path/image/memory/cpu/env/secret only apply
to internal mcps. Use --clear-env, --clear-secret, or --clear-headers to remove
those entirely. The type (internal/external) and, for external mcps, the
endpoint/catalog cannot change — delete and recreate instead.

Changing --credential, or switching --auth-type to "none", rotates the mcp's
Secret and restarts the mcp (if internal) and every agent currently attached
to it. Auth routing cannot change while agents are attached — detach them first.`,
	Example: `  iai mcps update my-tool --image-tag v2
  iai mcps update my-tool --memory 1G --cpu 500m
  iai mcps update acme --credential "$NEW_TOKEN"
  iai mcps update my-tool --clear-headers`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		cred, err := inputs.ResolveCredential(cmd.InOrStdin(), mcpCredential, mcpCredentialStdin)
		if err != nil {
			return err
		}

		patch, err := inputs.BuildMcpUpdatePatch(inputs.McpInput{
			Port:             mcpPort,
			Path:             mcpPath,
			ImageType:        mcpImageType,
			ImageRepository:  mcpImageRepository,
			ImageName:        mcpImageName,
			ImageTag:         mcpImageTag,
			Memory:           mcpMemory,
			CPU:              mcpCPU,
			EnvVars:          mcpEnvVars,
			SecretRefs:       mcpSecretRefs,
			AuthType:         mcpAuthType,
			Credential:       cred,
			AuthHeader:       mcpAuthHeader,
			AuthHeaderPrefix: mcpAuthHeaderPfx,
			Headers:          mcpHeaders,
		}, mcpClearEnv, mcpClearSecret, mcpClearHeaders, cmd.Flags().Changed)
		if err != nil {
			return err
		}
		if len(patch) == 0 {
			return fmt.Errorf("no fields to update; pass at least one flag")
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting mcp update request...")

		serverMessage, err := deployClient.PatchMcp(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			mcpName,
			patch,
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

var mcpDescribeCmd = &cobra.Command{
	Use:     "describe <mcp_name>",
	Aliases: []string{"desc"},
	Short:   "Show mcp details, verify state, and cached tools",
	Long: `Show the mcp's record (type, external URL, catalog origin) and its latest
verify result — a tool count, not the tool list itself (see 'iai mcps tools').`,
	Example: `  iai mcps describe my-tool
  iai mcps describe my-tool --json`,
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

		if mcpDescribeJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if mcpDescribeYAML {
			return output.PrintStructuredYAML(out, res)
		}
		return output.PrintMcpDetail(out, res)
	},
}

var mcpToolsCmd = &cobra.Command{
	Use:   "tools <mcp_name>",
	Short: "List an mcp's cached tools with descriptions",
	Long: `Show the full cached tool list — name, description, and arguments. 'iai mcps
describe' only shows a count; use this to see the tools themselves.`,
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
		return output.PrintMcpTools(
			out, res.Tools,
			res.ToolsAdded, res.ToolsRemoved, res.ChangedFromRevision,
		)
	},
}

var mcpRevisionsCmd = &cobra.Command{
	Use:     "revisions <mcp_name>",
	Aliases: []string{"revs"},
	Short:   "List revisions of an mcp",
	Long: `Show past revisions of an mcp, sorted newest-first. Up to 50 revisions are
retained per mcp. Every spec change — update, credential rotation, agent
attach/detach — creates a revision.`,
	Example: `  iai mcps revisions my-tool`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		revisions, err := deployClient.ListMcpRevisions(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName,
		)
		if err != nil {
			return err
		}
		return output.PrintServiceRevisions(out, revisions)
	},
}

var mcpDiffCmd = &cobra.Command{
	Use:   "diff <mcp_name> <revision_a> <revision_b>",
	Short: "Compare two revisions of an mcp",
	Long: `Show the config differences between two revisions of an mcp — spec only.
Cached tools change at verify time, not per revision; 'iai mcps tools' shows
what changed since the previous verify.`,
	Example: `  iai mcps diff my-tool 1 3`,
	Args:    cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		revA, err := inputs.ParseRevisionArg(args[1])
		if err != nil {
			return err
		}
		revB, err := inputs.ParseRevisionArg(args[2])
		if err != nil {
			return err
		}

		pCtx, _, deployClient, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		a, err := deployClient.DescribeMcpRevision(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName, revA,
		)
		if err != nil {
			return err
		}
		b, err := deployClient.DescribeMcpRevision(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName, revB,
		)
		if err != nil {
			return err
		}
		return output.PrintRevisionDiff(out, args[1], a, args[2], b)
	},
}

var mcpVerifyCmd = &cobra.Command{
	Use:   "verify <mcp_name>",
	Short: "Re-verify an external mcp and refresh its cached tools",
	Long: `Re-dial the mcp (initialize + list tools) and refresh the cached tool list.
External mcps only — internal mcps verify automatically once their status is
healthy (background reconciler; see 'iai mcps describe') and reject a manual verify.`,
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
omit both to send an empty object. If the tool itself returns an error, it is
reported and the command exits non-zero.`,
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
		if res.Error != nil {
			return fmt.Errorf(
				"mcp %q tool %q returned an error (JSON-RPC %d): %s",
				mcpName, tool, res.Error.Code, res.Error.Message,
			)
		}
		return output.PrintRawJSON(out, res.Result)
	},
}

// confirmDeletion tolerates io.EOF so input without a trailing newline (echo -n y) still counts.
func confirmDeletion(in io.Reader, out io.Writer, target string) (bool, error) {
	fmt.Fprintf(out, "This will delete %s. Continue? [y/N] ", target)
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
			confirmed, err := confirmDeletion(cmd.InOrStdin(), out, fmt.Sprintf("mcp %q", mcpName))
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

	// Flags shared by create and update; type/catalog-id/external-url are create-only (see below) since they're identity, not patchable.
	for _, c := range []*cobra.Command{mcpCreateCmd, mcpUpdateCmd} {
		c.Flags().IntVar(&mcpPort, "port", 0, "Port the mcp server listens on (internal)")
		c.Flags().
			StringVar(&mcpPath, "path", "", `Endpoint path the mcp's own server exposes (internal, default "/mcp") — set to whatever the mcp owner actually configured, don't assume`)
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
			StringArrayVar(&mcpSecretRefs, "secret", nil, "Existing secret to load as env vars; can be repeated (internal)")
		c.Flags().
			StringVar(&mcpAuthType, "auth-type", "", `How the credential is sent: "bearer", "api_key", "custom", or "none" (inferred: "custom" if --auth-header/--auth-header-prefix is set, else "bearer" if --credential is set, else "none")`)
		c.Flags().
			StringVar(&mcpCredential, "credential", "", "Credential the mcp server requires (bearer token, API key)")
		c.Flags().
			BoolVar(&mcpCredentialStdin, "credential-stdin", false, "Read the credential from stdin instead of --credential")
		c.Flags().
			StringVar(&mcpAuthHeader, "auth-header", "", `Header the credential is sent in — only valid with --auth-type custom (bearer/api_key/none each imply their own)`)
		c.Flags().
			StringVar(&mcpAuthHeaderPfx, "auth-header-prefix", "", `Credential value prefix — only valid with --auth-type custom`)
		c.Flags().
			StringArrayVar(&mcpHeaders, "header", nil, "Extra non-secret request header (NAME=VALUE); can be repeated")
		c.MarkFlagsMutuallyExclusive("credential", "credential-stdin")
	}

	mcpCreateCmd.Flags().
		StringVar(&mcpType, "type", "", `Mcp type: "internal" or "external" (inferred from other flags if omitted)`)
	mcpCreateCmd.Flags().
		StringVar(&mcpEndpointURL, "external-url", "", "External MCP server URL — not platform-owned, dialed directly (custom external mcp)")
	mcpCreateCmd.Flags().
		StringVar(&mcpCatalogID, "catalog-id", "", "Catalog entry id (see 'iai mcps catalog'); derives endpoint + auth (catalog external mcp)")
	mcpCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "external-url")
	mcpCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "image-name")
	mcpCreateCmd.MarkFlagsMutuallyExclusive("external-url", "image-name")
	// Catalog entries carry their own auth routing — these only apply to custom endpoints.
	mcpCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "auth-header")
	mcpCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "auth-header-prefix")
	mcpCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "header")

	mcpUpdateCmd.Flags().
		BoolVar(&mcpClearEnv, "clear-env", false, "Remove all environment variables from the mcp")
	mcpUpdateCmd.Flags().
		BoolVar(&mcpClearSecret, "clear-secret", false, "Remove all secret references from the mcp")
	mcpUpdateCmd.Flags().
		BoolVar(&mcpClearHeaders, "clear-headers", false, "Remove all extra request headers from the mcp")

	mcpRunToolCmd.Flags().
		StringVar(&mcpArgsJSON, "args", "", "Tool arguments as an inline JSON object")
	mcpRunToolCmd.Flags().
		StringVar(&mcpArgsFile, "args-file", "", "Path to a file containing the tool arguments as a JSON object")
	mcpRunToolCmd.MarkFlagsMutuallyExclusive("args", "args-file")

	mcpDeleteCmd.Flags().BoolVarP(&mcpForce, "force", "f", false, "Skip confirmation prompt")

	mcpListCmd.Flags().BoolVar(&mcpListJSON, "json", false, "Output raw API response as JSON")
	mcpListCmd.Flags().BoolVar(&mcpListYAML, "yaml", false, "Output raw API response as YAML")
	mcpListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	mcpDescribeCmd.Flags().
		BoolVar(&mcpDescribeJSON, "json", false, "Output raw API response as JSON")
	mcpDescribeCmd.Flags().
		BoolVar(&mcpDescribeYAML, "yaml", false, "Output raw API response as YAML")
	mcpDescribeCmd.MarkFlagsMutuallyExclusive("json", "yaml")
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
		mcpDescribeCmd,
		mcpToolsCmd,
		mcpRevisionsCmd,
		mcpDiffCmd,
		mcpVerifyCmd,
		mcpRunToolCmd,
		mcpDeleteCmd,
	)
}
