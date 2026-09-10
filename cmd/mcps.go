package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/auth"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	mcpProject      string
	mcpOrganization string
	mcpDescription  string
)

var (
	mcpType            string
	mcpPort            int
	mcpPath            string
	mcpImageName       string
	mcpImageTag        string
	mcpMemory          string
	mcpCPU             string
	mcpEndpointURL     string
	mcpCatalogID       string
	mcpAuthType        string
	mcpCredential      string
	mcpCredentialStdin bool
	mcpAuthHeader      string
	mcpAuthHeaderPfx   string
	mcpStackId         string
)

var mcpForce bool

var (
	mcpArgsJSON string
	mcpArgsFile string
)

var (
	mcpListJSON         bool
	mcpListYAML         bool
	mcpDescribeJSON     bool
	mcpDescribeYAML     bool
	mcpCatalogJSON      bool
	mcpCatalogYAML      bool
	mcpConnectNoBrowser bool
	mcpVerifyJSON       bool
	mcpVerifyYAML       bool
	mcpToolsJSON        bool
	mcpToolsYAML        bool
)

var mcpsCmd = &cobra.Command{
	Use:     "mcps",
	Aliases: []string{"mcp"},
	Short:   "Deploy and manage MCP servers",
	GroupID: groupInfra,
	Long: `Manage MCP servers for a project — hosted servers ("internal"), custom
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
	Long: `Create an mcp — a hosted MCP server ("internal"), a custom external URL,
or a catalog-backed provider.

Internal: --image-name and --image-tag identify the image. --port, --path,
--memory, and --cpu configure how it runs.
External custom: --external-url — a server not owned by the platform, dialed
directly at that URL, path included.
External catalog: --catalog-id (see 'iai mcps catalog'); external URL and auth are
derived from the catalog entry, which provides its own credential header and
prefix. The entry decides the auth type — omit --auth-type unless it accepts
more than one, in which case the error names the options.

The mcp is verified against the live server before it's kept: an internal mcp
is verified automatically once ready; an external mcp (custom or catalog) is verified immediately,
and the create fails if the server is unreachable. Verification lists the
server's tools, so it only catches a bad credential on providers that require
auth to list them — some serve tool discovery anonymously.
An --auth-type oauth mcp is the exception: there is no credential until the
user signs in, so it is created unverified and reports no tools until then.`,
	Example: `  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m
  iai mcps create my-tool --image-name my-mcp-server --image-tag v1 --port 8080 --memory 512M --cpu 250m --path /api/mcp
  iai mcps create acme --external-url https://mcp.acme.com/mcp --credential "$ACME_TOKEN"
  iai mcps create github --catalog-id github --credential "$GITHUB_TOKEN"
  iai mcps create github --catalog-id github --credential-stdin < token.txt
  iai mcps create notion --catalog-id notion
  iai mcps create newrelic --catalog-id newrelic --auth-type oauth`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		backend := platform.McpBackendInternal
		if mcpCatalogID != "" || mcpEndpointURL != "" {
			backend = platform.McpBackendExternal
		}
		if mcpType != "" {
			backend = platform.McpBackend(mcpType)
		}
		if err := validateMcpBackendFlags(cmd, backend); err != nil {
			return err
		}

		cred, err := inputs.ResolveCredential(cmd.InOrStdin(), mcpCredential, mcpCredentialStdin)
		if err != nil {
			return err
		}
		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		if mcpCatalogID != "" {
			entry, catErr := apiClient.McpCatalogEntry(
				cmd.Context(), pCtx.orgId, pCtx.projectId, mcpCatalogID,
			)
			if catErr != nil {
				return catErr
			}
			mcpAuthType, err = catalogAuthType(entry, mcpAuthType)
			if err != nil {
				return err
			}
		}

		endpointURL := mcpEndpointURL
		if mcpCatalogID != "" {
			endpointURL = ""
		}
		var workload *platform.McpWorkload
		if backend == platform.McpBackendInternal {
			if mcpImageName == "" || mcpImageTag == "" {
				return fmt.Errorf("internal mcp requires --image-name and --image-tag")
			}
			port := mcpPort
			if !cmd.Flags().Changed("port") {
				port = 3000
			}
			path := mcpPath
			if !cmd.Flags().Changed("path") {
				path = "/mcp"
			}
			memory := mcpMemory
			if !cmd.Flags().Changed("memory") {
				memory = "128M"
			}
			cpu := mcpCPU
			if !cmd.Flags().Changed("cpu") {
				cpu = "100m"
			}
			workload = &platform.McpWorkload{
				Image:   mcpImageName + ":" + mcpImageTag,
				Port:    port,
				Path:    path,
				Memory:  memory,
				CPU:     cpu,
				StackId: mcpStackId,
			}
		}

		authType := mcpAuthTypeOr(backend, mcpAuthType, cred, mcpAuthHeader, mcpAuthHeaderPfx)
		auth := platform.McpAuth{
			Type:         authType,
			HeaderName:   utils.NilIfZero(mcpAuthHeader),
			HeaderPrefix: utils.NilIfZero(mcpAuthHeaderPfx),
		}
		if cred != "" || cmd.Flags().Changed("credential") || mcpCredentialStdin {
			auth.Credential = &cred
		}
		res, _, err := apiClient.CreateMcp(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			platform.McpCreateRequest{
				Name:        mcpName,
				Backend:     backend,
				Description: utils.NilIfZero(mcpDescription),
				CatalogID:   utils.NilIfZero(mcpCatalogID),
				EndpointURL: utils.NilIfZero(endpointURL),
				Transport:   "streamable_http",
				Auth:        auth,
				Workload:    workload,
			},
		)
		if err != nil {
			return err
		}

		if authType == "oauth" {
			fmt.Fprintf(
				out,
				"Created %s — it needs a sign-in before it can be used.\n  iai mcps connect %s\n",
				mcpName,
				mcpName,
			)
		} else {
			fmt.Fprintf(out, "Created %s — %s\n", mcpName, res.Backend)
		}
		return nil
	},
}

func mcpAuthTypeOr(
	backend platform.McpBackend,
	explicit, credential, headerName, headerPrefix string,
) string {
	if explicit != "" {
		return explicit
	}
	if backend == platform.McpBackendExternal &&
		(credential != "" || headerName != "" || headerPrefix != "") {
		return "bearer"
	}
	return "none"
}

var mcpUpdateCmd = &cobra.Command{
	Use:   "update <mcp_name>",
	Short: "Update an mcp's spec",
	Long: `Partial update — only the fields whose flags you pass are changed; everything
else keeps its current value. The type (internal/external) and, for external
mcps, the endpoint/catalog cannot change — delete and recreate instead.

Internal workload flags can be updated independently, except --image-name and
--image-tag, which must be passed together. Changing authentication restarts an
internal mcp and every attached agent. Detach agents before changing auth.`,
	Example: `  iai mcps update my-tool --image-name my-mcp --image-tag v2
  iai mcps update my-tool --memory 1G --cpu 500m
  iai mcps update acme --auth-type bearer --credential "$NEW_TOKEN"
  iai mcps update acme --description "notes for the team"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		if err := validateMcpUpdateFlags(cmd); err != nil {
			return err
		}

		cred, err := inputs.ResolveCredential(cmd.InOrStdin(), mcpCredential, mcpCredentialStdin)
		if err != nil {
			return err
		}

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}
		current, _, err := apiClient.DescribeMcp(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName,
		)
		if err != nil {
			return err
		}
		if err := validateMcpBackendFlags(cmd, current.Backend); err != nil {
			return err
		}

		patch := platform.McpUpdateRequest{}
		if cmd.Flags().Changed("description") {
			patch["description"] = mcpDescription
		}
		if cmd.Flags().Changed("auth-type") {
			auth := platform.McpAuth{Type: mcpAuthType}
			if cmd.Flags().Changed("credential") || mcpCredentialStdin {
				auth.Credential = &cred
			}
			if cmd.Flags().Changed("auth-header") {
				auth.HeaderName = &mcpAuthHeader
			}
			if cmd.Flags().Changed("auth-header-prefix") {
				auth.HeaderPrefix = &mcpAuthHeaderPfx
			}
			patch["auth"] = auth
		}
		workload := map[string]any{}
		if cmd.Flags().Changed("image-name") {
			workload["image"] = mcpImageName + ":" + mcpImageTag
		}
		if cmd.Flags().Changed("port") {
			workload["port"] = mcpPort
		}
		if cmd.Flags().Changed("path") {
			workload["path"] = mcpPath
		}
		if cmd.Flags().Changed("memory") {
			workload["memory"] = mcpMemory
		}
		if cmd.Flags().Changed("cpu") {
			workload["cpu"] = mcpCPU
		}
		if cmd.Flags().Changed("stack-id") {
			workload["stack_id"] = mcpStackId
		}
		if len(workload) > 0 {
			patch["workload"] = workload
		}
		if len(patch) == 0 {
			return fmt.Errorf("no fields to update; pass at least one flag")
		}

		res, _, err := apiClient.UpdateMcp(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			mcpName,
			patch,
		)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Updated %s — %s\n", mcpName, res.Backend)
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

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, raw, err := apiClient.ListMcps(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}

		if mcpListJSON {
			return output.PrintRawJSON(out, raw)
		}
		if mcpListYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintMcpList(out, res.Mcps)
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

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, raw, err := apiClient.DescribeMcp(cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName)
		if err != nil {
			return err
		}

		if mcpDescribeJSON {
			return output.PrintRawJSON(out, raw)
		}
		if mcpDescribeYAML {
			return output.PrintRawYAML(out, raw)
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

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, raw, err := apiClient.ListMcpTools(cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName)
		if err != nil {
			return err
		}
		if mcpToolsJSON {
			return output.PrintRawJSON(out, raw)
		}
		if mcpToolsYAML {
			return output.PrintRawYAML(out, raw)
		}
		return output.PrintMcpTools(out, res.Backend, res.Tools)
	},
}

