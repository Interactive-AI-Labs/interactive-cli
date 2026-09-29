package output

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func PrintJobList(out io.Writer, jobs []deployment.JobOutput) error {
	if len(jobs) == 0 {
		fmt.Fprintln(out, "No jobs found.")
		return nil
	}
	rows := make([][]string, len(jobs))
	for i, job := range jobs {
		rows[i] = []string{
			job.Name,
			fmt.Sprintf("%d", job.Revision),
			job.Status,
			LocalTime(job.Updated),
		}
	}
	return PrintTable(out, []string{"NAME", "REVISION", "STATUS", "UPDATED"}, rows)
}

func PrintJobDescribe(out io.Writer, job *deployment.DescribeJobResponse) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", job.Name)
	if job.StackId != "" {
		fmt.Fprintf(w, "Stack Id:\t%s\n", job.StackId)
	}
	fmt.Fprintf(w, "Revision:\t%d\n", job.Revision)
	fmt.Fprintf(w, "Status:\t%s\n", job.Status)
	if job.Message != "" {
		fmt.Fprintf(w, "Message:\t%s\n", job.Message)
	}
	if job.Updated != "" {
		fmt.Fprintf(w, "Updated:\t%s\n", LocalTime(job.Updated))
	}
	fmt.Fprintf(w, "Type:\t%s\n", job.Type)
	// Printed before Image so the tabwriter aligns them with the top-level fields.
	if len(job.Command) > 0 {
		fmt.Fprintf(w, "Command:\t%s\n", shellJoin(job.Command))
	}
	if len(job.Args) > 0 {
		fmt.Fprintf(w, "Args:\t%s\n", shellJoin(job.Args))
	}
	if job.Image != nil {
		fmt.Fprintln(w, "Image:")
		fmt.Fprintf(w, "  Type:\t%s\n", job.Image.Type)
		fmt.Fprintf(w, "  Name:\t%s\n", job.Image.Name)
		fmt.Fprintf(w, "  Tag:\t%s\n", job.Image.Tag)
		if job.Image.Repository != "" {
			fmt.Fprintf(w, "  Repository:\t%s\n", job.Image.Repository)
		}
	}
	fmt.Fprintln(w, "Resources:")
	fmt.Fprintf(w, "  CPU:\t%s\n", job.Resources.CPU)
	fmt.Fprintf(w, "  Memory:\t%s\n", job.Resources.Memory)
	if job.Timeout != nil {
		fmt.Fprintf(w, "Timeout:\t%ds\n", *job.Timeout)
	}
	if job.Retries != nil {
		fmt.Fprintf(w, "Retries:\t%d\n", *job.Retries)
	}
	if job.Cron != nil {
		fmt.Fprintln(w, "Schedule:")
		fmt.Fprintf(w, "  Cron:\t%s\n", job.Cron.Schedule)
		fmt.Fprintf(w, "  Timezone:\t%s\n", job.Cron.Timezone)
	}
	if job.Retention != nil {
		fmt.Fprintln(w, "Retention:")
		if job.Retention.TTL != nil {
			fmt.Fprintf(w, "  TTL:\t%ds\n", *job.Retention.TTL)
		}
		if job.Retention.SuccessfulRuns != nil {
			fmt.Fprintf(w, "  Successful Runs:\t%d\n", *job.Retention.SuccessfulRuns)
		}
		if job.Retention.FailedRuns != nil {
			fmt.Fprintf(w, "  Failed Runs:\t%d\n", *job.Retention.FailedRuns)
		}
	}
	if len(job.Env) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Environment:")
		for _, env := range job.Env {
			fmt.Fprintf(w, "  %s=%s\n", env.Name, formatEnvValue(env))
		}
	}
	if len(job.SecretRefs) > 0 {
		names := make([]string, len(job.SecretRefs))
		for i, ref := range job.SecretRefs {
			names[i] = ref.SecretName
		}
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Secrets:\t%s\n", strings.Join(names, ", "))
	}
	if job.Type == "script" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Files:")
		fmt.Fprintf(w, "  Script:\t%s\n", fileSummary(job.Script))
		fmt.Fprintf(w, "  Pyproject:\t%s\n", fileSummary(job.Pyproject))
	}
	return w.Flush()
}

// PrintJobFile writes the stored script, or the project file, byte for byte.
func PrintJobFile(out io.Writer, job *deployment.DescribeJobResponse, pyproject bool) error {
	if job.Type != "script" {
		return fmt.Errorf("--script and --pyproject are only available for script jobs")
	}
	contents := job.Script
	if pyproject {
		contents = job.Pyproject
	}
	_, err := io.WriteString(out, contents)
	return err
}

// PrintJobRunList measures unfinished runs up to now.
func PrintJobRunList(out io.Writer, runs []deployment.JobRun, now time.Time) error {
	if len(runs) == 0 {
		fmt.Fprintln(out, "No runs found.")
		return nil
	}
	rows := make([][]string, len(runs))
	for i, run := range runs {
		rows[i] = []string{
			run.RunId,
			jobRunStatus(run.Status),
			LocalTime(run.Started),
			runDuration(run, now),
		}
	}
	return PrintTable(
		out,
		[]string{"RUN ID", "STATUS", "STARTED", "DURATION"},
		rows,
	)
}

func PrintJobRunDescribe(out io.Writer, run *deployment.JobRun, now time.Time) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Run Id:\t%s\n", run.RunId)
	fmt.Fprintf(w, "Job:\t%s\n", run.JobName)
	fmt.Fprintf(w, "Status:\t%s\n", jobRunStatus(run.Status))
	if run.Message != "" {
		fmt.Fprintf(w, "Message:\t%s\n", run.Message)
	}
	if run.Started != "" {
		fmt.Fprintf(w, "Started:\t%s\n", LocalTime(run.Started))
		fmt.Fprintf(w, "Duration:\t%s\n", runDuration(*run, now))
	}
	return w.Flush()
}

func jobRunStatus(status string) string {
	if status == "TimedOut" {
		return "Timed out"
	}
	return status
}

var shellSafe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// shellJoin renders arguments the way a POSIX shell would need them typed.
func shellJoin(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		if shellSafe.MatchString(arg) {
			quoted[i] = arg
			continue
		}
		quoted[i] = "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
	}
	return strings.Join(quoted, " ")
}

func fileSummary(contents string) string {
	lines := strings.Count(contents, "\n")
	if contents != "" && !strings.HasSuffix(contents, "\n") {
		lines++
	}
	unit := "lines"
	if lines == 1 {
		unit = "line"
	}
	return fmt.Sprintf("%d %s, %s", lines, unit, HumanBytes(int64(len(contents))))
}

func runDuration(run deployment.JobRun, now time.Time) string {
	started, err := time.Parse(time.RFC3339, run.Started)
	if err != nil {
		return ""
	}
	end := now
	if run.Finished != "" {
		if end, err = time.Parse(time.RFC3339, run.Finished); err != nil {
			return ""
		}
	}
	return max(0, end.Sub(started)).Round(time.Second).String()
}
