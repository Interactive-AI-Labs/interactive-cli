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
	Use:   "list [organization_id]",
	Short: "List projects in an organization",
	Long:  `List all projects within a specific organization.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		orgID := args[0]
		out := cmd.OutOrStdout()

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
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

		fmt.Fprintf(out, "Organization: %s (%s)\n", result.OrganizationName, result.OrganizationID)
		fmt.Fprintln(out)

		headers := []string{"ID", "NAME", "ROLE"}
		rows := make([][]string, len(result.Projects))
		for i, proj := range result.Projects {
			rows[i] = []string{proj.ID, proj.Name, proj.Role}
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