var mcpRevisionsCmd = &cobra.Command{
	Use:     "revisions <mcp_name>",
	Aliases: []string{"revs"},
	Short:   "List revisions of an mcp",
	Long: `Show past revisions of an mcp, sorted newest-first. Up to 50 revisions are
retained per mcp. Every spec change — update, credential rotation, agent
attach/detach — creates a revision. Server-recorded actor and source metadata is
shown when available.`,
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
		return output.PrintRevisions(out, revisions)
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

var mcpConnectCmd = &cobra.Command{
	Use:   "connect <mcp_name>",
	Short: "Sign in to an mcp that authenticates with your account",
	Long: `Open a browser and approve access, so the mcp can be used.

Your access token is kept and renewed for you, so this is normally needed once.
Run it again if access is revoked at the provider, or to approve different
permissions.

The account you sign in with is shared — every agent using this mcp acts as you,
for everyone in the project. The provider's audit log shows your name, and the
connection stops working if your access does.`,
	Example: `  iai mcps connect notion-demo
  iai mcps connect notion-demo --no-browser`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		started, _, err := apiClient.BeginMcpSignIn(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			mcpName,
		)
		if err != nil {
			return err
		}
		if started.AuthorizeURL == "" {
			return fmt.Errorf("sign-in response did not include a URL")
		}
		if mcpConnectNoBrowser {
			fmt.Fprintf(out, "Open this to sign in:\n  %s\n\n", started.AuthorizeURL)
		} else {
			fmt.Fprintln(out, "Opening your browser to approve access...")
			if err := auth.OpenBrowser(started.AuthorizeURL); err != nil {
				fmt.Fprintf(out, "The browser did not open: %v\n", err)
			}
			fmt.Fprintf(out, "If it did not open, visit:\n  %s\n\n", started.AuthorizeURL)
		}

		connected, err := waitForSignIn(cmd.Context(), apiClient, pCtx, mcpName)
		if err != nil {
			return err
		}
		if !connected {
			return fmt.Errorf(
				"timed out waiting for approval — run 'iai mcps connect %s' to try again",
				mcpName,
			)
		}

		res, _, err := apiClient.DescribeMcp(cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Connected %s — %d tool(s) available.\n", mcpName, res.ToolCount)
		return nil
	},
}

func catalogAuthType(entry *platform.McpCatalogEntry, explicit string) (string, error) {
	if len(entry.AuthMethods) == 0 {
		return "", fmt.Errorf(
			"%q is not available to connect yet — choose a different provider",
			entry.ID,
		)
	}
	if explicit != "" || len(entry.AuthMethods) > 1 {
		return explicit, nil
	}
	return entry.AuthMethods[0], nil
}

func validateMcpBackendFlags(cmd *cobra.Command, backend platform.McpBackend) error {
	flags := []string{"auth-header", "auth-header-prefix"}
	appliesTo := "external"
	if backend == platform.McpBackendExternal {
		flags = []string{"image-name", "image-tag", "port", "path", "memory", "cpu", "stack-id"}
		appliesTo = "internal"
	}
	for _, name := range flags {
		if cmd.Flags().Changed(name) {
			return fmt.Errorf("--%s only applies to an %s mcp", name, appliesTo)
		}
	}
	return nil
}

func validateMcpUpdateFlags(cmd *cobra.Command) error {
	changed := false
	for _, name := range []string{
		"description", "image-name", "image-tag", "port", "path", "memory", "cpu",
		"stack-id", "auth-type", "credential", "credential-stdin", "auth-header",
		"auth-header-prefix",
	} {
		changed = changed || cmd.Flags().Changed(name)
	}
	if !changed {
		return fmt.Errorf("no fields to update; pass at least one flag")
	}
	for _, name := range []string{"image-name", "image-tag", "path", "memory", "cpu", "stack-id"} {
		if cmd.Flags().Changed(name) {
			value, err := cmd.Flags().GetString(name)
			if err != nil {
				return err
			}
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("--%s must not be empty", name)
			}
		}
	}
	nameChanged := cmd.Flags().Changed("image-name")
	tagChanged := cmd.Flags().Changed("image-tag")
	if nameChanged != tagChanged {
		return fmt.Errorf("--image-name and --image-tag must be passed together")
	}
	for _, name := range []string{"credential", "credential-stdin", "auth-header", "auth-header-prefix"} {
		if cmd.Flags().Changed(name) && !cmd.Flags().Changed("auth-type") {
			return fmt.Errorf("--%s requires --auth-type", name)
		}
	}
	return nil
}

