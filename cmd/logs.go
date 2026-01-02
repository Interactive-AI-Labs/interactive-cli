package cmd

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
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

The project is selected with --project or via 'iai projects select'. If --organization is not provided,
the currently selected organization (via 'iai organizations select')
is used.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		replicaName := args[0]
		ctx := cmd.Context()
		if logsFollow {
			var stop func()
			ctx, stop = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer stop()
		}

		var cfg *files.StackConfig
		if cfgFilePath != "" {
			loadedCfg, err := files.LoadStackConfig(cfgFilePath)
			if err != nil {
				return fmt.Errorf("failed to load config file: %w", err)
			}
			cfg = loadedCfg
		} else {
			cfg = &files.StackConfig{}
		}

		cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}

		apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
		if err != nil {
			return err
		}

		timeout := defaultHTTPTimeout
		if logsFollow {
			timeout = 0
		}

		deployClient, err := clients.NewDeploymentClient(deploymentHostname, timeout, apiKey, cookies)
		if err != nil {
			return err
		}

		sess := session.NewSession(cfgDirName)

		orgName, err := sess.ResolveOrganization(cfg.Organization, logsOrganization)
		if err != nil {
			return err
		}

		projectName, err := sess.ResolveProject(cfg.Project, logsProject)
		if err != nil {
			return err
		}

		orgId, projectId, err := apiClient.GetProjectId(cmd.Context(), orgName, projectName)
		if err != nil {
			return fmt.Errorf("failed to resolve project %q: %w", projectName, err)
		}

		logReader, err := deployClient.GetLogs(ctx, orgId, projectId, replicaName, logsFollow)
		if err != nil {
			return err
		}
		defer logReader.Close()

		_, err = io.Copy(out, logReader)
		if logsFollow && ctx.Err() != nil {
			return nil
		}
		return err
	},
}

func init() {
	logsCmd.Flags().StringVarP(&logsProject, "project", "p", "", "Project name that owns the service")
	logsCmd.Flags().StringVarP(&logsOrganization, "organization", "o", "", "Organization name that owns the project")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")

	rootCmd.AddCommand(logsCmd)
}
