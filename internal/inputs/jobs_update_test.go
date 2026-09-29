package inputs

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/google/go-cmp/cmp"
)

func TestBuildJobUpdatePatch(t *testing.T) {
	tests := []struct {
		name  string
		in    JobUpdateInput
		flags []string
		want  map[string]any
		err   string
	}{
		{
			name: "omitted flags leave values unchanged",
			in: JobUpdateInput{
				JobInput: JobInput{
					Type:     "image",
					ImageTag: "ignored",
					Memory:   "1G",
					Timeout:  utils.ToPtr(int64(3600)),
				},
			},
			want: map[string]any{},
		},
		{
			name: "omitted numeric flags do not send their zero defaults",
			in: JobUpdateInput{
				JobInput: JobInput{
					Timeout: utils.ToPtr(int64(0)),
					Retries: utils.ToPtr(int32(0)),
					Retention: deployment.JobRetention{
						TTL:            utils.ToPtr(int32(0)),
						SuccessfulRuns: utils.ToPtr(int32(0)),
						FailedRuns:     utils.ToPtr(int32(0)),
					},
				},
			},
			want: map[string]any{},
		},
		{
			name: "partial image and resources",
			in: JobUpdateInput{
				JobInput: JobInput{ImageTag: "v2", Memory: "1G"},
			},
			flags: []string{"image-tag", "memory"},
			want: map[string]any{
				"image":     map[string]any{"tag": "v2"},
				"resources": map[string]any{"memory": "1G"},
			},
		},
		{
			name: "only script contents are sent",
			in: JobUpdateInput{
				JobInput: JobInput{Script: "# café\r\nprint('你好')\r\n"},
			},
			flags: []string{"script"},
			want:  map[string]any{"script": "# café\r\nprint('你好')\r\n"},
		},
		{
			name: "only project contents are sent",
			in: JobUpdateInput{
				JobInput: JobInput{Pyproject: "[project]\n"},
			},
			flags: []string{"pyproject"},
			want:  map[string]any{"pyproject": "[project]\n"},
		},
		{
			name: "arrays replace without splitting commas",
			in: JobUpdateInput{
				JobInput: JobInput{
					Command:    []string{"python"},
					Args:       []string{"-c", "print('a,b')"},
					EnvVars:    []string{" TOKEN =a=b"},
					SecretRefs: []string{" credentials "},
				},
			},
			flags: []string{"command", "args", "env", "secret"},
			want: map[string]any{
				"command":    []any{"python"},
				"args":       []any{"-c", "print('a,b')"},
				"env":        []any{map[string]any{"name": "TOKEN", "value": "a=b"}},
				"secretRefs": []any{map[string]any{"secretName": "credentials"}},
			},
		},
		{
			name: "timezone can update an existing schedule",
			in: JobUpdateInput{
				JobInput: JobInput{Timezone: "Europe/Madrid"},
			},
			flags: []string{"schedule-timezone"},
			want:  map[string]any{"cron": map[string]any{"timezone": "Europe/Madrid"}},
		},
		{
			name: "schedule and stack",
			in: JobUpdateInput{
				JobInput: JobInput{Schedule: "0 2 * * *", StackId: "batch"},
			},
			flags: []string{"schedule", "stack-id"},
			want: map[string]any{
				"cron":    map[string]any{"schedule": "0 2 * * *"},
				"stackId": "batch",
			},
		},
		{
			name: "explicit zero options are preserved",
			in: JobUpdateInput{
				JobInput: JobInput{
					Timeout: utils.ToPtr(int64(120)),
					Retries: utils.ToPtr(int32(0)),
					Retention: deployment.JobRetention{
						SuccessfulRuns: utils.ToPtr(int32(0)),
						TTL:            utils.ToPtr(int32(0)),
					},
				},
			},
			flags: []string{"timeout", "retries", "retention-successful-runs", "retention-ttl"},
			want: map[string]any{
				"timeout":   float64(120),
				"retries":   float64(0),
				"retention": map[string]any{"successfulRuns": float64(0), "ttl": float64(0)},
			},
		},
		{
			name: "clear optional fields",
			in: JobUpdateInput{
				ClearEnv:       true,
				ClearSecret:    true,
				ClearCommand:   true,
				ClearArgs:      true,
				ClearSchedule:  true,
				ClearRetention: true,
				ClearStackId:   true,
			},
			want: map[string]any{
				"env":        nil,
				"secretRefs": nil,
				"command":    nil,
				"args":       nil,
				"cron":       nil,
				"retention":  nil,
				"stackId":    nil,
			},
		},
		{
			name: "switch to script removes image configuration",
			in: JobUpdateInput{
				JobInput: JobInput{Type: "script", Script: "print(1)", Pyproject: "[project]\n"},
			},
			flags: []string{"type", "script", "pyproject"},
			want: map[string]any{
				"type":      "script",
				"image":     nil,
				"command":   nil,
				"args":      nil,
				"script":    "print(1)",
				"pyproject": "[project]\n",
			},
		},
		{
			name: "switch to image removes script configuration",
			in: JobUpdateInput{
				JobInput: JobInput{
					Type:      "image",
					ImageType: "internal",
					ImageName: "app",
					ImageTag:  "v1",
				},
			},
			flags: []string{"type", "image-type", "image-name", "image-tag"},
			want: map[string]any{
				"type":      "image",
				"script":    nil,
				"pyproject": nil,
				"image":     map[string]any{"type": "internal", "name": "app", "tag": "v1"},
			},
		},
		{
			name:  "invalid type",
			in:    JobUpdateInput{JobInput: JobInput{Type: "unknown"}},
			flags: []string{"type"},
			err:   "--type must be image or script",
		},
		{
			name:  "script type with image flag",
			in:    JobUpdateInput{JobInput: JobInput{Type: "script", ImageTag: "v2"}},
			flags: []string{"type", "image-tag"},
			err:   "image, command and args flags are only available for image jobs",
		},
		{
			name:  "image type with script",
			in:    JobUpdateInput{JobInput: JobInput{Type: "image", Script: "print(1)"}},
			flags: []string{"type", "script"},
			err:   "--script and --pyproject are only available for script jobs",
		},
		{
			name:  "clear env conflicts",
			in:    JobUpdateInput{ClearEnv: true},
			flags: []string{"env"},
			err:   "--clear-env cannot be combined with --env",
		},
		{
			name:  "clear secret conflicts",
			in:    JobUpdateInput{ClearSecret: true},
			flags: []string{"secret"},
			err:   "--clear-secret cannot be combined with --secret",
		},
		{
			name:  "clear command conflicts",
			in:    JobUpdateInput{ClearCommand: true},
			flags: []string{"command"},
			err:   "--clear-command cannot be combined with --command",
		},
		{
			name:  "clear args conflicts",
			in:    JobUpdateInput{ClearArgs: true},
			flags: []string{"args"},
			err:   "--clear-args cannot be combined with --args",
		},
		{
			name:  "clear stack conflicts",
			in:    JobUpdateInput{ClearStackId: true},
			flags: []string{"stack-id"},
			err:   "--clear-stack-id cannot be combined with --stack-id",
		},
		{
			name:  "clear schedule conflicts",
			in:    JobUpdateInput{ClearSchedule: true},
			flags: []string{"schedule"},
			err:   "--clear-schedule cannot be combined with --schedule or --schedule-timezone",
		},
		{
			name:  "clear schedule timezone conflicts",
			in:    JobUpdateInput{ClearSchedule: true},
			flags: []string{"schedule-timezone"},
			err:   "--clear-schedule cannot be combined with --schedule or --schedule-timezone",
		},
		{
			name:  "clear retention conflicts",
			in:    JobUpdateInput{ClearRetention: true},
			flags: []string{"retention-ttl"},
			err:   "--clear-retention cannot be combined with --retention-* flags",
		},
		{
			name:  "malformed env",
			in:    JobUpdateInput{JobInput: JobInput{EnvVars: []string{"BAD"}}},
			flags: []string{"env"},
			err:   "invalid --env value \"BAD\"; expected NAME=VALUE",
		},
		{
			name: "large supplied files are passed to the server",
			in: JobUpdateInput{
				JobInput: JobInput{Script: strings.Repeat("x", 400_000), Pyproject: "x"},
			},
			flags: []string{"script", "pyproject"},
			want:  map[string]any{"script": strings.Repeat("x", 400_000), "pyproject": "x"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch, err := BuildJobUpdatePatch(
				tt.in,
				func(name string) bool { return slices.Contains(tt.flags, name) },
			)
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != tt.err {
				t.Fatalf("error = %q, want %q", message, tt.err)
			}
			encoded, err := json.Marshal(patch)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("patch mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
