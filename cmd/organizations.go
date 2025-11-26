package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

const organizationsURL = "https://dev.interactive.ai/api/v1/session/organizations"

type Organization struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProjectCount int    `json:"project_count"`
	Role         string `json:"role"`
}

type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}

var organizationsCmd = &cobra.Command{
	Use:   "organizations",
	Short: "Manage organizations",
	Long:  `Manage organizations associated with your account.`,
}

var organizationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List organizations",
	Long:  `List all organizations you are a member of.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, organizationsURL, nil)
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
			return fmt.Errorf("failed to list organizations: server returned %s", resp.Status)
		}

		var result OrganizationsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		headers := []string{"ID", "NAME", "PROJECTS", "ROLE"}
		rows := make([][]string, len(result.Organizations))
		for i, org := range result.Organizations {
			rows[i] = []string{org.ID, org.Name, fmt.Sprintf("%d", org.ProjectCount), org.Role}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(organizationsCmd)
	organizationsCmd.AddCommand(organizationsListCmd)
}
