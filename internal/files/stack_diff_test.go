package files

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/google/go-cmp/cmp"
)

func TestDiffFields(t *testing.T) {
	tests := []struct {
		name  string
		live  any
		local any
		want  []fieldChange
	}{
		{
			name:  "no changes",
			live:  map[string]any{"a": "1", "b": "2"},
			local: map[string]any{"a": "1", "b": "2"},
			want:  nil,
		},
		{
			name:  "value changed",
			live:  map[string]any{"a": "1"},
			local: map[string]any{"a": "2"},
			want:  []fieldChange{{path: "a", old: "1", new: "2"}},
		},
		{
			name:  "field added",
			live:  map[string]any{},
			local: map[string]any{"a": "1"},
			want:  []fieldChange{{path: "a", new: "1"}},
		},
		{
			name:  "field removed",
			live:  map[string]any{"a": "1"},
			local: map[string]any{},
			want:  []fieldChange{{path: "a", old: "1"}},
		},
		{
			name:  "list cleared",
			live:  map[string]any{"tools": []any{"search"}},
			local: map[string]any{"tools": []any{}},
			want:  []fieldChange{{path: "tools", new: "[]"}, {path: "tools[0]", old: "search"}},
		},
		{
			name:  "map cleared",
			live:  map[string]any{"settings": map[string]any{"mode": "auto"}},
			local: map[string]any{"settings": map[string]any{}},
			want: []fieldChange{
				{path: "settings", new: "{}"},
				{path: "settings.mode", old: "auto"},
			},
		},
		{
			name:  "empty collections removed",
			live:  map[string]any{"settings": map[string]any{}, "tools": []any{}},
			local: map[string]any{},
			want:  []fieldChange{{path: "settings", old: "{}"}, {path: "tools", old: "[]"}},
		},
		{
			name:  "mixed changes",
			live:  map[string]any{"a": "1", "b": "2", "c": "3"},
			local: map[string]any{"a": "1", "b": "X", "d": "4"},
			want: []fieldChange{
				{path: "b", old: "2", new: "X"},
				{path: "c", old: "3"},
				{path: "d", new: "4"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffFields(tt.live, tt.local)
			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(fieldChange{})); diff != "" {
				t.Errorf("diffFields() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestToFlatMap(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  map[string]string
	}{
		{
			name:  "flat struct",
			input: struct{ A, B string }{"1", "2"},
			want:  map[string]string{"A": "1", "B": "2"},
		},
		{
			name: "nested map",
			input: map[string]any{
				"image": map[string]any{"tag": "v1"},
				"port":  float64(8080),
			},
			want: map[string]string{
				"image.tag": "v1",
				"port":      "8080",
			},
		},
		{
			name: "array of objects",
			input: map[string]any{
				"env": []any{
					map[string]any{"name": "K", "value": "V"},
				},
			},
			want: map[string]string{
				"env[0].name":  "K",
				"env[0].value": "V",
			},
		},
		{
			name:  "empty collections",
			input: map[string]any{"settings": map[string]any{}, "tools": []any{}},
			want:  map[string]string{"settings": "{}", "tools": "[]"},
		},
		{
			name:  "empty root map",
			input: map[string]any{},
			want:  map[string]string{},
		},
		{
			name:  "nested empty collections",
			input: map[string]any{"items": []any{map[string]any{}, []any{}}},
			want:  map[string]string{"items[0]": "{}", "items[1]": "[]"},
		},
		{
			name:  "bool and null",
			input: map[string]any{"active": true, "extra": nil},
			want:  map[string]string{"active": "true", "extra": "null"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toFlatMap(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("toFlatMap() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDiffStackConfigs(t *testing.T) {
	svc := ServiceConfig{
		ServicePort: 8080,
		Image:       clientsImageSpec("nginx", "latest"),
		Resources:   clientsResources("256M", "0.25"),
		Replicas:    1,
	}
	scaled := svc
	scaled.Replicas = 2
	tools := McpConfig{Type: "self-hosted", Port: 8080}
	public := tools
	public.Endpoint = true

	tests := []struct {
		name        string
		plan        *StackPlan
		live        *StackConfig
		want        *StackDiff
		wantChanges bool
		wantPrinted string
	}{
		{
			name:        "no resources",
			plan:        &StackPlan{StackId: "t"},
			live:        &StackConfig{StackId: "t"},
			want:        &StackDiff{StackID: "t"},
			wantPrinted: "No differences found.\n",
		},
		{
			name: "unchanged resources are left out",
			plan: &StackPlan{StackId: "t", Services: ResourcePlan[ServiceConfig]{
				Desired: map[string]ServiceConfig{"api": svc},
			}},
			live:        &StackConfig{StackId: "t", Services: map[string]ServiceConfig{"api": svc}},
			want:        &StackDiff{StackID: "t"},
			wantPrinted: "No differences found.\n",
		},
		{
			name:        "create, update, and delete",
			wantChanges: true,
			plan: &StackPlan{StackId: "t", Services: ResourcePlan[ServiceConfig]{
				Desired: map[string]ServiceConfig{"keep": svc, "update": scaled, "create": svc},
				Changed: map[string]bool{"update": true},
			}},
			live: &StackConfig{StackId: "t", Services: map[string]ServiceConfig{
				"keep": svc, "update": svc, "delete": svc,
			}},
			want: &StackDiff{StackID: "t", Services: ResourceTypeDiff{
				Created: []string{"create"},
				Updated: []ResourceChange{{
					Name:    "update",
					Changes: map[string]FieldDiff{"replicas": {Old: "1", New: "2"}},
				}},
				Deleted: []string{"delete"},
			}},
			wantPrinted: "Stack: t\n\n  + service create (create)\n\n  ~ service update (update)\n" +
				"    replicas: 1 → 2\n\n  - service delete (delete)\n",
		},
		{
			name:        "server-reported change without field differences",
			wantChanges: true,
			plan: &StackPlan{StackId: "t", Services: ResourcePlan[ServiceConfig]{
				Desired: map[string]ServiceConfig{"api": svc},
				Changed: map[string]bool{"api": true},
			}},
			live: &StackConfig{StackId: "t", Services: map[string]ServiceConfig{"api": svc}},
			want: &StackDiff{StackID: "t", Services: ResourceTypeDiff{
				Updated: []ResourceChange{{Name: "api", Changes: map[string]FieldDiff{}}},
			}},
			wantPrinted: "Stack: t\n\n  ~ service api (update)\n",
		},
		{
			name:        "mcp endpoint removed",
			wantChanges: true,
			plan: &StackPlan{StackId: "s", Mcps: ResourcePlan[McpConfig]{
				Desired: map[string]McpConfig{"tools": tools},
				Changed: map[string]bool{"tools": true},
			}},
			live: &StackConfig{StackId: "s", Mcps: map[string]McpConfig{"tools": public}},
			want: &StackDiff{StackID: "s", Mcps: ResourceTypeDiff{
				Updated: []ResourceChange{{
					Name:    "tools",
					Changes: map[string]FieldDiff{"endpoint": {Old: "true"}},
				}},
			}},
			wantPrinted: "Stack: s\n\n  ~ mcp tools (update)\n    - endpoint\n",
		},
		{
			name: "agent list cleared",
			plan: &StackPlan{StackId: "s", Agents: ResourcePlan[AgentConfig]{
				Desired: map[string]AgentConfig{
					"helper": {AgentConfig: map[string]any{"tools": []any{}}},
				},
				Changed: map[string]bool{"helper": true},
			}},
			live: &StackConfig{
				Agents: map[string]AgentConfig{
					"helper": {AgentConfig: map[string]any{"tools": []any{"search"}}},
				},
			},
			want: &StackDiff{StackID: "s", Agents: ResourceTypeDiff{Updated: []ResourceChange{{
				Name: "helper", Changes: map[string]FieldDiff{
					"agentConfig.tools": {New: "[]"}, "agentConfig.tools[0]": {Old: "search"},
				},
			}}}},
			wantChanges: true,
			wantPrinted: "Stack: s\n\n  ~ agent helper (update)\n    + agentConfig.tools: []\n    - agentConfig.tools[0]\n",
		},
		{
			name: "deferred plans are visible without predicted updates",
			plan: &StackPlan{StackId: "s", Agents: ResourcePlan[AgentConfig]{
				Desired:  map[string]AgentConfig{"helper": {}},
				Deferred: map[string][]string{"helper": {"tools", "web"}},
			}},
			live: &StackConfig{Agents: map[string]AgentConfig{"helper": {}}},
			want: &StackDiff{
				StackID: "s",
				Agents: ResourceTypeDiff{
					Deferred: map[string][]string{"helper": {"tools", "web"}},
				},
			},
			wantChanges: true,
			wantPrinted: "Stack: s\n\n  ? agent helper (plan deferred; changing mcps: tools, web)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DiffStackConfigs(tt.plan, tt.live)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("diff mismatch (-want +got):\n%s", diff)
			}
			if got.HasChanges() != tt.wantChanges {
				t.Fatalf("HasChanges() = %v", got.HasChanges())
			}
			var out bytes.Buffer
			if err := PrintStackDiffDetailed(&out, tt.plan, tt.live, got); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.wantPrinted {
				t.Fatalf("printed = %q, want %q", out.String(), tt.wantPrinted)
			}
		})
	}
}

func clientsImageSpec(name, tag string) deployment.ImageSpec {
	return deployment.ImageSpec{Type: "external", Repository: "docker.io", Name: name, Tag: tag}
}

func clientsResources(mem, cpu string) deployment.Resources {
	return deployment.Resources{Memory: mem, CPU: cpu}
}
