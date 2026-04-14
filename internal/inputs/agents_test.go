package inputs

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "agent-config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestBuildAgentRequestBody(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		input      AgentInput
		wantErr    bool
		errContain string
		want       clients.CreateAgentBody
	}{
		{
			name: "minimal valid input",
			yaml: "language: en\n",
			input: AgentInput{
				Id:      "interactive-agent",
				Version: "0.0.1",
			},
			want: clients.CreateAgentBody{
				Id:          "interactive-agent",
				Version:     "0.0.1",
				AgentConfig: map[string]any{"language": "en"},
			},
		},
		{
			name: "all fields populated",
			yaml: "context:\n  description: test\n",
			input: AgentInput{
				Id:               "interactive-agent",
				Version:          "1.0.0",
				Endpoint:         true,
				EnvVars:          []string{"KEY1=value1", "KEY2=value2"},
				SecretRefs:       []string{"my-secret", "other-secret"},
				ScheduleUptime:   "Mon-Fri 07:30-20:30",
				ScheduleTimezone: "Europe/Berlin",
			},
			want: clients.CreateAgentBody{
				Id:       "interactive-agent",
				Version:  "1.0.0",
				Endpoint: true,
				AgentConfig: map[string]any{
					"context": map[string]any{"description": "test"},
				},
				Env: []clients.EnvVar{
					{Name: "KEY1", Value: "value1"},
					{Name: "KEY2", Value: "value2"},
				},
				SecretRefs: []clients.SecretRef{
					{SecretName: "my-secret"},
					{SecretName: "other-secret"},
				},
				Schedule: &clients.Schedule{
					Uptime:   "Mon-Fri 07:30-20:30",
					Timezone: "Europe/Berlin",
				},
			},
		},
		{
			name: "schedule with downtime only",
			yaml: "language: en\n",
			input: AgentInput{
				Id:               "interactive-agent",
				Version:          "0.0.1",
				ScheduleDowntime: "Sat-Sun 00:00-24:00",
				ScheduleTimezone: "UTC",
			},
			want: clients.CreateAgentBody{
				Id:          "interactive-agent",
				Version:     "0.0.1",
				AgentConfig: map[string]any{"language": "en"},
				Schedule: &clients.Schedule{
					Downtime: "Sat-Sun 00:00-24:00",
					Timezone: "UTC",
				},
			},
		},
		{
			name: "env var with equals in value",
			yaml: "language: en\n",
			input: AgentInput{
				Id:      "interactive-agent",
				Version: "0.0.1",
				EnvVars: []string{"CONN=postgres://host:5432/db?opt=val"},
			},
			want: clients.CreateAgentBody{
				Id:          "interactive-agent",
				Version:     "0.0.1",
				AgentConfig: map[string]any{"language": "en"},
				Env: []clients.EnvVar{
					{Name: "CONN", Value: "postgres://host:5432/db?opt=val"},
				},
			},
		},
		{
			name: "parsed YAML structure",
			yaml: "context:\n  description:\n    prompt_id: my-prompt\n    version: 1\n",
			input: AgentInput{
				Id:      "interactive-agent",
				Version: "0.0.1",
			},
			want: clients.CreateAgentBody{
				Id:      "interactive-agent",
				Version: "0.0.1",
				AgentConfig: map[string]any{
					"context": map[string]any{
						"description": map[string]any{
							"prompt_id": "my-prompt",
							"version":   1,
						},
					},
				},
			},
		},
		{
			name: "invalid env var",
			yaml: "language: en\n",
			input: AgentInput{
				Id:      "interactive-agent",
				Version: "0.0.1",
				EnvVars: []string{"INVALID"},
			},
			wantErr:    true,
			errContain: "invalid --env value",
		},
		{
			name: "invalid secret ref",
			yaml: "language: en\n",
			input: AgentInput{
				Id:         "interactive-agent",
				Version:    "0.0.1",
				SecretRefs: []string{"valid", "  "},
			},
			wantErr:    true,
			errContain: "invalid --secret value",
		},
		{
			name: "file not found",
			input: AgentInput{
				Id:       "interactive-agent",
				Version:  "0.0.1",
				FilePath: "/nonexistent/path/agent-config.yaml",
			},
			wantErr:    true,
			errContain: "failed to read file",
		},
		{
			name: "invalid YAML",
			yaml: ":\n  :\n    - ][",
			input: AgentInput{
				Id:      "interactive-agent",
				Version: "0.0.1",
			},
			wantErr:    true,
			errContain: "failed to parse YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input.FilePath == "" && tt.yaml != "" {
				tt.input.FilePath = writeTempYAML(t, tt.yaml)
			}

			got, err := BuildAgentRequestBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildAgentRequestBody() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContain)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildAgentRequestBody() =\n  %+v\nwant:\n  %+v", got, tt.want)
			}
		})
	}
}
