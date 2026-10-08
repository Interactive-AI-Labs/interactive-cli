package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/google/go-cmp/cmp"
)

func TestLoadStackConfigJobs(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string]string // relative to the temp dir; {dir} is the temp dir
		config string
		want   map[string]JobConfig
		err    string
	}{
		{
			name: "image job needs no files",
			config: `stack-id: batch
jobs:
  backup:
    type: image
    image: {type: external, repository: docker.io, name: postgres, tag: "16"}
    command: [pg_dump]
    args: [--format=custom]
    resources: {cpu: "1", memory: 1G}
    cron: {schedule: "0 2 * * *", timezone: Europe/Berlin}
    timeout: 1800
    retries: 0
    retention: {ttl: 3600, successfulRuns: 2, failedRuns: 3}
`,
			want: map[string]JobConfig{
				"backup": {
					Type: "image",
					Image: &deployment.ImageSpec{
						Type:       "external",
						Repository: "docker.io",
						Name:       "postgres",
						Tag:        "16",
					},
					Command:   []string{"pg_dump"},
					Args:      []string{"--format=custom"},
					Resources: deployment.Resources{CPU: "1", Memory: "1G"},
					Cron: &deployment.JobCron{
						Schedule: "0 2 * * *",
						Timezone: "Europe/Berlin",
					},
					Timeout: utils.ToPtr(int64(1800)),
					Retries: utils.ToPtr(int32(0)),
					Retention: &deployment.JobRetention{
						TTL:            utils.ToPtr(int32(3600)),
						SuccessfulRuns: utils.ToPtr(int32(2)),
						FailedRuns:     utils.ToPtr(int32(3)),
					},
				},
			},
		},
		{
			name: "script files resolve against the config file directory",
			files: map[string]string{
				"infra/jobs/report/main.py":        "print('hi')\r\n",
				"infra/jobs/report/pyproject.toml": "[project]\n",
			},
			config: `stack-id: batch
jobs:
  report:
    type: script
    scriptFile: jobs/report/main.py
    pyprojectFile: ./jobs/report/pyproject.toml
    resources: {cpu: "0.5", memory: 512M}
`,
			want: map[string]JobConfig{
				"report": {
					Type:          "script",
					ScriptFile:    "jobs/report/main.py",
					PyprojectFile: "./jobs/report/pyproject.toml",
					Script:        "print('hi')\r\n",
					Pyproject:     "[project]\n",
					Resources:     deployment.Resources{CPU: "0.5", Memory: "512M"},
				},
			},
		},
		{
			name: "absolute script paths are used as is",
			files: map[string]string{
				"shared/main.py":        "print(1)\n",
				"shared/pyproject.toml": "[project]\n",
			},
			config: `stack-id: batch
jobs:
  report:
    type: script
    scriptFile: {dir}/shared/main.py
    pyprojectFile: ../shared/pyproject.toml
    resources: {cpu: "1", memory: 1G}
`,
			want: map[string]JobConfig{
				"report": {
					Type:          "script",
					ScriptFile:    "{dir}/shared/main.py",
					PyprojectFile: "../shared/pyproject.toml",
					Script:        "print(1)\n",
					Pyproject:     "[project]\n",
					Resources:     deployment.Resources{CPU: "1", Memory: "1G"},
				},
			},
		},
		{
			name: "script job without files",
			config: `stack-id: batch
jobs:
  report:
    type: script
    scriptFile: main.py
`,
			err: `job "report": scriptFile and pyprojectFile are required for script jobs`,
		},
		{
			name: "image job with files",
			config: `stack-id: batch
jobs:
  backup:
    type: image
    scriptFile: main.py
`,
			err: `job "backup": scriptFile and pyprojectFile are only available for script jobs`,
		},
		{
			name: "missing script file",
			config: `stack-id: batch
jobs:
  report:
    type: script
    scriptFile: missing.py
    pyprojectFile: pyproject.toml
`,
			err: `job "report": failed to read job file "{dir}/infra/missing.py": open {dir}/infra/missing.py: no such file or directory`,
		},
		{
			name: "invalid script encoding fails before JSON can replace bytes",
			files: map[string]string{
				"infra/main.py":        "\xff",
				"infra/pyproject.toml": "[project]\n",
			},
			config: `stack-id: batch
jobs:
  report:
    type: script
    scriptFile: main.py
    pyprojectFile: pyproject.toml
`,
			err: `job "report": job file "{dir}/infra/main.py" must contain valid UTF-8`,
		},
		{
			name: "jobs require a stack id",
			config: `jobs:
  backup:
    type: image
`,
			err: "stack-id is required when services, agents, databases, mcps, or jobs are defined in config file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for path, contents := range tt.files {
				path = filepath.Join(dir, path)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			configPath := filepath.Join(dir, "infra", "stack.yaml")
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				t.Fatal(err)
			}
			config := strings.ReplaceAll(tt.config, "{dir}", dir)
			if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
				t.Fatal(err)
			}

			cfg, err := LoadStackConfig(configPath)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if wantErr := strings.ReplaceAll(tt.err, "{dir}", dir); gotErr != wantErr {
				t.Fatalf("error = %q, want %q", gotErr, wantErr)
			}
			if err != nil {
				return
			}
			want := make(map[string]JobConfig, len(tt.want))
			for name, job := range tt.want {
				job.ScriptFile = strings.ReplaceAll(job.ScriptFile, "{dir}", dir)
				want[name] = job
			}
			if diff := cmp.Diff(want, cfg.Jobs); diff != "" {
				t.Fatalf("jobs mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestJobConfigToCreateRequest(t *testing.T) {
	tests := []struct {
		name    string
		job     JobConfig
		stackId string
		want    deployment.CreateJobBody
	}{
		{
			name: "script contents are sent without file paths",
			job: JobConfig{
				Type:          "script",
				ScriptFile:    "main.py",
				PyprojectFile: "pyproject.toml",
				Script:        "print(1)\n",
				Pyproject:     "[project]\n",
				Resources:     deployment.Resources{CPU: "1", Memory: "1G"},
				Env:           []deployment.EnvVar{{Name: "LEVEL", Value: "info"}},
				SecretRefs:    []deployment.SecretRef{{SecretName: "token"}},
				Cron:          &deployment.JobCron{Schedule: "0 2 * * *"},
				Timeout:       utils.ToPtr(int64(60)),
				Retries:       utils.ToPtr(int32(1)),
				Retention:     &deployment.JobRetention{TTL: utils.ToPtr(int32(0))},
			},
			stackId: "batch",
			want: deployment.CreateJobBody{
				Type:       "script",
				Script:     "print(1)\n",
				Pyproject:  "[project]\n",
				Resources:  deployment.Resources{CPU: "1", Memory: "1G"},
				Env:        []deployment.EnvVar{{Name: "LEVEL", Value: "info"}},
				SecretRefs: []deployment.SecretRef{{SecretName: "token"}},
				StackId:    "batch",
				Cron:       &deployment.JobCron{Schedule: "0 2 * * *"},
				Timeout:    utils.ToPtr(int64(60)),
				Retries:    utils.ToPtr(int32(1)),
				Retention:  &deployment.JobRetention{TTL: utils.ToPtr(int32(0))},
			},
		},
		{
			name: "image job",
			job: JobConfig{
				Type:      "image",
				Image:     &deployment.ImageSpec{Type: "internal", Name: "app", Tag: "v1"},
				Command:   []string{"run"},
				Args:      []string{"--once"},
				Resources: deployment.Resources{CPU: "1", Memory: "1G"},
			},
			stackId: "batch",
			want: deployment.CreateJobBody{
				Type:      "image",
				Image:     &deployment.ImageSpec{Type: "internal", Name: "app", Tag: "v1"},
				Command:   []string{"run"},
				Args:      []string{"--once"},
				Resources: deployment.Resources{CPU: "1", Memory: "1G"},
				StackId:   "batch",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, tt.job.ToCreateRequest(tt.stackId)); diff != "" {
				t.Fatalf("body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestJobConfigFromRequest(t *testing.T) {
	tests := []struct {
		name string
		job  deployment.DescribeJobResponse
		want JobConfig
	}{
		{
			name: "status, revision and stack are dropped",
			job: deployment.DescribeJobResponse{
				JobOutput: deployment.JobOutput{Name: "report", Revision: 4, Status: "Ready"},
				CreateJobBody: deployment.CreateJobBody{
					Type:      "script",
					Script:    "print(1)\n",
					Pyproject: "[project]\n",
					Resources: deployment.Resources{CPU: "1", Memory: "1G"},
					StackId:   "batch",
					Cron:      &deployment.JobCron{Schedule: "0 2 * * *", Timezone: "UTC"},
					Timeout:   utils.ToPtr(int64(3600)),
					Retries:   utils.ToPtr(int32(0)),
					Retention: &deployment.JobRetention{FailedRuns: utils.ToPtr(int32(10))},
				},
				Message: "Job is ready",
			},
			want: JobConfig{
				Type:      "script",
				Script:    "print(1)\n",
				Pyproject: "[project]\n",
				Resources: deployment.Resources{CPU: "1", Memory: "1G"},
				Cron:      &deployment.JobCron{Schedule: "0 2 * * *", Timezone: "UTC"},
				Timeout:   utils.ToPtr(int64(3600)),
				Retries:   utils.ToPtr(int32(0)),
				Retention: &deployment.JobRetention{FailedRuns: utils.ToPtr(int32(10))},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, JobConfigFromRequest(tt.job.CreateJobBody)); diff != "" {
				t.Fatalf("config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestWriteJobFiles(t *testing.T) {
	tests := []struct {
		name      string
		jobs      map[string]JobConfig
		wantJobs  map[string]JobConfig
		wantFiles map[string]string
	}{
		{
			name: "script jobs get files and image jobs are unchanged",
			jobs: map[string]JobConfig{
				"report": {Type: "script", Script: "print(1)\r\n", Pyproject: "[project]\n"},
				"backup": {Type: "image", Command: []string{"pg_dump"}},
			},
			wantJobs: map[string]JobConfig{
				"report": {
					Type:          "script",
					ScriptFile:    "jobs/report/main.py",
					PyprojectFile: "jobs/report/pyproject.toml",
					Script:        "print(1)\r\n",
					Pyproject:     "[project]\n",
				},
				"backup": {Type: "image", Command: []string{"pg_dump"}},
			},
			wantFiles: map[string]string{
				"jobs/report/main.py":        "print(1)\r\n",
				"jobs/report/pyproject.toml": "[project]\n",
			},
		},
		{
			name:      "no jobs",
			jobs:      map[string]JobConfig{},
			wantJobs:  map[string]JobConfig{},
			wantFiles: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := &StackConfig{Jobs: tt.jobs}
			if err := WriteJobFiles(cfg, dir); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.wantJobs, cfg.Jobs); diff != "" {
				t.Fatalf("jobs mismatch (-want +got):\n%s", diff)
			}
			gotFiles := map[string]string{}
			err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
				if err != nil || entry.IsDir() {
					return err
				}
				data, err := os.ReadFile(path)
				rel, _ := filepath.Rel(dir, path)
				gotFiles[rel] = string(data)
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.wantFiles, gotFiles); diff != "" {
				t.Fatalf("files mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestScriptJobNames(t *testing.T) {
	tests := []struct {
		name string
		jobs map[string]JobConfig
		want []string
	}{
		{name: "no jobs", jobs: map[string]JobConfig{}},
		{
			name: "only script jobs, sorted",
			jobs: map[string]JobConfig{
				"report":  {Type: "script"},
				"backup":  {Type: "image"},
				"cleanup": {Type: "script"},
			},
			want: []string{"cleanup", "report"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScriptJobNames(&StackConfig{Jobs: tt.jobs})
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("names mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDiffStackConfigsJobs(t *testing.T) {
	report := JobConfig{
		Type:      "script",
		Script:    "print(1)\n",
		Pyproject: "[project]\nname = \"r\"\n",
		Resources: deployment.Resources{CPU: "1", Memory: "1G"},
	}
	changedScript := report
	changedScript.Script = "print(2)\n"

	tests := []struct {
		name        string
		plan        ResourcePlan[JobConfig]
		live        map[string]JobConfig
		want        ResourceTypeDiff
		wantPrinted string
	}{
		{
			name:        "unchanged job",
			plan:        ResourcePlan[JobConfig]{Desired: map[string]JobConfig{"report": report}},
			live:        map[string]JobConfig{"report": report},
			wantPrinted: "No differences found.\n",
		},
		{
			name: "changed script contents show as a digest change",
			plan: ResourcePlan[JobConfig]{
				Desired: map[string]JobConfig{"report": changedScript},
				Changed: map[string]bool{"report": true},
			},
			live: map[string]JobConfig{"report": report},
			want: ResourceTypeDiff{Updated: []ResourceChange{{
				Name: "report",
				Changes: map[string]FieldDiff{"script": {
					Old: "sha256:cc42155088fc (9 B)",
					New: "sha256:0111afd387e1 (9 B)",
				}},
			}}},
			wantPrinted: "Stack: batch\n\n  ~ job report (update)\n" +
				"    script: sha256:cc42155088fc (9 B) → sha256:0111afd387e1 (9 B)\n",
		},
		{
			name:        "created and deleted jobs",
			plan:        ResourcePlan[JobConfig]{Desired: map[string]JobConfig{"new": report}},
			live:        map[string]JobConfig{"old": report},
			want:        ResourceTypeDiff{Created: []string{"new"}, Deleted: []string{"old"}},
			wantPrinted: "Stack: batch\n\n  + job new (create)\n\n  - job old (delete)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan := &StackPlan{StackId: "batch", Jobs: tt.plan}
			live := &StackConfig{StackId: "batch", Jobs: tt.live}
			d := DiffStackConfigs(plan, live)
			if diff := cmp.Diff(tt.want, d.Jobs); diff != "" {
				t.Fatalf("diff mismatch (-want +got):\n%s", diff)
			}
			var out strings.Builder
			if err := PrintStackDiffDetailed(&out, plan, live, d); err != nil {
				t.Fatal(err)
			}
			if got := out.String(); got != tt.wantPrinted {
				t.Fatalf("printed = %q, want %q", got, tt.wantPrinted)
			}
		})
	}
}
