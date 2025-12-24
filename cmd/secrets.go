package cmd

import (
	"encoding/base64"
	"fmt"
	"maps"
	"sort"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	secretsProject      string
	secretsOrganization string
	secretName          string
	secretDataKVs       []string
	secretEnvFile       string
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

The project is selected with --project.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cfg := &files.StackConfig{}
		var err error
		if cfgFilePath != "" {
			cfg, err = files.LoadStackConfig(cfgFilePath)
		}
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

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, secretsOrganization, selectedOrg)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := files.ResolveProject(cfg.Project, secretsProject)
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
				strings.Join(s.Keys, ", "),
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

The project is selected with --project.

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

		cfg := &files.StackConfig{}
		var err error
		if cfgFilePath != "" {
			cfg, err = files.LoadStackConfig(cfgFilePath)
		}
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

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, secretsOrganization, selectedOrg)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := files.ResolveProject(cfg.Project, secretsProject)
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
	Use:   "update [secret_name]",
	Short: "Update a secret in a project",
	Long: `Update a secret in a specific project using the deployment service.

The project is selected with --project.

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

		cfg := &files.StackConfig{}
		var err error
		if cfgFilePath != "" {
			cfg, err = files.LoadStackConfig(cfgFilePath)
		}
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

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, secretsOrganization, selectedOrg)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := files.ResolveProject(cfg.Project, secretsProject)
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
		fmt.Fprintln(out, "Submitting secret update request...")

		serverMessage, err := deployClient.UpdateSecret(cmd.Context(), orgId, projectId, secretName, data)
		if err != nil {
			return err
		}

		if serverMessage != "" {
			fmt.Fprintln(out, serverMessage)
		}

		return nil
	},
}

var secretsDeleteCmd = &cobra.Command{
	Use:     "delete <secret_name>",
	Aliases: []string{"rm"},
	Short:   "Delete a secret in a project",
	Long: `Delete a secret in a specific project using the deployment service.

The project is selected with --project.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		secretToDelete := strings.TrimSpace(args[0])
		if secretToDelete == "" {
			return fmt.Errorf("secret name is required")
		}

		cfg := &files.StackConfig{}
		var err error
		if cfgFilePath != "" {
			cfg, err = files.LoadStackConfig(cfgFilePath)
		}
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

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, secretsOrganization, selectedOrg)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := files.ResolveProject(cfg.Project, secretsProject)
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

The project is selected with --project.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		secretName := strings.TrimSpace(args[0])
		if secretName == "" {
			return fmt.Errorf("secret name is required")
		}

		cfg := &files.StackConfig{}
		var err error
		if cfgFilePath != "" {
			cfg, err = files.LoadStackConfig(cfgFilePath)
		}
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

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		orgName, err := files.ResolveOrganization(cfg.Organization, secretsOrganization, selectedOrg)
		if err != nil {
			return fmt.Errorf("failed to resolve organization: %w", err)
		}

		projectName, err := files.ResolveProject(cfg.Project, secretsProject)
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

		var keysDisplay string
		if len(secret.Data) > 0 {
			keys := make([]string, 0, len(secret.Data))
			for k := range secret.Data {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var pairs []string
			for _, k := range keys {
				val := secret.Data[k]
				if decoded, err := base64.StdEncoding.DecodeString(val); err == nil {
					val = string(decoded)
				}
				pairs = append(pairs, fmt.Sprintf("%s=%s", k, val))
			}
			keysDisplay = strings.Join(pairs, ", ")
		} else {
			keysDisplay = strings.Join(secret.Keys, ", ")
		}

		headers := []string{"NAME", "TYPE", "CREATED", "KEYS"}
		rows := [][]string{
			{
				secret.Name,
				secret.Type,
				secret.CreatedAt,
				keysDisplay,
			},
		}

		if err := output.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
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
	secretsUpdateCmd.Flags().StringVarP(&secretName, "secret-name", "s", "", "Name of the secret")
	secretsUpdateCmd.Flags().StringArrayVarP(&secretDataKVs, "data", "d", nil, "Secret data in KEY=VALUE form (repeatable)")
	secretsUpdateCmd.Flags().StringVar(&secretEnvFile, "from-env-file", "", "Path to env file with KEY=VALUE pairs (one per line)")

	// secrets delete
	secretsDeleteCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsDeleteCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")

	// secrets get
	secretsGetCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsGetCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")

	secretsCmd.AddCommand(secretsListCmd, secretsCreateCmd, secretsUpdateCmd, secretsDeleteCmd, secretsGetCmd)
	rootCmd.AddCommand(secretsCmd)
}
