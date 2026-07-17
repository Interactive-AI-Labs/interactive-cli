package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/auth"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	apiKeysProject      string
	apiKeysOrganization string
	apiKeysJSON         bool
	apiKeysYAML         bool
	apiKeysListColumns  []string
	apiKeyNote          string
	apiKeyUpdateNote    string

	routerKeysProject      string
	routerKeysOrganization string
	routerKeysJSON         bool
	routerKeysYAML         bool
	routerKeysListColumns  []string
	routerKeyDescription   string
	routerKeyLimit         float64
	routerKeyLimitReset    string
	routerKeyExpiresAt     string
	routerKeyClearLimit    bool
	routerKeyDisable       bool
	routerKeyEnable        bool
)

func requireKeyManagementAuth() error {
	if apiKey != "" {
		return auth.KeyManagementLoginRequiredError()
	}
	if token != "" {
		return nil
	}

	cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
	if err != nil {
		return fmt.Errorf("failed to load login session: %w", err)
	}
	if len(cookies) == 0 {
		return auth.KeyManagementLoginRequiredError()
	}

	return nil
}

var apiKeysCmd = &cobra.Command{
	Use:     "api-keys",
	Aliases: []string{"api-key"},
	Short:   "Project API keys",
	GroupID: groupAuth,
	Long: `Manage project API keys. Requires iai login or JWT authentication. API key authentication is not supported.

Project API keys authenticate platform/API access for reading and writing project context, such as prompts, routines, policies, variables, glossaries, macros, traces, scores, datasets, and for creating infrastructure resources such as agents, services, and databases.`,
}

var apiKeysListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List project API keys",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireKeyManagementAuth(); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			apiKeysOrganization,
			apiKeysProject,
		)
		if err != nil {
			return err
		}

		columns := apiKeysListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultProjectAPIKeyColumns
		}
		if apiKeysJSON || apiKeysYAML {
			columns = nil
		} else if err := inputs.ValidateColumns(columns, inputs.AllProjectAPIKeyColumns); err != nil {
			return err
		}

		keys, err := apiClient.ListProjectAPIKeys(cmd.Context(), pCtx.orgId, pCtx.projectId)
		if err != nil {
			return err
		}

		if apiKeysJSON {
			return output.PrintStructuredJSON(out, keys)
		}
		if apiKeysYAML {
			return output.PrintStructuredYAML(out, keys)
		}
		return output.PrintProjectAPIKeyList(out, keys, columns)
	},
}

var apiKeysCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a project API key",
	Long: `Create a project API key.

Project API keys authenticate platform/API access for reading and writing project context, such as prompts, routines, policies, variables, glossaries, macros, traces, scores, datasets, and for creating infrastructure resources such as agents, services, and databases.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireKeyManagementAuth(); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			apiKeysOrganization,
			apiKeysProject,
		)
		if err != nil {
			return err
		}

		key, err := apiClient.CreateProjectAPIKey(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			clients.CreateProjectAPIKeyBody{Note: apiKeyNote},
		)
		if err != nil {
			return err
		}

		if apiKeysJSON {
			return output.PrintStructuredJSON(out, key)
		}
		if apiKeysYAML {
			return output.PrintStructuredYAML(out, key)
		}
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Public key: %s\n", key.PublicKey)
		fmt.Fprintf(out, "Secret key: %s\n", key.SecretKey)
		fmt.Fprintln(out, "Save this secret now; it is only shown once.")
		return nil
	},
}

var apiKeysUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a project API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireKeyManagementAuth(); err != nil {
			return err
		}
		if !cmd.Flags().Changed("note") {
			return fmt.Errorf("at least one update flag is required")
		}

		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			apiKeysOrganization,
			apiKeysProject,
		)
		if err != nil {
			return err
		}

		res, err := apiClient.UpdateProjectAPIKey(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			args[0],
			clients.UpdateProjectAPIKeyBody{Note: apiKeyUpdateNote},
		)
		if err != nil {
			return err
		}
		if !res.Success {
			if res.Message != "" {
				return fmt.Errorf("update failed: %s", res.Message)
			}
			return fmt.Errorf("update failed")
		}

		if apiKeysJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if apiKeysYAML {
			return output.PrintStructuredYAML(out, res)
		}

		fmt.Fprintln(out, "Updated")
		return nil
	},
}

var apiKeysDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a project API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireKeyManagementAuth(); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			apiKeysOrganization,
			apiKeysProject,
		)
		if err != nil {
			return err
		}

		res, err := apiClient.DeleteProjectAPIKey(
			cmd.Context(),
			pCtx.orgId,
			pCtx.projectId,
			args[0],
		)
		if err != nil {
			return err
		}

		if !res.Success {
			if res.Message != "" {
				return fmt.Errorf("delete failed: %s", res.Message)
			}
			return fmt.Errorf("delete failed")
		}

		if apiKeysJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if apiKeysYAML {
			return output.PrintStructuredYAML(out, res)
		}

		fmt.Fprintln(out, "Deleted")
		return nil
	},
}

var routerKeysCmd = &cobra.Command{
	Use:     "keys",
	Aliases: []string{"key"},
	Short:   "Router API keys",
	Long: `Manage InteractiveAI Router API keys. Requires iai login or JWT authentication. API key authentication is not supported.

