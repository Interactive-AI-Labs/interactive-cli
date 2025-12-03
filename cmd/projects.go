package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

type Project struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type ProjectsResponse struct {
	OrganizationId   string    `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	UserRole         string    `json:"user_role"`
	Projects         []Project `json:"projects"`
}

var projectsOrganization string

var projectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project"},
	Short:   "Manage projects",
	Long:    `Manage projects associated with an organization.`,
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects in an organization",
	Long:  `List all projects within a specific organization. The organization name will be resolved to its Id before making API calls.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var orgName string
		if projectsOrganization != "" {
			orgName = projectsOrganization
		} else {
			selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			orgName = selectedOrg
		}

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		orgId, err := internal.GetOrgId(cmd.Context(), hostname, cfgDirName, sessionFileName, orgName, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		url := fmt.Sprintf("%s/api/v1/session/organizations/%s/projects", hostname, orgId)
		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}

		client := &http.Client{
			Timeout: defaultHTTPTimeout,
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("failed to list projects: server returned %s", resp.Status)
		}

		var result ProjectsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		fmt.Fprintln(out)

		headers := []string{"NAME", "ROLE"}
		rows := make([][]string, len(result.Projects))
		for i, proj := range result.Projects {
			rows[i] = []string{proj.Name, proj.Role}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	projectsListCmd.Flags().StringVar(&projectsOrganization, "organization", "", "Organization name that owns the projects")

	projectsCmd.AddCommand(projectsListCmd)
}