var mcpDisconnectCmd = &cobra.Command{
	Use:   "disconnect <mcp_name>",
	Short: "Forget an mcp's stored provider credential",
	Long: `Remove the provider credential this mcp holds, so it can no longer be used
until someone signs in again.

Every agent using this mcp loses access, for everyone in the project. The mcp
itself is kept — use 'iai mcps delete' to remove that.`,
	Example: `  iai mcps disconnect notion-demo`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}
		if err := apiClient.Disconnect(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName,
		); err != nil {
			return err
		}
		fmt.Fprintf(
			out,
			"Disconnected %s — sign in again with 'iai mcps connect %s'.\n",
			mcpName, mcpName,
		)
		return nil
	},
}

func waitForSignIn(
	ctx context.Context, apiClient *platform.APIClient, pCtx *projectContext, mcpName string,
) (bool, error) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for i := 0; i < 100; i++ {
		connected, err := apiClient.ConnectionStatus(ctx, pCtx.orgId, pCtx.projectId, mcpName)
		if err != nil || connected {
			return connected, err
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
		}
	}
	return false, nil
}

var mcpVerifyCmd = &cobra.Command{
	Use:   "verify <mcp_name>",
	Short: "Re-verify an external mcp and refresh its cached tools",
	Long: `Re-dial the mcp (initialize + list tools) and refresh the cached tool list.
External mcps only — internal mcps verify automatically once ready and reject a
manual verify.`,
	Example: `  iai mcps verify my-tool
  iai mcps verify my-tool --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		mcpName := strings.TrimSpace(args[0])

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, raw, err := apiClient.VerifyMcp(cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName)
		if err != nil {
			return err
		}

		if mcpVerifyJSON {
			return output.PrintRawJSON(out, raw)
		}
		if mcpVerifyYAML {
			return output.PrintRawYAML(out, raw)
		}
		if res.Status != "ok" {
			reason := res.Status
			if res.ErrorClass != nil {
				reason = *res.ErrorClass
			}
			return fmt.Errorf("mcp %q did not verify: %s", mcpName, reason)
		}
		fmt.Fprintf(out, "Verified — %d tool(s) discovered\n", res.ToolCount)
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

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, _, err := apiClient.RunMcpTool(
			cmd.Context(), pCtx.orgId, pCtx.projectId, mcpName, tool, toolArgs,
		)
		if err != nil {
			return err
		}
		if res.Status != "ok" {
			reason := res.Status
			if res.ErrorClass != nil {
				reason = *res.ErrorClass
			}
			return fmt.Errorf("mcp %q tool %q returned an error: %s", mcpName, tool, reason)
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
	Long: `Remove the mcp, its stored credential, and cached tools. The command is
rejected while agents are attached. -f only skips the confirmation prompt.

Detach it from any attached agent first with 'iai agents update <agent> --detach-mcp <mcp_name>'.

If you signed in to this mcp, the access you approved stays granted with the
provider. Revoke it in your account there if you want it withdrawn.`,
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

		pCtx, apiClient, _, err := resolveProject(cmd.Context(), mcpOrganization, mcpProject)
		if err != nil {
			return err
		}

		res, _, err := apiClient.DeleteMcp(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			mcpName,
		)
		if err != nil {
			return err
		}
		if !res.Deleted {
			return fmt.Errorf("mcp %q was not deleted", mcpName)
		}
		fmt.Fprintf(out, "Successfully deleted mcp %q.\n", mcpName)
		return nil
	},
}

func init() {
	mcpsCmd.PersistentFlags().
		StringVarP(&mcpProject, "project", "p", "", "Project name that owns the mcps")
	mcpsCmd.PersistentFlags().
		StringVarP(&mcpOrganization, "organization", "o", "", "Organization name that owns the project")

	for _, c := range []*cobra.Command{mcpCreateCmd, mcpUpdateCmd} {
		c.Flags().IntVar(&mcpPort, "port", 0, "Port the mcp server listens on (internal)")
		c.Flags().
			StringVar(&mcpPath, "path", "", `Endpoint path the mcp's own server exposes (internal, default "/mcp")`)
		c.Flags().StringVar(&mcpImageName, "image-name", "", "Container image name (internal)")
		c.Flags().StringVar(&mcpImageTag, "image-tag", "", "Container image tag (internal)")
		c.Flags().StringVar(&mcpMemory, "memory", "", "Memory request/limit, e.g. 512M (internal)")
		c.Flags().StringVar(&mcpCPU, "cpu", "", "CPU request/limit, e.g. 250m (internal)")
		c.Flags().
			StringVar(&mcpAuthType, "auth-type", "", `How the credential is sent: "bearer", "api_key", "none", or "oauth" (inferred on create; required when changing authentication)`)
		c.Flags().
			StringVar(&mcpCredential, "credential", "", "Credential the mcp server requires (bearer token, API key)")
		c.Flags().
			BoolVar(&mcpCredentialStdin, "credential-stdin", false, "Read the credential from stdin instead of --credential")
		c.Flags().
			StringVar(&mcpAuthHeader, "auth-header", "", "Header used to send the credential")
		c.Flags().
			StringVar(&mcpAuthHeaderPfx, "auth-header-prefix", "", "Credential value prefix")
		c.Flags().StringVar(&mcpStackId, "stack-id", "", "Stack ID to assign the mcp to (internal)")
		c.Flags().
			StringVar(&mcpDescription, "description", "", "Human-readable description of the mcp")
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
	mcpCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "auth-header")
	mcpCreateCmd.MarkFlagsMutuallyExclusive("catalog-id", "auth-header-prefix")

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
	mcpConnectCmd.Flags().BoolVar(&mcpConnectNoBrowser, "no-browser", false,
		"Print the sign-in URL instead of opening it")

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
		mcpConnectCmd,
		mcpDisconnectCmd,
		mcpVerifyCmd,
		mcpRunToolCmd,
		mcpDeleteCmd,
	)
}
