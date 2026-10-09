package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	jobProject, jobOrganization                                   string
	jobCreateInput                                                inputs.JobInput
	jobScriptPath, jobPyprojectPath                               string
	jobTimeout                                                    int64
	jobRetries, jobRetentionTTL, jobSuccessfulRuns, jobFailedRuns int32
	jobListJSON, jobListYAML, jobListWatch                        bool
	jobDescribeJSON, jobDescribeYAML, jobDescribeWatch            bool
	jobDescribeScript, jobDescribePyproject                       bool
)

var jobsCmd = &cobra.Command{
	Use:     "jobs",
	Aliases: []string{"job"},
	Short:   "Manage saved jobs and executions",
	GroupID: groupInfra,
	Long:    `Manage saved jobs and their executions in InteractiveAI projects.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		chainRootPersistentPreRun(cmd, args)
		return inputs.ValidateJobArgs(args)
	},
}

var jobCreateCmd = &cobra.Command{
	Use:   "create <job_name>",
	Short: "Create a job in a project",
	Long: `Save an image or Python-script job without starting a run.
Use 'iai jobs run' for manual execution.`,
	Example: `  iai jobs create my-job --image-type external --image-repository docker.io --image-name python --image-tag 3.12-slim --command python --args=-c --args="print('hello')" --memory 512M --cpu 0.5
  iai jobs create my-job --image-type internal --image-name my-app --image-tag v1 --memory 1G --cpu 1 --env LOG_LEVEL=debug --secret DB_PASSWORD
  iai jobs create report --type script --script main.py --pyproject pyproject.toml --memory 512M --cpu 0.5
  iai jobs create daily-report --type script --script main.py --pyproject pyproject.toml --memory 512M --cpu 0.5 --schedule "0 2 * * *" --schedule-timezone Europe/Berlin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])
		in := jobCreateInput

		var err error
		in.Script, err = inputs.ReadJobFile(jobScriptPath)
		if err != nil {
			return err
		}
		in.Pyproject, err = inputs.ReadJobFile(jobPyprojectPath)
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("timeout") {
			in.Timeout = utils.ToPtr(jobTimeout)
		}
		if cmd.Flags().Changed("retries") {
			in.Retries = utils.ToPtr(jobRetries)
		}
		if cmd.Flags().Changed("retention-ttl") {
			in.Retention.TTL = utils.ToPtr(jobRetentionTTL)
		}
		if cmd.Flags().Changed("retention-successful-runs") {
			in.Retention.SuccessfulRuns = utils.ToPtr(jobSuccessfulRuns)
		}
		if cmd.Flags().Changed("retention-failed-runs") {
			in.Retention.FailedRuns = utils.ToPtr(jobFailedRuns)
		}

		body, err := inputs.BuildJobRequestBody(in)
		if err != nil {
			return err
		}

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting job creation request...")
		message, err := client.CreateJob(cmd.Context(), pCtx.orgId, pCtx.projectId, name, body)
		if err != nil {
			return err
		}

		if message != "" {
			fmt.Fprintln(out, message)
		}
		return nil
	},
}

var jobListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List jobs in a project",
	Long:    `List jobs in a specific project using the deployment service.`,
	Example: `  iai jobs list
  iai jobs list --project my-project
  iai jobs list --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		render := func(ctx context.Context, w io.Writer) error {
			jobs, err := client.ListJobs(ctx, pCtx.orgId, pCtx.projectId, "")
			if err != nil {
				return err
			}

			if jobListJSON {
				return output.PrintStructuredJSON(w, jobs)
			}
			if jobListYAML {
				return output.PrintStructuredYAML(w, jobs)
			}
			return output.PrintJobList(w, jobs)
		}

		if jobListWatch {
			return runWatch(cmd, render)
		}
		return render(cmd.Context(), cmd.OutOrStdout())
	},
}

var jobDescribeCmd = &cobra.Command{
	Use:     "describe <job_name>",
	Aliases: []string{"desc", "get"},
	Short:   "Describe a job in detail",
	Long: `Show detailed information about a specific job including its configuration.

