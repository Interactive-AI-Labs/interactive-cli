package cmd

import (
	"encoding/base64"
	"fmt"
	"maps"
	"sort"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
	"github.com/spf13/cobra"
)

var (
	secretsProject      string
	secretsOrganization string
	secretName          string
	secretDataKVs       []string
	secretEnvFile       string
	secretReplaceFlag   bool
)

var secretsCmd = &cobra.Command{
	Use:     "secrets",
	Aliases: []string{"secret"},
	Short:   "Manage secrets",
	Long:    `Manage secrets in InteractiveAI projects.`,
}

var secretsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List secrets in a project",
	Long: `List secrets in a specific project.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, secretsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		secrets, err := deployClient.ListSecrets(cmd.Context(), orgId, projectId)
		if err != nil {
			return err
		}

		if len(secrets) == 0 {
			fmt.Fprintln(out, "No secrets found.")
			return nil
		}

		headers := []string{"NAME", "TYPE", "CREATED", "KEYS"}
		rows := make([][]string, len(secrets))
		for i, s := range secrets {
			rows[i] = []string{
				s.Name,
				s.Type,
				s.CreatedAt,
				formatSecretKeys(s.Keys, 3),
			}
		}

		if err := output.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var secretsCreateCmd = &cobra.Command{
	Use:   "create [secret_name]",
	Short: "Create a secret in a project",
	Long: `Create a secret in a specific project using the deployment service.

The project is selected with --project or via 'iai projects select'.

Secret data can be provided via:
  --data KEY=VALUE         (can be repeated)
  --from-env-file FILE     (KEY=VALUE pairs, one per line)

