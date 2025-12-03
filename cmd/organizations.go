package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

type Organization struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	ProjectCount int    `json:"project_count"`
	Role         string `json:"role"`
}

type OrganizationsResponse struct {
	Organizations []Organization `json:"organizations"`
}

var organizationsCmd = &cobra.Command{
	Use:     "organizations",
	Aliases: []string{"organization"},
	Short:   "Manage organizations",
	Long:    `Manage organizations associated with your account.`,
}

var organizationsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List organizations",
	Long:    `List all organizations you are a member of.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		url := fmt.Sprintf("%s/api/v1/session/organizations", hostname)
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
			return fmt.Errorf("failed to list organizations: server returned %s", resp.Status)
		}

		var result OrganizationsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		headers := []string{"NAME", "PROJECTS", "ROLE"}
		rows := make([][]string, len(result.Organizations))
		for i, org := range result.Organizations {
			displayName := org.Name
			if selectedOrg != "" && strings.EqualFold(org.Name, selectedOrg) {
				displayName = displayName + " *"
			}
			rows[i] = []string{displayName, fmt.Sprintf("%d", org.ProjectCount), org.Role}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var organizationsSelectCmd = &cobra.Command{
	Use:     "select [organization_name]",
	Aliases: []string{"set"},
	Short:   "Select an organization for subsequent commands",
	Long:    `Select an organization by name and store it in the local CLI configuration so other commands can use it without specifying the organization each time.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		orgName := args[0]

		if _, err := internal.GetOrgId(cmd.Context(), hostname, cfgDirName, sessionFileName, orgName, defaultHTTPTimeout); err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		if err := internal.SelectOrg(cfgDirName, orgName); err != nil {
			return fmt.Errorf("failed to store selected organization: %w", err)
		}

		fmt.Fprintf(out, "Selected organization %s\n", orgName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(organizationsCmd)
	organizationsCmd.AddCommand(organizationsListCmd)
	organizationsCmd.AddCommand(organizationsSelectCmd)
}