For script jobs, the script and project file are summarized; use --script or
--pyproject to print that file exactly as stored, e.g. to save it locally.`,
	Example: `  iai jobs describe my-job
  iai jobs describe my-job --json
  iai jobs describe report --script > main.py
  iai jobs describe report --pyproject > pyproject.toml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(args[0])

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		render := func(ctx context.Context, w io.Writer) error {
			job, err := client.DescribeJob(ctx, pCtx.orgId, pCtx.projectId, name)
			if err != nil {
				return err
			}

			if jobDescribeScript || jobDescribePyproject {
				return output.PrintJobFile(w, job, jobDescribePyproject)
			}
			if jobDescribeJSON {
				return output.PrintStructuredJSON(w, job)
			}
			if jobDescribeYAML {
				return output.PrintStructuredYAML(w, job)
			}
			return output.PrintJobDescribe(w, job)
		}

		if jobDescribeWatch {
			return runWatch(cmd, render)
		}
		return render(cmd.Context(), cmd.OutOrStdout())
	},
}

var jobDeleteCmd = &cobra.Command{
	Use:     "delete <job_name>",
	Aliases: []string{"rm"},
	Short:   "Delete a job from a project",
	Long: `Delete a job from a specific project using the deployment service.

A job can only be deleted when all its runs have finished.`,
	Example: `  iai jobs delete my-job
  iai jobs delete my-job --project my-project`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}

		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting job deletion request...")
		message, err := client.DeleteJob(cmd.Context(), pCtx.orgId, pCtx.projectId, name)
		if err != nil {
			return err
		}

		if message != "" {
			fmt.Fprintln(out, message)
		}
		return nil
	},
}

var jobActivateCmd = &cobra.Command{
	Use:   "activate <job_name>",
	Short: "Activate a deactivated job in a project",
	Long:  `Activate a deactivated job, allowing new runs and restoring its schedule, if configured.`,
	Example: `  iai jobs activate my-job
  iai jobs activate my-job --project my-project`,
	Args: cobra.ExactArgs(1),
	RunE: runJobAction,
}

var jobDeactivateCmd = &cobra.Command{
	Use:   "deactivate <job_name>",
	Short: "Deactivate a job in a project",
	Long: `Deactivate a job, preventing new runs. Existing runs continue.
The current configuration is preserved and will be restored when the job is activated again.`,
	Example: `  iai jobs deactivate my-job
  iai jobs deactivate my-job --project my-project`,
	Args: cobra.ExactArgs(1),
	RunE: runJobAction,
}

func runJobAction(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	name := strings.TrimSpace(args[0])

	pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
	if err != nil {
		return err
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "Submitting job %s request...\n", cmd.Name())
	message, err := client.JobAction(cmd.Context(), pCtx.orgId, pCtx.projectId, name, cmd.Name())
	if err != nil {
		return err
	}

	if message != "" {
		fmt.Fprintln(out, message)
	}
	return nil
}