When both are provided, --data values take precedence.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 && strings.TrimSpace(secretName) == "" {
			secretName = args[0]
		}

		if strings.TrimSpace(secretName) == "" {
			return fmt.Errorf("secret name is required; please provide --secret-name or positional argument")
		}
		if len(secretDataKVs) == 0 && strings.TrimSpace(secretEnvFile) == "" {
			return fmt.Errorf("at least one --data KEY=VALUE pair or --from-env-file is required")
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, secretsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		data, err := buildSecretDataWithEnvFile(secretDataKVs, secretEnvFile)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting secret creation request...")

		serverMessage, err := deployClient.CreateSecret(cmd.Context(), orgId, projectId, secretName, data)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var secretsUpdateCmd = &cobra.Command{
	Use:   "update <secret_name>",
	Short: "Update keys in a secret",
	Long: `Update one or more keys in an existing secret.

By default, only the specified keys are updated (merge/upsert). Existing keys
not included in the update are preserved.

With --replace, ALL secret data is replaced. Any keys not included in the new
data will be permanently deleted.

The project is selected with --project or via 'iai projects select'.

Secret data can be provided via:
  --data KEY=VALUE         (can be repeated)
  --from-env-file FILE     (KEY=VALUE pairs, one per line)

When both are provided, --data values take precedence.

Examples:
  # Update a single key (other keys preserved)
  iai secrets update my-secret -d API_KEY=new-value

  # Update multiple keys (other keys preserved)
  iai secrets update my-secret -d API_KEY=val1 -d DB_PASS=val2

  # Replace all keys (keys not provided will be deleted)
  iai secrets update my-secret -d API_KEY=val1 --replace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		secretName := strings.TrimSpace(args[0])
		if secretName == "" {
			return fmt.Errorf("secret name is required")
		}

		if len(secretDataKVs) == 0 && strings.TrimSpace(secretEnvFile) == "" {
			return fmt.Errorf("at least one --data KEY=VALUE pair or --from-env-file is required")
		}

		data, err := buildSecretDataWithValidation(secretDataKVs, secretEnvFile)
		if err != nil {
			return err
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, secretsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		fmt.Fprintln(out)

		if secretReplaceFlag {
			fmt.Fprintln(out, "Submitting secret replace request...")

			serverMessage, err := deployClient.ReplaceSecret(cmd.Context(), orgId, projectId, secretName, data)
			if err != nil {
				return err
			}

			if serverMessage != "" {
				fmt.Fprintln(out, serverMessage)
			}
			return nil
		}

		fmt.Fprintln(out, "Submitting secret update request...")

		for keyName, value := range data {
			serverMessage, err := deployClient.UpdateSecretKey(cmd.Context(), orgId, projectId, secretName, keyName, value)
			if err != nil {
				return fmt.Errorf("failed to update key %q: %w", keyName, err)
			}

			if serverMessage != "" {
				fmt.Fprintln(out, serverMessage)
			}
		}

		return nil
	},
}

var secretsDeleteCmd = &cobra.Command{
	Use:     "delete <secret_name>",
	Aliases: []string{"rm"},
	Short:   "Delete a secret in a project",
	Long: `Delete a secret in a specific project using the deployment service.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		secretToDelete := strings.TrimSpace(args[0])
		if secretToDelete == "" {
			return fmt.Errorf("secret name is required")
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, secretsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting secret delete request...")

		serverMessage, err := deployClient.DeleteSecret(cmd.Context(), orgId, projectId, secretToDelete)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var secretsGetCmd = &cobra.Command{
	Use:   "get <secret_name>",
	Short: "Get a secret in a project",
	Long: `Get a secret in a specific project using the deployment service.

The project is selected with --project or via 'iai projects select'.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		secretName := strings.TrimSpace(args[0])
		if secretName == "" {
			return fmt.Errorf("secret name is required")
		}

		cfg, err := files.LoadStackConfig(cfgFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create API client: %w", err)
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return fmt.Errorf("failed to create deployment client: %w", err)
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, secretsOrganization)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := sess.ResolveProject(cfg.Project, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		secret, err := deployClient.GetSecret(cmd.Context(), orgId, projectId, secretName)
		if err != nil {
			return fmt.Errorf("failed to get secret %q: %w", secretName, err)
		}

		headers := []string{"KEYS", "VALUES"}
		var rows [][]string

		if len(secret.Data) > 0 {
			keys := make([]string, 0, len(secret.Data))
			for k := range secret.Data {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				val := secret.Data[k]
				if decoded, err := base64.StdEncoding.DecodeString(val); err == nil {
					val = string(decoded)
				}
				rows = append(rows, []string{k, val})
			}
		} else {
			fmt.Fprintln(out, "No data found in secret.")
			return nil
		}

		if err := output.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}


func formatSecretKeys(keys []string, maxVisible int) string {
	if len(keys) == 0 {
		return ""
	}
	if len(keys) <= maxVisible {
		return strings.Join(keys, ", ")
	}
	visible := strings.Join(keys[:maxVisible], ", ")
	return fmt.Sprintf("%s (+%d more)", visible, len(keys)-maxVisible)
}

func buildSecretData(pairs []string) (map[string]string, error) {
	data := make(map[string]string, len(pairs))

	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --data value %q; expected KEY=VALUE", p)
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			return nil, fmt.Errorf("invalid --data value %q; key must not be empty", p)
		}

		value := parts[1]
		if err := inputs.ValidateSecretValue(key, value); err != nil {
			return nil, err
		}
		data[key] = value
	}

	return data, nil
}

func buildSecretDataWithEnvFile(pairs []string, envFilePath string) (map[string]string, error) {
	data := make(map[string]string)

	if strings.TrimSpace(envFilePath) != "" {
		envData, err := files.ParseEnvFile(envFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load env file: %w", err)
		}
		maps.Copy(data, envData)
	}

	if len(pairs) > 0 {
		pairData, err := buildSecretData(pairs)
		if err != nil {
			return nil, err
		}
		// We don't run maps.copy to avoid panicking with duplicated keys
		for k, v := range pairData {
			data[k] = v
		}
	}

	return data, nil
}

func buildSecretDataWithValidation(pairs []string, envFilePath string) (map[string]string, error) {
	data, err := buildSecretDataWithEnvFile(pairs, envFilePath)
	if err != nil {
		return nil, err
	}

	for key := range data {
		if err := inputs.ValidateSecretKey(key); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func init() {
	// secrets list
	secretsListCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsListCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")

	// secrets create
	secretsCreateCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsCreateCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")
	secretsCreateCmd.Flags().StringVarP(&secretName, "secret-name", "s", "", "Name of the secret")
	secretsCreateCmd.Flags().StringArrayVarP(&secretDataKVs, "data", "d", nil, "Secret data in KEY=VALUE form (repeatable)")
	secretsCreateCmd.Flags().StringVar(&secretEnvFile, "from-env-file", "", "Path to env file with KEY=VALUE pairs (one per line)")

	// secrets update
	secretsUpdateCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsUpdateCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")
	secretsUpdateCmd.Flags().StringArrayVarP(&secretDataKVs, "data", "d", nil, "Secret data in KEY=VALUE form (repeatable)")
	secretsUpdateCmd.Flags().StringVar(&secretEnvFile, "from-env-file", "", "Path to env file with KEY=VALUE pairs (one per line)")
	secretsUpdateCmd.Flags().BoolVar(&secretReplaceFlag, "replace", false, "Replace all secret data (keys not provided will be deleted)")

	// secrets delete
	secretsDeleteCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsDeleteCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")

	// secrets get
	secretsGetCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsGetCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")

	// Wire up the command hierarchy
	secretsCmd.AddCommand(secretsListCmd, secretsCreateCmd, secretsUpdateCmd, secretsDeleteCmd, secretsGetCmd)
	rootCmd.AddCommand(secretsCmd)
}
