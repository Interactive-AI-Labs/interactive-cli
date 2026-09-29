package inputs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/google/go-cmp/cmp"
)

func TestValidateJobArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		err  string
	}{
		{"list", nil, ""},
		{"job", []string{"report"}, ""},
		{"run", []string{"report", "run-1"}, ""},
		{"empty job", []string{""}, "arguments must not be empty"},
		{"blank job", []string{" \t"}, "arguments must not be empty"},
		{"empty run", []string{"report", ""}, "arguments must not be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJobArgs(tt.args)
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != tt.err {
				t.Fatalf("error = %q, want %q", message, tt.err)
			}
		})
	}
}

func TestReadJobFile(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		files map[string]string
		want  string
		err   string
	}{
		{name: "omitted path"},
		{name: "empty file", path: "empty.py", files: map[string]string{"empty.py": ""}},
		{
			name:  "UTF-8 and whitespace preserved",
			path:  "main.py",
			files: map[string]string{"main.py": "# café\r\n\tprint('你好')\r\n"},
			want:  "# café\r\n\tprint('你好')\r\n",
		},
		{
			name:  "large file is read without truncation",
			path:  "main.py",
			files: map[string]string{"main.py": strings.Repeat("é", 200_000)},
			want:  strings.Repeat("é", 200_000),
		},
		{
			name:  "invalid UTF-8 script",
			path:  "main.py",
			files: map[string]string{"main.py": "\xff"},
			err:   `job file "{dir}/main.py" must contain valid UTF-8`,
		},
		{
			name:  "invalid UTF-8 project",
			path:  "pyproject.toml",
			files: map[string]string{"pyproject.toml": "\xff"},
			err:   `job file "{dir}/pyproject.toml" must contain valid UTF-8`,
		},
		{
			name: "missing file",
			path: "missing.py",
			err:  `failed to read job file "{dir}/missing.py": open {dir}/missing.py: no such file or directory`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, contents := range tt.files {
				if err := os.WriteFile(
					filepath.Join(dir, name),
					[]byte(contents),
					0o600,
				); err != nil {
					t.Fatal(err)
				}
			}
			path := tt.path
			if path != "" {
				path = filepath.Join(dir, path)
			}
			got, err := ReadJobFile(path)
			message := ""
			if err != nil {
				message = err.Error()
			}
			if want := strings.ReplaceAll(tt.err, "{dir}", dir); message != want {
				t.Fatalf("error = %q, want %q", message, want)
			}
			if got != tt.want {
				t.Fatal("file contents differ from the expected bytes")
			}
		})
	}
}

