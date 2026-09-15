package inputs

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
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
		want       deployment.CreateAgentBody
	}{
		{
			name: "minimal valid input",
			yaml: "language: en\n",
			input: AgentInput{
				Id:      "interactive-agent",
				Version: "0.0.1",
			},
			want: deployment.CreateAgentBody{
				Id:          "interactive-agent",
				Version:     "0.0.1",
				AgentConfig: map[string]any{"language": "en"},
			},
		},
		{
			name: "--mcp reaches the config the create sends",
			yaml: "language: en\n",
			input: AgentInput{
				Id:      "interactive-agent",
				Version: "0.0.1",
				McpRefs: []McpRef{{Name: "tools-dev", Id: "tools", SetId: true}},
			},
			want: deployment.CreateAgentBody{
				Id:      "interactive-agent",
				Version: "0.0.1",
				AgentConfig: map[string]any{
					"language": "en",
					"mcps":     []any{map[string]any{"ref": "tools-dev", "id": "tools"}},
				},
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
			want: deployment.CreateAgentBody{
				Id:       "interactive-agent",
				Version:  "1.0.0",
				Endpoint: true,
				AgentConfig: map[string]any{
					"context": map[string]any{"description": "test"},
				},
				Env: []deployment.EnvVar{
					{Name: "KEY1", Value: "value1"},
					{Name: "KEY2", Value: "value2"},
				},
				SecretRefs: []deployment.SecretRef{
					{SecretName: "my-secret"},
					{SecretName: "other-secret"},
				},
				Schedule: &deployment.Schedule{
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
			want: deployment.CreateAgentBody{
				Id:          "interactive-agent",
				Version:     "0.0.1",
				AgentConfig: map[string]any{"language": "en"},
				Schedule: &deployment.Schedule{
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
			want: deployment.CreateAgentBody{
				Id:          "interactive-agent",
				Version:     "0.0.1",
				AgentConfig: map[string]any{"language": "en"},
				Env: []deployment.EnvVar{
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
			want: deployment.CreateAgentBody{
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

func TestMcpRefsFor(t *testing.T) {
	tests := []struct {
		name       string
		names      []string
		id         string
		idGiven    bool
		want       []McpRef
		errContain string
	}{
		{
			name:  "a name alone takes the mcp's own name as its prefix",
			names: []string{"github"},
			want:  []McpRef{{Name: "github"}},
		},
		{
			name:       "--mcp attaches one, so a second is refused rather than dropped",
			names:      []string{"github", "stripe"},
			errContain: "--mcp attaches one mcp (got 2)",
		},
		{
			name:    "--mcp-id is the prefix the agent calls its tools by",
			names:   []string{"tools-dev"},
			id:      "tools",
			idGiven: true,
			want:    []McpRef{{Name: "tools-dev", Id: "tools", SetId: true}},
		},
		{
			name:    "a prefix equal to the name is the default, so only the clear survives",
			names:   []string{"github"},
			id:      "github",
			idGiven: true,
			want:    []McpRef{{Name: "github", SetId: true}},
		},
		{
			name:       "--mcp-id with no --mcp at all",
			id:         "tools",
			idGiven:    true,
			errContain: "pass --mcp <name> too",
		},
		{
			name:       "an empty --mcp-id is a typo, not the default",
			names:      []string{"tools-dev"},
			idGiven:    true,
			errContain: "--mcp-id must name the prefix",
		},
		{
			name:       "an empty --mcp",
			names:      []string{"  "},
			errContain: "--mcp must name an mcp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := McpRefsFor(tt.names, tt.id, tt.idGiven)
			if tt.errContain != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContain) {
					t.Fatalf("error = %v, want to contain %q", err, tt.errContain)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("McpRefsFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInjectMcpRefs(t *testing.T) {
	mslearn := map[string]any{
		"id":        "mslearn",
		"hostname":  "https://learn.microsoft.com",
		"port":      443,
		"transport": "streamable-http",
	}
	tests := []struct {
		name       string
		cfg        map[string]any
		refs       []McpRef
		want       []any
		errContain string
	}{
		{
			name: "an mcp already attached is not attached twice",
			cfg: map[string]any{
				"mcps": []any{
					"github",
					mslearn,
				},
			},
			refs: []McpRef{{Name: "github"}},
			want: []any{"github", mslearn},
		},
		{
			name: "a new one is appended after what is there",
			cfg: map[string]any{
				"mcps": []any{"github", mslearn},
			},
			refs: []McpRef{{Name: "acme"}},
			want: []any{"github", mslearn, "acme"},
		},
		{
			name: "no existing mcps key",
			cfg:  map[string]any{},
			refs: []McpRef{{Name: "github"}},
			want: []any{"github"},
		},
		{
			name: "a prefix attaches as a ref, so the name survives alongside it",
			cfg:  map[string]any{},
			refs: []McpRef{{Name: "tools-dev", Id: "tools", SetId: true}},
			want: []any{map[string]any{"ref": "tools-dev", "id": "tools"}},
		},
		{
			name: "a prefix on an mcp already attached replaces its entry",
			cfg:  map[string]any{"mcps": []any{"tools-dev"}},
			refs: []McpRef{{Name: "tools-dev", Id: "tools", SetId: true}},
			want: []any{map[string]any{"ref": "tools-dev", "id": "tools"}},
		},
		{
			name: "attaching again with no prefix leaves the one already set alone",
			cfg: map[string]any{
				"mcps": []any{map[string]any{"ref": "tools-dev", "id": "tools"}},
			},
			refs: []McpRef{{Name: "tools-dev"}},
			want: []any{map[string]any{"ref": "tools-dev", "id": "tools"}},
		},
		{
			name: "asking for the name as the prefix puts the entry back in its short form",
			cfg: map[string]any{
				"mcps": []any{map[string]any{"ref": "tools-dev", "id": "tools"}},
			},
			refs: []McpRef{{Name: "tools-dev", SetId: true}},
			want: []any{"tools-dev"},
		},
		{
			name: "asking for the name as the prefix leaves a bare entry bare",
			cfg:  map[string]any{"mcps": []any{"tools-dev"}},
			refs: []McpRef{{Name: "tools-dev", SetId: true}},
			want: []any{"tools-dev"},
		},
		{
			name:       "a server configured in full is not turned into a ref",
			cfg:        map[string]any{"mcps": []any{mslearn}},
			refs:       []McpRef{{Name: "mslearn", Id: "docs", SetId: true}},
			errContain: "configured in full in this agent's config",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := InjectMcpRefs(tt.cfg, tt.refs)
			if tt.errContain != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContain) {
					t.Fatalf("error = %v, want to contain %q", err, tt.errContain)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			mcps := out.(map[string]any)["mcps"].([]any)
			if !reflect.DeepEqual(mcps, tt.want) {
				t.Errorf("mcps = %v, want %v", mcps, tt.want)
			}
		})
	}
}

func TestDetachMcpRefs(t *testing.T) {
	mslearn := map[string]any{
		"id":        "mslearn",
		"hostname":  "https://learn.microsoft.com",
		"port":      443,
		"transport": "streamable-http",
	}
	tests := []struct {
		name  string
		cfg   map[string]any
		names []string
		want  []any
	}{
		{
			name: "removes matching bare and resolved refs, ignores unmatched names",
			cfg: map[string]any{
				"mcps": []any{
					"github",
					mslearn,
					"stripe",
				},
			},
			names: []string{"mslearn", "not-attached"},
			want: []any{
				"github",
				"stripe",
			},
		},
		{
			name: "detaches by the mcp's name, not by the prefix it was given",
			cfg: map[string]any{
				"mcps": []any{
					map[string]any{"ref": "tools-dev", "id": "tools"},
					"github",
				},
			},
			names: []string{"tools-dev"},
			want:  []any{"github"},
		},
		{
			name: "the prefix is not the mcp's name, so it does not detach it",
			cfg: map[string]any{
				"mcps": []any{map[string]any{"ref": "tools-dev", "id": "tools"}},
			},
			names: []string{"tools"},
			want:  []any{map[string]any{"ref": "tools-dev", "id": "tools"}},
		},
		{
			name:  "no names is a no-op",
			cfg:   map[string]any{"mcps": []any{"github"}},
			names: nil,
			want:  []any{"github"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := DetachMcpRefs(tt.cfg, tt.names)
			if err != nil {
				t.Fatal(err)
			}
			mcps := out.(map[string]any)["mcps"].([]any)
			if !reflect.DeepEqual(mcps, tt.want) {
				t.Errorf("mcps = %v, want %v", mcps, tt.want)
			}
		})
	}
}
