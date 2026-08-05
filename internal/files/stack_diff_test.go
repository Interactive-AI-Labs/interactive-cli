package files

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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
	svc1 := ServiceConfig{
		ServicePort: 8080,
		Image:       clientsImageSpec("nginx", "latest"),
		Resources:   clientsResources("256M", "0.25"),
		Replicas:    1,
	}
	svc1Mod := svc1
	svc1Mod.Replicas = 2

	live := &StackConfig{
		StackId: "test-stack",
		Services: map[string]ServiceConfig{
			"keep":   svc1,
			"update": svc1,
			"delete": svc1,
		},
		Agents:    map[string]AgentConfig{},
		Databases: map[string]DatabaseConfig{},
	}
	local := &StackConfig{
		StackId: "test-stack",
		Services: map[string]ServiceConfig{
			"keep":   svc1,
			"update": svc1Mod,
			"create": svc1,
		},
		Agents:    map[string]AgentConfig{},
		Databases: map[string]DatabaseConfig{},
	}

	d := DiffStackConfigs(local, live)

	if d.StackID != "test-stack" {
		t.Errorf("StackID = %q, want %q", d.StackID, "test-stack")
	}

	if diff := cmp.Diff([]string{"create"}, d.Services.Created); diff != "" {
		t.Errorf("Created mismatch (-want +got):\n%s", diff)
	}
	if len(d.Services.Updated) != 1 || d.Services.Updated[0].Name != "update" {
		t.Errorf("Updated = %+v, want [update]", d.Services.Updated)
	}
	if diff := cmp.Diff([]string{"delete"}, d.Services.Deleted); diff != "" {
		t.Errorf("Deleted mismatch (-want +got):\n%s", diff)
	}
	if !d.HasChanges() {
		t.Error("HasChanges() = false, want true")
	}
}

func TestDiffStackConfigsNoChanges(t *testing.T) {
	live := &StackConfig{
		StackId:   "t",
		Services:  map[string]ServiceConfig{},
		Agents:    map[string]AgentConfig{},
		Databases: map[string]DatabaseConfig{},
	}
	local := &StackConfig{
		StackId:   "t",
		Services:  map[string]ServiceConfig{},
		Agents:    map[string]AgentConfig{},
		Databases: map[string]DatabaseConfig{},
	}
	d := DiffStackConfigs(local, live)
	if d.HasChanges() {
		t.Error("HasChanges() = true, want false")
	}
}

func TestPrintStackDiffDetailed(t *testing.T) {
	live := &StackConfig{
		StackId: "t",
		Services: map[string]ServiceConfig{
			"svc": {ServicePort: 8080, Replicas: 1},
		},
		Agents:    map[string]AgentConfig{},
		Databases: map[string]DatabaseConfig{},
	}
	local := &StackConfig{
		StackId: "t",
		Services: map[string]ServiceConfig{
			"svc": {ServicePort: 8080, Replicas: 2},
		},
		Agents:    map[string]AgentConfig{},
		Databases: map[string]DatabaseConfig{},
	}

	d := DiffStackConfigs(local, live)
	var buf bytes.Buffer
	if err := PrintStackDiffDetailed(&buf, local, live, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !bytes.Contains([]byte(got), []byte("svc")) {
		t.Errorf("output missing 'svc':\n%s", got)
	}
	if !bytes.Contains([]byte(got), []byte("1 → 2")) {
		t.Errorf("output missing '1 → 2':\n%s", got)
	}
}

func TestPrintStackDiffDetailedNoChanges(t *testing.T) {
	cfg := &StackConfig{
		StackId:   "t",
		Services:  map[string]ServiceConfig{},
		Agents:    map[string]AgentConfig{},
		Databases: map[string]DatabaseConfig{},
	}
	d := DiffStackConfigs(cfg, cfg)
	var buf bytes.Buffer
	if err := PrintStackDiffDetailed(&buf, cfg, cfg, d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); got != "No differences found.\n" {
		t.Errorf("got %q, want %q", got, "No differences found.\n")
	}
}

func clientsImageSpec(name, tag string) clients.ImageSpec {
	return clients.ImageSpec{Type: "external", Repository: "docker.io", Name: name, Tag: tag}
}

func clientsResources(mem, cpu string) clients.Resources {
	return clients.Resources{Memory: mem, CPU: cpu}
}
