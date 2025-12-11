package cmd

import (
	"fmt"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	output "github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

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

		if apiKey != "" {
			return fmt.Errorf("organizations list is not available when using API key authentication")
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		orgs, err := apiClient.ListOrganizations(cmd.Context())
		if err != nil {
			return err
		}

		selectedOrg, err := files.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		headers := []string{"NAME", "PROJECTS", "ROLE"}
		rows := make([][]string, len(orgs))
		for i, org := range orgs {
			displayName := org.Name
			if selectedOrg != "" && strings.EqualFold(org.Name, selectedOrg) {
				displayName = displayName + " *"
			}
			rows[i] = []string{displayName, fmt.Sprintf("%d", org.ProjectCount), org.Role}
		}

		if err := output.PrintTable(out, headers, rows); err != nil {
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

		if apiKey != "" {
			return fmt.Errorf("organizations select is not available when using API key authentication")
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		if _, err := apiClient.GetOrgIdByName(cmd.Context(), orgName); err != nil {
			return fmt.Errorf("failed to resolve organization %q: %w", orgName, err)
		}

		if err := files.SelectOrg(cfgDirName, orgName); err != nil {
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
