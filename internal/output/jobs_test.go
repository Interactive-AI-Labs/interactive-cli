package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
)

func TestPrintJobList(t *testing.T) {
	tests := []struct {
		name string
		jobs []deployment.JobOutput
		want string
	}{
		{name: "empty", want: "No jobs found.\n"},
		{
			name: "jobs use service list columns",
			jobs: []deployment.JobOutput{
				{Name: "report", Revision: 1, Status: "Ready", Updated: "2026-01-15"},
			},
			want: "NAME     REVISION   STATUS   UPDATED\nreport   1          Ready    2026-01-15\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintJobList(&out, tt.jobs); err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintJobDescribe(t *testing.T) {
	tests := []struct {
		name string
		job  deployment.DescribeJobResponse
		want string
	}{
		{
			name: "image",
			job: deployment.DescribeJobResponse{
				JobOutput: deployment.JobOutput{Name: "report", Revision: 1, Status: "Ready"},
				CreateJobBody: deployment.CreateJobBody{
					Type:      "image",
					Image:     &deployment.ImageSpec{Type: "internal", Name: "app", Tag: "v1"},
					Resources: deployment.Resources{CPU: "1", Memory: "1G"},
				},
			},
			want: "Name:       report\nRevision:   1\nStatus:     Ready\nType:       image\nImage:\n  Type:   internal\n  Name:   app\n  Tag:    v1\nResources:\n  CPU:      1\n  Memory:   1G\n",
		},
		{
			name: "command and args are shell quoted beside top-level fields",
			job: deployment.DescribeJobResponse{
				JobOutput: deployment.JobOutput{Name: "backup", Revision: 1, Status: "Ready"},
				CreateJobBody: deployment.CreateJobBody{
					Type: "image",
					Image: &deployment.ImageSpec{
						Type:       "external",
						Repository: "docker.io",
						Name:       "python",
						Tag:        "3.12",
					},
					Command:   []string{"python"},
					Args:      []string{"-c", "print('hi there')", "--file=/tmp/a.txt", ""},
					Resources: deployment.Resources{CPU: "1", Memory: "1G"},
				},
			},
			want: "Name:       backup\nRevision:   1\nStatus:     Ready\nType:       image\n" +
				"Command:    python\nArgs:       -c 'print('\\''hi there'\\'')' --file=/tmp/a.txt ''\n" +
				"Image:\n  Type:         external\n  Name:         python\n  Tag:          3.12\n  Repository:   docker.io\n" +
				"Resources:\n  CPU:      1\n  Memory:   1G\n",
		},
		{
			name: "scheduled script with zero retries and retention",
			job: deployment.DescribeJobResponse{
				JobOutput: deployment.JobOutput{
					Name:     "report",
					Revision: 1,
					Status:   "Deactivated",
					Updated:  "2026-01-15",
				},
				Message: "Job is deactivated",
				CreateJobBody: deployment.CreateJobBody{
					Type:      "script",
					StackId:   "batch",
					Script:    "print('hi')",
					Pyproject: "[project]\nname = \"report\"\n",
					Resources: deployment.Resources{CPU: "1", Memory: "1G"},
					Timeout:   utils.ToPtr(int64(120)),
					Retries:   utils.ToPtr(int32(0)),
					Cron:      &deployment.JobCron{Schedule: "0 2 * * *", Timezone: "UTC"},
					Retention: &deployment.JobRetention{
						TTL:            utils.ToPtr(int32(0)),
						SuccessfulRuns: utils.ToPtr(int32(0)),
						FailedRuns:     utils.ToPtr(int32(0)),
					},
					Env: []deployment.EnvVar{
						{Name: "TOKEN", Value: "abc"},
					},
					SecretRefs: []deployment.SecretRef{{SecretName: "credentials"}},
				},
			},
			want: "Name:       report\nStack Id:   batch\nRevision:   1\nStatus:     Deactivated\nMessage:    Job is deactivated\nUpdated:    2026-01-15\nType:       script\nResources:\n  CPU:      1\n  Memory:   1G\nTimeout:    120s\nRetries:    0\nSchedule:\n  Cron:       0 2 * * *\n  Timezone:   UTC\nRetention:\n  TTL:               0s\n  Successful Runs:   0\n  Failed Runs:       0\n\nEnvironment:\n  TOKEN=abc\n\nSecrets:   credentials\n\nFiles:\n  Script:      1 line, 11 B\n  Pyproject:   2 lines, 26 B\n",
		},
		{
			name: "large script uses KiB and a missing project file shows zero",
			job: deployment.DescribeJobResponse{
				JobOutput: deployment.JobOutput{Name: "report", Revision: 1, Status: "Unhealthy"},
				CreateJobBody: deployment.CreateJobBody{
					Type:      "script",
					Script:    string(bytes.Repeat([]byte("x = 1\n"), 2400)),
					Resources: deployment.Resources{CPU: "1", Memory: "1G"},
				},
			},
			want: "Name:       report\nRevision:   1\nStatus:     Unhealthy\nType:       script\nResources:\n  CPU:      1\n  Memory:   1G\n\nFiles:\n  Script:      2400 lines, 14.1 KiB\n  Pyproject:   0 lines, 0 B\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintJobDescribe(&out, &tt.job); err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintJobRunList(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")
	now := time.Date(2026, 1, 15, 10, 5, 30, 0, time.UTC)
	tests := []struct {
		name string
		runs []deployment.JobRun
		want string
	}{
		{name: "empty", want: "No runs found.\n"},
		{
			name: "pending run has no start or duration",
			runs: []deployment.JobRun{
				{RunId: "run-1", Status: "Pending", Created: "2026-01-15T10:00:00Z"},
			},
			want: "RUN ID   STATUS    STARTED   DURATION\n" +
				"run-1    Pending             \n",
		},
		{
			name: "running run is measured to now and finished run to its finish",
			runs: []deployment.JobRun{
				{
					RunId:   "run-2",
					Status:  "Running",
					Created: "2026-01-15T10:00:00Z",
					Started: "2026-01-15T10:00:04Z",
				},
				{
					RunId:    "run-1",
					Status:   "Succeeded",
					Created:  "2026-01-15T09:00:00Z",
					Started:  "2026-01-15T09:00:03Z",
					Finished: "2026-01-15T09:01:57Z",
				},
			},
			want: "RUN ID   STATUS      STARTED                   DURATION\n" +
				"run-2    Running     2026-01-15 11:00:04 CET   5m26s\n" +
				"run-1    Succeeded   2026-01-15 10:00:03 CET   1m54s\n",
		},
		{
			name: "running duration is zero when the client clock trails the server",
			runs: []deployment.JobRun{
				{RunId: "run-1", Status: "Running", Started: "2026-01-15T10:05:34Z"},
			},
			want: "RUN ID   STATUS    STARTED                   DURATION\n" +
				"run-1    Running   2026-01-15 11:05:34 CET   0s\n",
		},
		{
			name: "timeout has a human readable label",
			runs: []deployment.JobRun{
				{
					RunId: "run-1", Status: "TimedOut", Created: "2026-01-15T09:00:00Z",
					Started: "2026-01-15T09:00:03Z", Finished: "2026-01-15T09:01:57Z",
				},
			},
			want: "RUN ID   STATUS      STARTED                   DURATION\n" +
				"run-1    Timed out   2026-01-15 10:00:03 CET   1m54s\n",
		},
		{
			name: "run that failed before starting has no duration",
			runs: []deployment.JobRun{
				{
					RunId:    "run-1",
					Status:   "Failed",
					Created:  "2026-01-15T09:00:00Z",
					Finished: "2026-01-15T09:00:10Z",
				},
			},
			want: "RUN ID   STATUS   STARTED   DURATION\n" +
				"run-1    Failed             \n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintJobRunList(&out, tt.runs, now); err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintJobRunDescribe(t *testing.T) {
	t.Setenv("TZ", "Europe/Madrid")
	now := time.Date(2026, 1, 15, 10, 5, 30, 0, time.UTC)
	tests := []struct {
		name string
		run  deployment.JobRun
		want string
	}{
		{
			name: "pending",
			run: deployment.JobRun{
				RunId:   "run-1",
				JobName: "report",
				Status:  "Pending",
				Created: "2026-01-15",
			},
			want: "Run Id:   run-1\nJob:      report\nStatus:   Pending\n",
		},
		{
			name: "timeout with server explanation",
			run: deployment.JobRun{
				RunId:    "run-1",
				JobName:  "report",
				Status:   "TimedOut",
				Message:  "The run exceeded its configured timeout.",
				Created:  "2026-01-15T09:00:00Z",
				Started:  "2026-01-15T09:00:03Z",
				Finished: "2026-01-15T09:01:57Z",
			},
			want: "Run Id:     run-1\nJob:        report\nStatus:     Timed out\nMessage:    The run exceeded its configured timeout.\nStarted:    2026-01-15 10:00:03 CET\nDuration:   1m54s\n",
		},
		{
			name: "termination is unfinished and shows the server message",
			run: deployment.JobRun{
				RunId: "run-1", JobName: "report", Status: "Terminating",
				Message: "A stop was requested; termination is pending.", Created: "2026-01-15",
			},
			want: "Run Id:    run-1\nJob:       report\nStatus:    Terminating\nMessage:   A stop was requested; termination is pending.\n",
		},
		{
			name: "running duration uses the supplied current time",
			run: deployment.JobRun{
				RunId: "run-1", JobName: "report", Status: "Running",
				Created: "2026-01-15T10:00:00Z", Started: "2026-01-15T10:00:04Z",
			},
			want: "Run Id:     run-1\nJob:        report\nStatus:     Running\nStarted:    2026-01-15 11:00:04 CET\nDuration:   5m26s\n",
		},
		{
			name: "running duration is zero when the client clock trails the server",
			run: deployment.JobRun{
				RunId: "run-1", JobName: "report", Status: "Running",
				Started: "2026-01-15T10:05:34Z",
			},
			want: "Run Id:     run-1\nJob:        report\nStatus:     Running\nStarted:    2026-01-15 11:05:34 CET\nDuration:   0s\n",
		},
		{
			name: "finished",
			run: deployment.JobRun{
				RunId:    "run-1",
				JobName:  "report",
				Status:   "Succeeded",
				Created:  "2026-01-15T09:00:00Z",
				Started:  "2026-01-15T09:00:03Z",
				Finished: "2026-01-15T09:01:57Z",
			},
			want: "Run Id:     run-1\nJob:        report\nStatus:     Succeeded\nStarted:    2026-01-15 10:00:03 CET\nDuration:   1m54s\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintJobRunDescribe(&out, &tt.run, now); err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintJobFile(t *testing.T) {
	script := deployment.DescribeJobResponse{
		CreateJobBody: deployment.CreateJobBody{
			Type:      "script",
			Script:    "# café\r\nprint('hi')",
			Pyproject: "[project]\n",
		},
	}
	tests := []struct {
		name      string
		job       deployment.DescribeJobResponse
		pyproject bool
		want      string
		err       string
	}{
		{name: "script is written unchanged", job: script, want: "# café\r\nprint('hi')"},
		{
			name:      "project file is written unchanged",
			job:       script,
			pyproject: true,
			want:      "[project]\n",
		},
		{
			name: "image job has no files",
			job: deployment.DescribeJobResponse{
				CreateJobBody: deployment.CreateJobBody{Type: "image"},
			},
			err: "--script and --pyproject are only available for script jobs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := PrintJobFile(&out, &tt.job, tt.pyproject)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.err {
				t.Fatalf("error = %q, want %q", gotErr, tt.err)
			}
			if got := out.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
