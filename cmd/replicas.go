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

var (
	replicasProject      string
	replicasOrganization string
)

var replicasCmd = &cobra.Command{
	Use:   "replicas",
	Short: "Manage service replicas",
	Long:  `Manage pods backing services in a specific project.`,
}

var replicasListCmd = &cobra.Command{
	Use:     "list [service_name]",
	Aliases: []string{"ls"},
	Short:   "List replicas for a service",
	Long: `List pods backing a service in a specific project.

The project is selected with --project.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		serviceName := strings.TrimSpace(args[0])

		if strings.TrimSpace(replicasProject) == "" {
			return fmt.Errorf("project is required; please provide --project")
		}
		if serviceName == "" {
			return fmt.Errorf("service name is required")
		}

		// Ensure the user is logged in and load session cookies.
		cookies, err := internal.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if len(cookies) == 0 {
			return fmt.Errorf("not logged in. Please run '%s login' first", rootCmd.Use)
		}

		// Resolve organization name.
		selectedOrg, err := internal.GetSelectedOrg(cfgDirName)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if strings.TrimSpace(replicasOrganization) == "" {
			if strings.TrimSpace(selectedOrg) == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			replicasOrganization = selectedOrg
		}

		orgID, projectID, err := internal.GetProjectId(
			cmd.Context(),
			hostname,
			cfgDirName,
			sessionFileName,
			replicasOrganization,
			replicasProject,
			defaultHTTPTimeout,
		)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", replicasProject, err)
		}

		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/organizations/%s/projects/%s/services/%s/replicas", orgID, projectID, serviceName)

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, u.String(), nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}

		client := &http.Client{
			Timeout: defaultHTTPTimeout,
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("replicas request failed: %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			msg := internal.ExtractServerMessage(respBody)
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("replicas request failed with status %s", resp.Status)
		}

		var result ListServiceReplicasResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("failed to decode replicas response: %w", err)
		}

		headers := []string{"NAME", "PHASE", "STATUS", "READY", "STARTED"}
		rows := make([][]string, len(result.Replicas))
		for i, r := range result.Replicas {
			rows[i] = []string{
				r.Name,
				r.Phase,
				r.Status,
				fmt.Sprintf("%t", r.Ready),
				r.StartTime,
			}
		}

		if err := internal.PrintTable(out, headers, rows); err != nil {
			return fmt.Errorf("failed to print table: %w", err)
		}

		return nil
	},
}

func init() {
	// Flags for "replicas list"
	replicasListCmd.Flags().StringVarP(&replicasProject, "project", "p", "", "Project name that owns the service")
	replicasListCmd.Flags().StringVarP(&replicasOrganization, "organization", "o", "", "Organization name that owns the project")

	replicasCmd.AddCommand(replicasListCmd)
	rootCmd.AddCommand(replicasCmd)
}
