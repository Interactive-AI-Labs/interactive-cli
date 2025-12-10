package cmd

import (
	"fmt"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

var projectsOrganization string

var projectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project"},
	Short:   "Manage projects",
	Long:    `Manage projects associated with an organization.`,
}

var projectsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List projects in an organization",
	Long:    `List all projects within a specific organization. The organization name will be resolved to its Id before making API calls.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if apiKey != "" {
			return fmt.Errorf("projects list is not available when using API key authentication")
		}

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

		apiClient, err := internal.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		orgId, err := apiClient.GetOrganizationByName(cmd.Context(), orgName)
		if err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		projects, err := apiClient.ListProjects(cmd.Context(), orgId)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)

		headers := []string{"NAME", "ROLE"}
		rows := make([][]string, len(projects))
		for i, proj := range projects {
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

	projectsListCmd.Flags().StringVarP(&projectsOrganization, "organization", "o", "", "Organization name that owns the projects")

	projectsCmd.AddCommand(projectsListCmd)
}
