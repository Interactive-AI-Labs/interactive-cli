package cmd

import (
	"fmt"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
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

		sess := session.NewSession(cfgDirName)
		orgName, err := sess.ResolveOrganization("", projectsOrganization)
		if err != nil {
			return err
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		orgId, err := apiClient.GetOrgIdByName(cmd.Context(), orgName)
		if err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		projects, err := apiClient.ListProjects(cmd.Context(), orgId)
		if err != nil {
			return err
		}

		selectedProject, err := files.GetSelectedProject(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Fprintln(out)

		headers := []string{"NAME", "ROLE"}
		rows := make([][]string, len(projects))
		for i, proj := range projects {
			displayName := proj.Name
			if selectedProject != "" && strings.EqualFold(proj.Name, selectedProject) {
				displayName = displayName + " *"
			}
			rows[i] = []string{displayName, proj.Role}
		}

		if err := output.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

var projectsSelectCmd = &cobra.Command{
	Use:     "select [project_name]",
	Aliases: []string{"set"},
	Short:   "Select a project for subsequent commands",
	Long:    `Select a project by name and store it in the local CLI configuration so other commands can use it without specifying the project each time.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		projectName := args[0]

		if apiKey != "" {
			return fmt.Errorf("projects select is not available when using API key authentication")
		}

		sess := session.NewSession(cfgDirName)
		orgName, err := sess.ResolveOrganization("", projectsOrganization)
		if err != nil {
			return err
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		orgId, err := apiClient.GetOrgIdByName(cmd.Context(), orgName)
		if err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		if _, err := apiClient.GetProjectByName(cmd.Context(), orgId, projectName); err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		if err := files.SelectProject(cfgDirName, projectName); err != nil {
			return fmt.Errorf("failed to store selected project: %w", err)
		}

		fmt.Fprintf(out, "Selected project %s\n", projectName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	projectsListCmd.Flags().StringVarP(&projectsOrganization, "organization", "o", "", "Organization name that owns the projects")
	projectsSelectCmd.Flags().StringVarP(&projectsOrganization, "organization", "o", "", "Organization name that owns the project")

	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsSelectCmd)
}