func init() {
	jobsCmd.PersistentFlags().StringVarP(&jobProject, "project", "p", "", "Project name")
	jobsCmd.PersistentFlags().
		StringVarP(&jobOrganization, "organization", "o", "", "Organization name that owns the project")

	flags := jobCreateCmd.Flags()
	flags.StringVar(&jobCreateInput.Type, "type", "image", "Job type: 'image' or 'script'")
	flags.StringVar(
		&jobCreateInput.ImageType,
		"image-type",
		"",
		"Image type: 'internal' (project's private registry), 'external' (any public registry), or 'platform' (Interactive AI registries)",
	)
	flags.StringVar(
		&jobCreateInput.ImageRepository,
		"image-repository",
		"",
		"Image repository; required for external and platform images",
	)
	flags.StringVar(&jobCreateInput.ImageName, "image-name", "", "Container image name")
	flags.StringVar(&jobCreateInput.ImageTag, "image-tag", "", "Container image tag")
	flags.StringArrayVar(
		&jobCreateInput.Command,
		"command",
		nil,
		"Container entrypoint override; can be repeated; image jobs only",
	)
	flags.StringArrayVar(
		&jobCreateInput.Args,
		"args",
		nil,
		"Container arguments; can be repeated; image jobs only",
	)
	flags.StringVar(
		&jobScriptPath,
		"script",
		"",
		"Python script path; required for script jobs; combined with --pyproject, maximum 350,000 bytes",
	)
	flags.StringVar(
		&jobPyprojectPath,
		"pyproject",
		"",
		"pyproject.toml path; required for script jobs; combined with --script, maximum 350,000 bytes",
	)
	flags.StringVar(
		&jobCreateInput.Memory,
		"memory",
		"",
		"Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G)",
	)
	flags.StringVar(
		&jobCreateInput.CPU,
		"cpu",
		"",
		"CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m)",
	)
	_ = jobCreateCmd.MarkFlagRequired("memory")
	_ = jobCreateCmd.MarkFlagRequired("cpu")
	flags.StringArrayVar(
		&jobCreateInput.EnvVars,
		"env",
		nil,
		"Environment variable (NAME=VALUE); can be repeated",
	)
	flags.StringArrayVar(
		&jobCreateInput.SecretRefs,
		"secret",
		nil,
		"Secrets to be loaded as env vars; can be repeated",
	)
	flags.StringVar(&jobCreateInput.StackId, "stack-id", "", "Stack ID to assign the job to")
	flags.StringVar(
		&jobCreateInput.Schedule,
		"schedule",
		"",
		"Five-field cron schedule or calendar shortcut (e.g. '0 2 * * *', '@daily'); omit for manual execution only",
	)
	flags.StringVar(
		&jobCreateInput.Timezone,
		"schedule-timezone",
		"",
		"IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC); defaults to UTC; requires --schedule",
	)
	flags.Int64Var(
		&jobTimeout,
		"timeout",
		0,
		"Run time budget in seconds (1-21600, max 6h), including startup, dependency installation, and all retries; server default: 3600 seconds",
	)
	indexFlag(jobCreateCmd, "timeout", "1-21600, max 6h")
	flags.Int32Var(
		&jobRetries,
		"retries",
		0,
		"Maximum retries after failure; share the run timeout and may repeat side effects; server default: 0",
	)
	flags.Int32Var(
		&jobRetentionTTL,
		"retention-ttl",
		0,
		"Seconds to keep finished runs (0-2592000); zero requests immediate cleanup; server default: 604800 seconds",
	)
	flags.Int32Var(
		&jobSuccessfulRuns,
		"retention-successful-runs",
		0,
		"Successful runs to retain (0-20); server default: 5",
	)
	flags.Int32Var(
		&jobFailedRuns,
		"retention-failed-runs",
		0,
		"Failed, timed-out, or stopped runs to retain (0-20); server default: 10",
	)
	jobCreateCmd.MarkFlagsRequiredTogether("script", "pyproject")

	jobListCmd.Flags().BoolVar(&jobListJSON, "json", false, "Output raw API response as JSON")
	jobListCmd.Flags().BoolVar(&jobListYAML, "yaml", false, "Output raw API response as YAML")
	jobListCmd.Flags().
		BoolVarP(&jobListWatch, "watch", "w", false, "Poll and refresh the list every 2s until interrupted")
	jobListCmd.MarkFlagsMutuallyExclusive("json", "yaml")
	jobListCmd.MarkFlagsMutuallyExclusive("watch", "json")
	jobListCmd.MarkFlagsMutuallyExclusive("watch", "yaml")
	jobDescribeCmd.Flags().
		BoolVar(&jobDescribeJSON, "json", false, "Output raw API response as JSON")
	jobDescribeCmd.Flags().
		BoolVar(&jobDescribeYAML, "yaml", false, "Output raw API response as YAML")
	jobDescribeCmd.Flags().
		BoolVarP(&jobDescribeWatch, "watch", "w", false, "Poll and refresh every 2s until interrupted")
	jobDescribeCmd.Flags().
		BoolVar(&jobDescribeScript, "script", false, "Print only the script of a script job")
	jobDescribeCmd.Flags().
		BoolVar(&jobDescribePyproject, "pyproject", false, "Print only the pyproject.toml of a script job")
	jobDescribeCmd.MarkFlagsMutuallyExclusive("json", "yaml", "watch", "script", "pyproject")

	rootCmd.AddCommand(jobsCmd)
	jobsCmd.AddCommand(
		jobCreateCmd,
		jobListCmd,
		jobDescribeCmd,
		jobDeleteCmd,
		jobActivateCmd,
		jobDeactivateCmd,
	)
}