func TestBuildJobRequestBody(t *testing.T) {
	tests := []struct {
		name string
		in   JobInput
		want deployment.CreateJobBody
		err  string
	}{
		{
			name: "image with resources and arguments",
			in: JobInput{
				Type:            "image",
				ImageType:       "external",
				ImageRepository: "docker.io",
				ImageName:       "python",
				ImageTag:        "3.12",
				CPU:             "500m",
				Memory:          "512M",
				Command:         []string{"python"},
				Args:            []string{"-c", "print('a,b')"},
				StackId:         "batch",
			},
			want: deployment.CreateJobBody{
				Type: "image",
				Image: &deployment.ImageSpec{
					Type:       "external",
					Repository: "docker.io",
					Name:       "python",
					Tag:        "3.12",
				},
				Resources: deployment.Resources{CPU: "500m", Memory: "512M"},
				Command:   []string{"python"},
				Args:      []string{"-c", "print('a,b')"},
				StackId:   "batch",
			},
		},
		{
			name: "script contents are preserved",
			in: JobInput{
				Type:      "script",
				Script:    "# café\r\nprint('你好')\r\n",
				Pyproject: "[project]\nname = 'report'\n",
				CPU:       "1",
				Memory:    "1G",
			},
			want: deployment.CreateJobBody{
				Type:      "script",
				Script:    "# café\r\nprint('你好')\r\n",
				Pyproject: "[project]\nname = 'report'\n",
				Resources: deployment.Resources{CPU: "1", Memory: "1G"},
			},
		},
		{
			name: "env and secret parsing matches services",
			in: JobInput{
				Type:       "image",
				EnvVars:    []string{" TOKEN =a=b", "EMPTY=", "UV_CACHE_DIR=/tmp/custom"},
				SecretRefs: []string{" credentials "},
			},
			want: deployment.CreateJobBody{
				Type:  "image",
				Image: &deployment.ImageSpec{},
				Env: []deployment.EnvVar{
					{Name: "TOKEN", Value: "a=b"},
					{Name: "EMPTY", Value: ""},
					{Name: "UV_CACHE_DIR", Value: "/tmp/custom"},
				},
				SecretRefs: []deployment.SecretRef{{SecretName: "credentials"}},
			},
		},
		{
			name: "schedule leaves timezone default to server",
			in:   JobInput{Type: "image", Schedule: "0 2 * * *"},
			want: deployment.CreateJobBody{
				Type:  "image",
				Image: &deployment.ImageSpec{},
				Cron:  &deployment.JobCron{Schedule: "0 2 * * *"},
			},
		},
		{
			name: "explicit options and zero retention are preserved",
			in: JobInput{
				Type:     "image",
				Schedule: "@daily",
				Timezone: "Europe/Berlin",
				Timeout:  utils.ToPtr(int64(120)),
				Retries:  utils.ToPtr(int32(0)),
				Retention: deployment.JobRetention{
					TTL:            utils.ToPtr(int32(0)),
					SuccessfulRuns: utils.ToPtr(int32(0)),
					FailedRuns:     utils.ToPtr(int32(0)),
				},
			},
			want: deployment.CreateJobBody{
				Type:    "image",
				Image:   &deployment.ImageSpec{},
				Cron:    &deployment.JobCron{Schedule: "@daily", Timezone: "Europe/Berlin"},
				Timeout: utils.ToPtr(int64(120)),
				Retries: utils.ToPtr(int32(0)),
				Retention: &deployment.JobRetention{
					TTL:            utils.ToPtr(int32(0)),
					SuccessfulRuns: utils.ToPtr(int32(0)),
					FailedRuns:     utils.ToPtr(int32(0)),
				},
			},
		},
		{
			name: "partial retention leaves omitted fields to server",
			in: JobInput{
				Type:      "image",
				Retention: deployment.JobRetention{FailedRuns: utils.ToPtr(int32(20))},
			},
			want: deployment.CreateJobBody{
				Type:      "image",
				Image:     &deployment.ImageSpec{},
				Retention: &deployment.JobRetention{FailedRuns: utils.ToPtr(int32(20))},
			},
		},
		{
			name: "large script contents are passed to the server",
			in: JobInput{
				Type:      "script",
				Script:    strings.Repeat("é", 200_000),
				Pyproject: "[project]\n",
			},
			want: deployment.CreateJobBody{
				Type:      "script",
				Script:    strings.Repeat("é", 200_000),
				Pyproject: "[project]\n",
			},
		},
		{name: "missing type", in: JobInput{}, err: "--type must be image or script"},
		{name: "unknown type", in: JobInput{Type: "other"}, err: "--type must be image or script"},
		{
			name: "image with script",
			in:   JobInput{Type: "image", Script: "print(1)"},
			err:  "--script and --pyproject are only available for script jobs",
		},
		{
			name: "image with project",
			in:   JobInput{Type: "image", Pyproject: "[project]"},
			err:  "--script and --pyproject are only available for script jobs",
		},
		{
			name: "script with image flags",
			in:   JobInput{Type: "script", ImageName: "python"},
			err:  "image, command and args flags are only available for image jobs",
		},
		{
			name: "script with command",
			in:   JobInput{Type: "script", Command: []string{"sh"}},
			err:  "image, command and args flags are only available for image jobs",
		},
		{
			name: "script with args",
			in:   JobInput{Type: "script", Args: []string{"hello"}},
			err:  "image, command and args flags are only available for image jobs",
		},
		{
			name: "script missing project",
			in:   JobInput{Type: "script", Script: "print(1)"},
			err:  "--script and --pyproject are required for script jobs",
		},
		{
			name: "whitespace script",
			in:   JobInput{Type: "script", Script: " \n\t", Pyproject: "[project]"},
			err:  "--script and --pyproject are required for script jobs",
		},
		{
			name: "invalid env",
			in:   JobInput{Type: "image", EnvVars: []string{"BAD"}},
			err:  "invalid --env value \"BAD\"; expected NAME=VALUE",
		},
		{
			name: "empty secret",
			in:   JobInput{Type: "image", SecretRefs: []string{" "}},
			err:  "invalid --secret value \" \"; name must not be empty",
		},
		{
			name: "timezone without schedule",
			in:   JobInput{Type: "image", Timezone: "UTC"},
			err:  "--schedule-timezone requires --schedule",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildJobRequestBody(tt.in)
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != tt.err {
				t.Fatalf("error = %q, want %q", message, tt.err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
