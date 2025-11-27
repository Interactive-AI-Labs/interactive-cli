package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type ProjectsResponse struct {
	OrganizationID   string    `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	UserRole         string    `json:"user_role"`
	Projects         []Project `json:"projects"`
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Long:  `Manage projects associated with an organization.`,
}

var projectsListCmd = &cobra.Command{
	Use:   "list [organization_name]",
	Short: "List projects in an organization",
	Long:  `List all projects within a specific organization. The organization name will be resolved to its ID before making API calls.`,
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		var orgName string
		if len(args) > 0 {
			orgName = args[0]
		} else {
			selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide a name or run '%s organizations select <name>'", rootCmd.Use)
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

		orgID, err := internal.GetOrgId(cmd.Context(), hostname, cfgDirName, sessionFileName, orgName, defaultHTTPTimeout)
		if err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		url := fmt.Sprintf("%s/api/v1/session/organizations/%s/projects", hostname, orgID)
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
	projectsCmd.AddCommand(projectsListCmd)
}
