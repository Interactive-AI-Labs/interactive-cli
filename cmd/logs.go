package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

var (
	logsProject      string
	logsOrganization string
	logsFollow       bool
)

var logsCmd = &cobra.Command{
	Use:   "logs [replica_name]",
	Short: "Show logs for a specific replica",
	Long: `Show logs for a specific replica (pod) in a project.

The project is selected with --project. If --organization is not provided,
the currently selected organization (via 'interactiveai organizations select')
is used.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		if logsProject == "" {
			return fmt.Errorf("project is required; please provide --project")
		}

		replicaName := args[0]

		// Load session cookies.
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
		if logsOrganization == "" {
			if selectedOrg == "" {
				return fmt.Errorf("organization is required; please provide --organization or run '%s organizations select <name>'", rootCmd.Use)
			}
			logsOrganization = selectedOrg
		}

		orgId, projectId, err := internal.GetProjectId(
			cmd.Context(),
			hostname,
			cfgDirName,
			sessionFileName,
			logsOrganization,
			logsProject,
			defaultHTTPTimeout,
		)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", logsProject, err)
		}

		// Build request URL to deployment-operator.
		u, err := url.Parse(deploymentHostname)
		if err != nil {
			return fmt.Errorf("failed to parse deployment service URL: %w", err)
		}
		u.Path = fmt.Sprintf("/v1/organizations/%s/projects/%s/services/replicas/%s/logs", orgId, projectId, replicaName)

		q := u.Query()
		if logsFollow {
			q.Set("follow", "true")
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, u.String(), nil)
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
			return fmt.Errorf("logs request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			msg := internal.ExtractServerMessage(respBody)
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return fmt.Errorf("logs request failed with status %s", resp.Status)
		}

		_, err = io.Copy(out, resp.Body)
		return err
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsProject, "project", "p", "", "Project name that owns the service")
	logsCmd.Flags().StringVarP(&logsOrganization, "organization", "o", "", "Organization name that owns the project")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")

	rootCmd.AddCommand(logsCmd)
}
