package cmd

import (
	"fmt"
	"strings"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

var (
	secretsProject      string
	secretsOrganization string
	secretName          string
	secretDataKVs       []string
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

		if strings.TrimSpace(secretsProject) == "" {
			return fmt.Errorf("project is required; please provide --project")
		}

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := internal.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := internal.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if strings.TrimSpace(secretsOrganization) == "" {
			if strings.TrimSpace(selectedOrg) == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			secretsOrganization = selectedOrg
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), secretsOrganization, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", secretsProject, err)
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

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var secretsCreateCmd = &cobra.Command{
	Use:   "create [secret_name]",
	Short: "Create a secret in a project",
	Long: `Create a secret in a specific project using the deployment service.

The project is selected with --project.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 && strings.TrimSpace(secretName) == "" {
			secretName = args[0]
		}

		if strings.TrimSpace(secretsProject) == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if strings.TrimSpace(secretName) == "" {
			return fmt.Errorf("secret name is required; please provide --secret-name or positional argument")
		}
		if len(secretDataKVs) == 0 {
			return fmt.Errorf("at least one --data KEY=VALUE pair is required")
		}

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := internal.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := internal.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if strings.TrimSpace(secretsOrganization) == "" {
			if strings.TrimSpace(selectedOrg) == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			secretsOrganization = selectedOrg
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), secretsOrganization, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", secretsProject, err)
		}

		data, err := buildSecretData(secretDataKVs)
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

The project is selected with --project.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if len(args) > 0 && strings.TrimSpace(secretName) == "" {
			secretName = args[0]
		}

		if strings.TrimSpace(secretsProject) == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if strings.TrimSpace(secretName) == "" {
			return fmt.Errorf("secret name is required; please provide --secret-name or positional argument")
		}
		if len(secretDataKVs) == 0 {
			return fmt.Errorf("at least one --data KEY=VALUE pair is required")
		}

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := internal.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := internal.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if strings.TrimSpace(secretsOrganization) == "" {
			if strings.TrimSpace(selectedOrg) == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			secretsOrganization = selectedOrg
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), secretsOrganization, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", secretsProject, err)
		}

		data, err := buildSecretData(secretDataKVs)
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
		if strings.TrimSpace(secretsProject) == "" {
			return fmt.Errorf("project is required; please provide --project")
		}

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := internal.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		deployClient, err := internal.NewDeploymentClient(deploymentHostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if strings.TrimSpace(secretsOrganization) == "" {
			if strings.TrimSpace(selectedOrg) == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			secretsOrganization = selectedOrg
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), secretsOrganization, secretsProject)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", secretsProject, err)
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

func init() {
	// secrets list
	secretsListCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsListCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")

	// secrets create
	secretsCreateCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsCreateCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")
	secretsCreateCmd.Flags().StringVarP(&secretName, "secret-name", "s", "", "Name of the secret")
	secretsCreateCmd.Flags().StringArrayVarP(&secretDataKVs, "data", "d", nil, "Secret data in KEY=VALUE form (repeatable)")

	// secrets update
	secretsUpdateCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsUpdateCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")
	secretsUpdateCmd.Flags().StringVarP(&secretName, "secret-name", "s", "", "Name of the secret")
	secretsUpdateCmd.Flags().StringArrayVarP(&secretDataKVs, "data", "d", nil, "Secret data in KEY=VALUE form (repeatable)")

	// secrets delete
	secretsDeleteCmd.Flags().StringVarP(&secretsProject, "project", "p", "", "Project name that owns the secrets")
	secretsDeleteCmd.Flags().StringVarP(&secretsOrganization, "organization", "o", "", "Organization name that owns the project")

	secretsCmd.AddCommand(secretsListCmd, secretsCreateCmd, secretsUpdateCmd, secretsDeleteCmd)
	rootCmd.AddCommand(secretsCmd)
}