Router keys authenticate inference requests to the InteractiveAI Router, for example chat completions and model calls. They are used as bearer tokens for runtime inference, not for managing project context or infrastructure.`,
}

var routerKeysListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List router API keys",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireKeyManagementAuth(); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			routerKeysOrganization,
			routerKeysProject,
		)
		if err != nil {
			return err
		}

		columns := routerKeysListColumns
		if len(columns) == 0 {
			columns = inputs.DefaultRouterAPIKeyColumns
		}
		if routerKeysJSON || routerKeysYAML {
			columns = nil
		} else if err := inputs.ValidateColumns(columns, inputs.AllRouterAPIKeyColumns); err != nil {
			return err
		}

		res, err := apiClient.ListRouterAPIKeys(cmd.Context(), pCtx.projectId)
		if err != nil {
			return err
		}

		if routerKeysJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if routerKeysYAML {
			return output.PrintStructuredYAML(out, res)
		}
		return output.PrintRouterAPIKeyList(out, res.Keys, columns)
	},
}

var routerKeysCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a router API key",
	Long: `Create a router API key.

Router keys authenticate inference requests to the InteractiveAI Router, for example chat completions and model calls. They are used as bearer tokens for runtime inference, not for managing project context or infrastructure.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireKeyManagementAuth(); err != nil {
			return err
		}

		name := strings.TrimSpace(args[0])
		if name == "" {
			return fmt.Errorf("name is required")
		}
		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			routerKeysOrganization,
			routerKeysProject,
		)
		if err != nil {
			return err
		}

		res, err := apiClient.CreateRouterAPIKey(
			cmd.Context(),
			pCtx.projectId,
			clients.CreateRouterAPIKeyBody{
				Name:               name,
				Description:        routerKeyDescription,
				Limit:              routerKeyLimit,
				LimitReset:         routerKeyLimitReset,
				IncludeBYOKInLimit: false,
				ExpiresAt:          routerKeyExpiresAt,
			},
		)
		if err != nil {
			return err
		}

		if routerKeysJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if routerKeysYAML {
			return output.PrintStructuredYAML(out, res)
		}
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Router key: %s\n", res.Key)
		fmt.Fprintln(out, "Save this secret now; it is only shown once.")
		return nil
	},
}

var routerKeysUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a router API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireKeyManagementAuth(); err != nil {
			return err
		}
		if routerKeyClearLimit && cmd.Flags().Changed("limit") {
			return fmt.Errorf("--clear-limit cannot be used with --limit")
		}
		if routerKeyDisable && routerKeyEnable {
			return fmt.Errorf("--disable cannot be used with --enable")
		}

		patch := clients.UpdateRouterAPIKeyBody{}
		if cmd.Flags().Changed("limit") {
			patch["limit"] = routerKeyLimit
		}
		if routerKeyClearLimit {
			patch["limit"] = nil
		}
		if cmd.Flags().Changed("limit-reset") {
			if routerKeyLimitReset == "none" {
				patch["limit_reset"] = nil
			} else {
				patch["limit_reset"] = routerKeyLimitReset
			}
		}
		if routerKeyDisable {
			patch["disabled"] = true
		}
		if routerKeyEnable {
			patch["disabled"] = false
		}
		if len(patch) == 0 {
			return fmt.Errorf("at least one update flag is required")
		}

		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			routerKeysOrganization,
			routerKeysProject,
		)
		if err != nil {
			return err
		}

		key, err := apiClient.UpdateRouterAPIKey(cmd.Context(), pCtx.projectId, args[0], patch)
		if err != nil {
			return err
		}

		if routerKeysJSON {
			return output.PrintStructuredJSON(out, key)
		}
		if routerKeysYAML {
			return output.PrintStructuredYAML(out, key)
		}

		fmt.Fprintln(out, "Updated")
		return nil
	},
}

var routerKeysDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a router API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireKeyManagementAuth(); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		pCtx, apiClient, _, err := resolveProject(
			cmd.Context(),
			routerKeysOrganization,
			routerKeysProject,
		)
		if err != nil {
			return err
		}

		res, err := apiClient.DeleteRouterAPIKey(cmd.Context(), pCtx.projectId, args[0])
		if err != nil {
			return err
		}

		if !res.Success {
			if res.Message != "" {
				return fmt.Errorf("delete failed: %s", res.Message)
			}
			return fmt.Errorf("delete failed")
		}

		if routerKeysJSON {
			return output.PrintStructuredJSON(out, res)
		}
		if routerKeysYAML {
			return output.PrintStructuredYAML(out, res)
		}
		if res.Message != "" {
			fmt.Fprintln(out, res.Message)
			return nil
		}

		fmt.Fprintln(out, "Deleted")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(apiKeysCmd)

	for _, c := range []*cobra.Command{apiKeysListCmd, apiKeysCreateCmd, apiKeysUpdateCmd, apiKeysDeleteCmd} {
		c.Flags().StringVarP(&apiKeysProject, "project", "p", "", "Project name")
		c.Flags().StringVarP(&apiKeysOrganization, "organization", "o", "", "Organization name")
		c.Flags().BoolVar(&apiKeysJSON, "json", false, "Output response as JSON")
		c.Flags().BoolVar(&apiKeysYAML, "yaml", false, "Output response as YAML")
		c.MarkFlagsMutuallyExclusive("json", "yaml")
	}
	apiKeysListCmd.Flags().
		StringSliceVar(&apiKeysListColumns, "columns", nil, "Columns to display for table output only (comma-separated, default: id,public_key,secret,note,created_at). Cannot be used with --json or --yaml.\nAvailable: id,public_key,secret,note,status,expires_at,last_used_at,created_at")
	apiKeysListCmd.MarkFlagsMutuallyExclusive("columns", "json", "yaml")
	apiKeysCreateCmd.Flags().StringVar(&apiKeyNote, "note", "", "API key note")
	apiKeysUpdateCmd.Flags().StringVar(&apiKeyUpdateNote, "note", "", "API key note")
	apiKeysCmd.AddCommand(apiKeysListCmd, apiKeysCreateCmd, apiKeysUpdateCmd, apiKeysDeleteCmd)

	for _, c := range []*cobra.Command{routerKeysListCmd, routerKeysCreateCmd, routerKeysUpdateCmd, routerKeysDeleteCmd} {
		c.Flags().StringVarP(&routerKeysProject, "project", "p", "", "Project name")
		c.Flags().StringVarP(&routerKeysOrganization, "organization", "o", "", "Organization name")
		c.Flags().BoolVar(&routerKeysJSON, "json", false, "Output response as JSON")
		c.Flags().BoolVar(&routerKeysYAML, "yaml", false, "Output response as YAML")
		c.MarkFlagsMutuallyExclusive("json", "yaml")
	}
	routerKeysListCmd.Flags().
		StringSliceVar(&routerKeysListColumns, "columns", nil, "Columns to display for table output only (comma-separated, default: id,name,key,limit,created_at). Cannot be used with --json or --yaml.\nAvailable: id,name,description,status,key,disabled,limit,remaining,limit_reset,expires_at,last_used_at,created_at,updated_at,project_id,user_id")
	routerKeysListCmd.MarkFlagsMutuallyExclusive("columns", "json", "yaml")
	routerKeysCreateCmd.Flags().
		StringVar(&routerKeyDescription, "description", "", "Router key description")
	routerKeysCreateCmd.Flags().
		Float64Var(&routerKeyLimit, "limit", 0, "Credit limit in USD. If omitted, defaults to $100. Maximum is $2500.")
	routerKeysCreateCmd.Flags().
		StringVar(&routerKeyLimitReset, "limit-reset", "", "Limit reset period: daily, weekly, monthly. If omitted, defaults to monthly.")
	routerKeysCreateCmd.Flags().
		StringVar(&routerKeyExpiresAt, "expires-at", "", "Expiration timestamp (RFC3339). If omitted, keys do not expire by default.")
	routerKeysUpdateCmd.Flags().
		Float64Var(&routerKeyLimit, "limit", 0, "Credit limit in USD")
	routerKeysUpdateCmd.Flags().
		BoolVar(&routerKeyClearLimit, "clear-limit", false, "Remove the credit limit")
	routerKeysUpdateCmd.Flags().
		StringVar(&routerKeyLimitReset, "limit-reset", "", "Limit reset period: none, daily, weekly, monthly")
	routerKeysUpdateCmd.Flags().
		BoolVar(&routerKeyDisable, "disable", false, "Disable this key")
	routerKeysUpdateCmd.Flags().
		BoolVar(&routerKeyEnable, "enable", false, "Enable this key")
	routerKeysCmd.AddCommand(
		routerKeysListCmd,
		routerKeysCreateCmd,
		routerKeysUpdateCmd,
		routerKeysDeleteCmd,
	)
	routerCmd.AddCommand(routerKeysCmd)
}
