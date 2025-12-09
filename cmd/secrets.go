package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

type SecretInfo struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	CreatedAt string   `json:"createdAt"`
	Keys      []string `json:"keys"`
}

type ListSecretsResponse struct {
	Secrets []SecretInfo `json:"secrets"`
}

type CreateSecretRequest struct {
	Data map[string]string `json:"data"`
}

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
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
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

		orgID, projectID, err := internal.GetProjectId(
			cmd.Context(),
			hostname,
			cfgDirName,
			sessionFileName,
			secretsOrganization,
			secretsProject,
			defaultHTTPTimeout,
		)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", secretsProject, err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/secrets", orgID, projectID)

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		for _, c := range cookies {
			if c != nil {
				req.AddCookie(c)
			}
		}

		client := &http.Client{Timeout: defaultHTTPTimeout}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("secrets request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := internal.ExtractServerMessage(respBody)
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("failed to list secrets: server returned %s", resp.Status)
		}

		var result ListSecretsResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("failed to decode secrets response: %w", err)
		}

		if len(result.Secrets) == 0 {
			fmt.Fprintln(out, "No secrets found.")
			return nil
		}

		headers := []string{"NAME", "TYPE", "CREATED", "KEYS"}
		rows := make([][]string, len(result.Secrets))
		for i, s := range result.Secrets {
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
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
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

		orgID, projectID, err := internal.GetProjectId(
			cmd.Context(),
			hostname,
			cfgDirName,
			sessionFileName,
			secretsOrganization,
			secretsProject,
			defaultHTTPTimeout,
		)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", secretsProject, err)
		}

		data, err := buildSecretData(secretDataKVs)
		if err != nil {
			return err
		}

		reqBody := CreateSecretRequest{
			Data: data,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/secrets/%s", orgID, projectID, secretName)

		req, err := internal.NewRequestWCookies(cmd.Context(), http.MethodPost, u.String(), bodyBytes, cookies)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		client := &http.Client{Timeout: defaultHTTPTimeout}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting secret creation request...")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("secret creation request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		serverMessage := internal.ExtractServerMessage(respBody)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if serverMessage != "" {
				return fmt.Errorf("%s", serverMessage)
			}
			return fmt.Errorf("secret creation failed with status %s", resp.Status)
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
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
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

		orgID, projectID, err := internal.GetProjectId(
			cmd.Context(),
			hostname,
			cfgDirName,
			sessionFileName,
			secretsOrganization,
			secretsProject,
			defaultHTTPTimeout,
		)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", secretsProject, err)
		}

		data, err := buildSecretData(secretDataKVs)
		if err != nil {
			return err
		}

		reqBody := CreateSecretRequest{
			Data: data,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/secrets/%s", orgID, projectID, secretName)

		req, err := internal.NewRequestWCookies(cmd.Context(), http.MethodPut, u.String(), bodyBytes, cookies)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		client := &http.Client{Timeout: defaultHTTPTimeout}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting secret update request...")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("secret update request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		serverMessage := internal.ExtractServerMessage(respBody)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if serverMessage != "" {
				return fmt.Errorf("%s", serverMessage)
			}
			return fmt.Errorf("secret update failed with status %s", resp.Status)
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
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
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

		orgID, projectID, err := internal.GetProjectId(
			cmd.Context(),
			hostname,
			cfgDirName,
			sessionFileName,
			secretsOrganization,
			secretsProject,
			defaultHTTPTimeout,
		)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", secretsProject, err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/secrets/%s", orgID, projectID, secretToDelete)

		req, err := internal.NewRequestWCookies(cmd.Context(), http.MethodDelete, u.String(), nil, cookies)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		client := &http.Client{Timeout: defaultHTTPTimeout}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting secret delete request...")

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("secret delete request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		serverMessage := internal.ExtractServerMessage(respBody)

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if serverMessage != "" {
				return fmt.Errorf("%s", serverMessage)
			}
			return fmt.Errorf("secret delete failed with status %s", resp.Status)
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
