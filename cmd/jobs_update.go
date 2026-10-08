package cmd

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/spf13/cobra"
)

var (
	jobUpdateInput                                                                        inputs.JobUpdateInput
	jobUpdateScriptPath, jobUpdatePyprojectPath                                           string
	jobUpdateTimeout                                                                      int64
	jobUpdateRetries, jobUpdateRetentionTTL, jobUpdateSuccessfulRuns, jobUpdateFailedRuns int32
	jobExpectRevision                                                                     int
	jobUpdateForce                                                                        bool
)

var jobUpdateCmd = &cobra.Command{
	Use:   "update <job_name>",
	Short: "Update a job in a project",
	Long: `Update supplied fields of a saved job; omitted fields keep their values.
Preserves activation state and run history. All runs must be finished.`,
	Example: `  iai jobs update my-job --image-tag v2
  iai jobs update my-job --image-tag v2 --expect-revision 3
  iai jobs update my-job --memory 1G --cpu 0.5
  iai jobs update report --script main.py
  iai jobs update report --schedule "0 2 * * *" --schedule-timezone Europe/Berlin
  iai jobs update report --clear-schedule
  iai jobs update report --retention-successful-runs 0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		name := strings.TrimSpace(args[0])
		in := jobUpdateInput
		var err error
		if cmd.Flags().Changed("script") {
			in.Script, err = inputs.ReadJobFile(jobUpdateScriptPath)
			if err != nil {
				return err
			}
		}
		if cmd.Flags().Changed("pyproject") {
			in.Pyproject, err = inputs.ReadJobFile(jobUpdatePyprojectPath)
			if err != nil {
				return err
			}
		}
		in.Timeout, in.Retries = utils.ToPtr(jobUpdateTimeout), utils.ToPtr(jobUpdateRetries)
		in.Retention = deployment.JobRetention{
			TTL: utils.ToPtr(
				jobUpdateRetentionTTL,
			),
			SuccessfulRuns: utils.ToPtr(jobUpdateSuccessfulRuns),
			FailedRuns:     utils.ToPtr(jobUpdateFailedRuns),
		}
		patch, err := inputs.BuildJobUpdatePatch(in, cmd.Flags().Changed)
		if err != nil {
			return err
		}
		if len(patch) == 0 {
			return fmt.Errorf("no fields to update; pass at least one flag")
		}

		pCtx, _, client, err := resolveProject(cmd.Context(), jobOrganization, jobProject)
		if err != nil {
			return err
		}
		live, liveErr := client.DescribeJob(cmd.Context(), pCtx.orgId, pCtx.projectId, name)
		var revision int
		var updated string
		if liveErr == nil {
			revision, updated = live.Revision, live.Updated
		}
		if err := runUpdatePreflight(
			cmd.ErrOrStderr(),
			revision,
			updated,
			liveErr,
			cmd.Flags().Changed("expect-revision"),
			jobExpectRevision,
		); err != nil {
			return err
		}
		dropped := false
		if liveErr == nil {
			dropped = printDroppedEnvSecretWarnings(
				cmd.ErrOrStderr(),
				cmd.Flags().Changed("env"),
				cmd.Flags().Changed("secret"),
				live.Env,
				live.SecretRefs,
				in.EnvVars,
				in.SecretRefs,
			)
		}
		if err := checkUpdateGates(jobUpdateForce, false, dropped); err != nil {
			return err
		}
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Submitting job update request...")
		message, err := client.PatchJob(cmd.Context(), pCtx.orgId, pCtx.projectId, name, patch)
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
	flags := jobUpdateCmd.Flags()
	flags.StringVar(
		&jobUpdateInput.Type,
		"type",
		"",
		"Job type: 'image' or 'script'; clears the old type's fields; supply the new type's inputs",
	)
	flags.StringVar(
		&jobUpdateInput.ImageType,
		"image-type",
		"",
		"Image type: 'internal' (project's private registry), 'external' (any public registry), or 'platform' (Interactive AI registries)",
	)
	flags.StringVar(
		&jobUpdateInput.ImageRepository,
		"image-repository",
		"",
		"Image repository; required for external and platform images",
	)
	flags.StringVar(&jobUpdateInput.ImageName, "image-name", "", "Container image name")
	flags.StringVar(&jobUpdateInput.ImageTag, "image-tag", "", "Container image tag")
	flags.StringArrayVar(
		&jobUpdateInput.Command,
		"command",
		nil,
		"Replace the entrypoint; repeat for every argument to keep; image jobs only",
	)
	flags.StringArrayVar(
		&jobUpdateInput.Args,
		"args",
		nil,
		"Replace image arguments; repeat for every argument to keep; image jobs only",
	)
	flags.StringVar(
		&jobUpdateScriptPath,
		"script",
		"",
		"Replace the script from a local file; combined with the new or retained project file, maximum 350,000 bytes",
	)
	flags.StringVar(
		&jobUpdatePyprojectPath,
		"pyproject",
		"",
		"Replace pyproject.toml from a local file; combined with the new or retained script, maximum 350,000 bytes",
	)
	flags.StringVar(
		&jobUpdateInput.Memory,
		"memory",
		"",
		"Memory in megabytes (M) or gigabytes (G) (e.g. 128M, 512M, 1G, 1.5G)",
	)
	flags.StringVar(
		&jobUpdateInput.CPU,
		"cpu",
		"",
		"CPU cores or millicores (e.g. 0.5, 1, 2, 500m, 1000m)",
	)
	flags.StringArrayVar(
		&jobUpdateInput.EnvVars,
		"env",
		nil,
		"Replace environment variables (NAME=VALUE); repeat for every entry to keep; dropping entries requires --force",
	)
	flags.StringArrayVar(
		&jobUpdateInput.SecretRefs,
		"secret",
		nil,
		"Replace secret references; repeat for every entry to keep; dropping entries requires --force",
	)
	flags.StringVar(&jobUpdateInput.StackId, "stack-id", "", "Stack ID to assign the job to")
	flags.StringVar(
		&jobUpdateInput.Schedule,
		"schedule",
		"",
		"Five-field cron schedule or calendar shortcut (e.g. '0 2 * * *', '@daily')",
	)
	flags.StringVar(
		&jobUpdateInput.Timezone,
		"schedule-timezone",
		"",
		"IANA timezone for the schedule (e.g. Europe/Berlin, US/Eastern, UTC)",
	)
	flags.Int64Var(
		&jobUpdateTimeout,
		"timeout",
		0,
		"Run time budget in seconds (1-21600, max 6h), including startup, dependency installation, and all retries; unchanged when omitted",
	)
	flags.Int32Var(
		&jobUpdateRetries,
		"retries",
		0,
		"Maximum retries after failure; share the run timeout and may repeat side effects; unchanged when omitted",
	)
	flags.Int32Var(
		&jobUpdateRetentionTTL,
		"retention-ttl",
		0,
		"Seconds to keep finished runs (0-2592000); zero requests immediate cleanup; unchanged when omitted",
	)
	flags.Int32Var(
		&jobUpdateSuccessfulRuns,
		"retention-successful-runs",
		0,
		"Successful runs to retain (0-20); unchanged when omitted",
	)
	flags.Int32Var(
		&jobUpdateFailedRuns,
		"retention-failed-runs",
		0,
		"Failed, timed-out, or stopped runs to retain (0-20); unchanged when omitted",
	)
	flags.BoolVar(
		&jobUpdateInput.ClearEnv,
		"clear-env",
		false,
		"Remove all environment variables from the job",
	)
	flags.BoolVar(
		&jobUpdateInput.ClearSecret,
		"clear-secret",
		false,
		"Remove all secret references from the job",
	)
	flags.BoolVar(
		&jobUpdateInput.ClearCommand,
		"clear-command",
		false,
		"Remove the container entrypoint override",
	)
	flags.BoolVar(
		&jobUpdateInput.ClearArgs,
		"clear-args",
		false,
		"Remove the container arguments override",
	)
	flags.BoolVar(
		&jobUpdateInput.ClearSchedule,
		"clear-schedule",
		false,
		"Remove the schedule configuration from the job",
	)
	flags.BoolVar(
		&jobUpdateInput.ClearRetention,
		"clear-retention",
		false,
		"Reset retention to its defaults",
	)
	flags.BoolVar(
		&jobUpdateInput.ClearStackId,
		"clear-stack-id",
		false,
		"Remove the job from its stack",
	)
	flags.IntVar(
		&jobExpectRevision,
		"expect-revision",
		0,
		"Require this live revision before updating; also fail if the revision cannot be fetched",
	)
	flags.BoolVar(
		&jobUpdateForce,
		"force",
		false,
		"Allow --env/--secret to drop live entries; --clear-env/--clear-secret need no override. Drop checks are skipped if live state cannot be fetched",
	)
	indexFlag(jobUpdateCmd, "env", "replaces")
	indexFlag(jobUpdateCmd, "secret", "replaces")
	jobsCmd.AddCommand(jobUpdateCmd)
}
