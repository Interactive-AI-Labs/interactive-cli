package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	jobRunJSON, jobRunYAML                                      bool
	jobRunsListJSON, jobRunsListYAML, jobRunsListWatch          bool
	jobRunDescribeJSON, jobRunDescribeYAML, jobRunDescribeWatch bool
	jobRunStopJSON, jobRunStopYAML                              bool
)

var jobRunCmd = &cobra.Command{
	Use:   "run <job_name>",
	Short: "Run a job in a project",
	Long: `Start a run using the job's saved configuration.

The job must be active and have no unfinished run. The command returns the run
without waiting for completion; use 'iai jobs runs describe --watch' to follow its status.`,
	Example: `  iai jobs run my-job
  iai jobs run my-job --project my-project
  iai jobs run my-job --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		if !jobRunJSON && !jobRunYAML {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Submitting job run request...")
		}
		run, err := client.StartJobRun(cmd.Context(), pCtx.orgId, pCtx.projectId, name)
		if err != nil {
			return err
		}

		if jobRunJSON {
			return output.PrintStructuredJSON(out, run)
		}
		if jobRunYAML {
			return output.PrintStructuredYAML(out, run)
		}
		return output.PrintJobRunDescribe(out, run, time.Now())
	},
}

var jobRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Inspect and manage job runs",
	Long:  `Manage retained runs of jobs in a project.`,
}

var jobRunsListCmd = &cobra.Command{
	Use:     "list <job_name>",
	Aliases: []string{"ls"},
	Short:   "List runs for a job",
	Long: `List retained runs of a job in a project, sorted newest-first.

Logs have a separate retention period.`,
	Example: `  iai jobs runs list my-job
  iai jobs runs list my-job --watch
  iai jobs runs list my-job --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		render := func(ctx context.Context, w io.Writer) error {
			runs, err := client.ListJobRuns(ctx, pCtx.orgId, pCtx.projectId, name)
			if err != nil {
				return err
			}

			if jobRunsListJSON {
				return output.PrintStructuredJSON(w, runs)
			}
			if jobRunsListYAML {
				return output.PrintStructuredYAML(w, runs)
			}
			return output.PrintJobRunList(w, runs, time.Now())
		}

		if jobRunsListWatch {
			return runWatch(cmd, render)
		}
		return render(cmd.Context(), cmd.OutOrStdout())
	},
}

var jobRunDescribeCmd = &cobra.Command{
	Use:     "describe <run_id>",
	Aliases: []string{"desc", "get"},
	Short:   "Describe a job run in detail",
	Long:    `Show the latest execution state and explanation for one retained run.`,
	Example: `  iai jobs runs describe <run_id>
  iai jobs runs describe <run_id> --watch
  iai jobs runs describe <run_id> --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runId := strings.TrimSpace(args[0])

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		render := func(ctx context.Context, w io.Writer) error {
			run, err := client.DescribeJobRun(ctx, pCtx.orgId, pCtx.projectId, runId)
			if err != nil {
				return err
			}

			if jobRunDescribeJSON {
				return output.PrintStructuredJSON(w, run)
			}
			if jobRunDescribeYAML {
				return output.PrintStructuredYAML(w, run)
			}
			return output.PrintJobRunDescribe(w, run, time.Now())
		}

		if jobRunDescribeWatch {
			return runWatch(cmd, render)
		}
		return render(cmd.Context(), cmd.OutOrStdout())
	},
}

var jobRunStopCmd = &cobra.Command{
	Use:   "stop <run_id>",
	Short: "Stop an unfinished job run",
	Long: `Request a permanent stop of one unfinished run without disabling future runs.
Returns without waiting for termination; an already-finished run is unchanged.`,
	Example: `  iai jobs runs stop <run_id>
  iai jobs runs stop <run_id> --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		runId := strings.TrimSpace(args[0])

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		if !jobRunStopJSON && !jobRunStopYAML {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Submitting job run stop request...")
		}
		run, err := client.StopJobRun(cmd.Context(), pCtx.orgId, pCtx.projectId, runId)
		if err != nil {
			return err
		}

		if jobRunStopJSON {
			return output.PrintStructuredJSON(out, run)
		}
		if jobRunStopYAML {
			return output.PrintStructuredYAML(out, run)
		}
		return output.PrintJobRunDescribe(out, run, time.Now())
	},
}

var jobRunDeleteCmd = &cobra.Command{
	Use:     "delete <run_id>",
	Aliases: []string{"rm"},
	Short:   "Delete a finished job run",
	Long:    `Delete a finished run from history. Unfinished runs cannot be deleted.`,
	Example: `  iai jobs runs delete <run_id>
  iai jobs runs delete <run_id> --project my-project`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		runId := strings.TrimSpace(args[0])

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting job run deletion request...")
		message, err := client.DeleteJobRun(cmd.Context(), pCtx.orgId, pCtx.projectId, runId)
		if err != nil {
			return err
		}

		if message != "" {
			fmt.Fprintln(out, message)
		}
		return nil
	},
}

func init() {
	jobRunCmd.Flags().BoolVar(&jobRunJSON, "json", false, "Output raw API response as JSON")
	jobRunCmd.Flags().BoolVar(&jobRunYAML, "yaml", false, "Output raw API response as YAML")
	jobRunCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	jobRunsListCmd.Flags().
		BoolVar(&jobRunsListJSON, "json", false, "Output raw API response as JSON")
	jobRunsListCmd.Flags().
		BoolVar(&jobRunsListYAML, "yaml", false, "Output raw API response as YAML")
	jobRunsListCmd.Flags().
		BoolVarP(&jobRunsListWatch, "watch", "w", false, "Poll and refresh the list every 2s until interrupted")
	jobRunsListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	jobRunsListCmd.MarkFlagsMutuallyExclusive("watch", "json")
	jobRunsListCmd.MarkFlagsMutuallyExclusive("watch", "yaml")
	jobRunDescribeCmd.Flags().
		BoolVar(&jobRunDescribeJSON, "json", false, "Output raw API response as JSON")
	jobRunDescribeCmd.Flags().
		BoolVar(&jobRunDescribeYAML, "yaml", false, "Output raw API response as YAML")
	jobRunDescribeCmd.Flags().
		BoolVarP(&jobRunDescribeWatch, "watch", "w", false, "Poll and refresh every 2s until interrupted")
	jobRunDescribeCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	jobRunDescribeCmd.MarkFlagsMutuallyExclusive("watch", "json")
	jobRunDescribeCmd.MarkFlagsMutuallyExclusive("watch", "yaml")

	jobRunStopCmd.Flags().BoolVar(&jobRunStopJSON, "json", false, "Output raw API response as JSON")
	jobRunStopCmd.Flags().BoolVar(&jobRunStopYAML, "yaml", false, "Output raw API response as YAML")
	jobRunStopCmd.MarkFlagsMutuallyExclusive("json", "yaml")

	jobsCmd.AddCommand(jobRunCmd, jobRunsCmd)
	jobRunsCmd.AddCommand(jobRunsListCmd, jobRunDescribeCmd, jobRunStopCmd, jobRunDeleteCmd)
}
