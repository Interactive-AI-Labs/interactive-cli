package output

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStripRevisionMeta(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		wantKeys []string
		dropKeys []string
	}{
		{
			name: "strips revision meta fields",
			input: map[string]any{
				"revision": 5,
				"updated":  "2026-01-01T00:00:00Z",
				"status":   "deployed",
				"id":       "my-agent",
				"version":  "1.0.0",
			},
			wantKeys: []string{"id", "version"},
			dropKeys: []string{"revision", "updated", "status"},
		},
		{
			name: "preserves all fields when no meta present",
			input: map[string]any{
				"id":      "my-agent",
				"version": "1.0.0",
			},
			wantKeys: []string{"id", "version"},
			dropKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stripRevisionMeta(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, k := range tt.wantKeys {
				if _, ok := got[k]; !ok {
					t.Errorf("expected key %q to be present", k)
				}
			}
			for _, k := range tt.dropKeys {
				if _, ok := got[k]; ok {
					t.Errorf("expected key %q to be stripped", k)
				}
			}
		})
	}
}

func TestPrintRevisionDiff(t *testing.T) {
	tests := []struct {
		name string
		a    any
		b    any
		want string
	}{
		{
			name: "identical revisions",
			a:    map[string]any{"id": "agent", "version": "1.0"},
			b:    map[string]any{"id": "agent", "version": "1.0"},
			want: "No differences found.\n",
		},
		{
			name: "single field changed",
			a:    map[string]any{"id": "agent", "version": "1.0"},
			b:    map[string]any{"id": "agent", "version": "2.0"},
			want: "--- revision 1\n+++ revision 2\n" +
				"@@ -1,2 +1,2 @@\n" +
				" id: agent\n" +
				"-version: \"1.0\"\n" +
				"+version: \"2.0\"\n",
		},
		{
			name: "revision meta excluded from diff",
			a: map[string]any{
				"revision": 1,
				"status":   "superseded",
				"updated":  "2026-01-01",
				"id":       "agent",
				"version":  "1.0",
			},
			b: map[string]any{
				"revision": 2,
				"status":   "deployed",
				"updated":  "2026-01-02",
				"id":       "agent",
				"version":  "1.0",
			},
			want: "No differences found.\n",
		},
		{
			name: "multiple fields changed",
			a:    map[string]any{"id": "agent", "version": "1.0", "env": []any{"FOO=bar"}},
			b:    map[string]any{"id": "agent", "version": "2.0", "env": []any{"FOO=baz"}},
			want: "--- revision 1\n+++ revision 2\n" +
				"@@ -1,4 +1,4 @@\n" +
				" env:\n" +
				"-    - FOO=bar\n" +
				"+    - FOO=baz\n" +
				" id: agent\n" +
				"-version: \"1.0\"\n" +
				"+version: \"2.0\"\n",
		},
		{
			name: "field added",
			a:    map[string]any{"id": "agent"},
			b:    map[string]any{"id": "agent", "endpoint": "example.com"},
			want: "--- revision 1\n+++ revision 2\n" +
				"@@ -1,1 +1,2 @@\n" +
				"+endpoint: example.com\n" +
				" id: agent\n",
		},
		{
			name: "field removed",
			a:    map[string]any{"id": "agent", "endpoint": "example.com"},
			b:    map[string]any{"id": "agent"},
			want: "--- revision 1\n+++ revision 2\n" +
				"@@ -1,2 +1,1 @@\n" +
				"-endpoint: example.com\n" +
				" id: agent\n",
		},
		{
			name: "nested field changed",
			a:    map[string]any{"config": map[string]any{"language": "en"}},
			b:    map[string]any{"config": map[string]any{"language": "fr"}},
			want: "--- revision 1\n+++ revision 2\n" +
				"@@ -1,2 +1,2 @@\n" +
				" config:\n" +
				"-    language: en\n" +
				"+    language: fr\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintRevisionDiff(&buf, "1", tt.a, "2", tt.b)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.want, buf.String()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
