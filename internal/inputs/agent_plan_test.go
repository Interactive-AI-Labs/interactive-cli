package inputs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPendingMcpRefs(t *testing.T) {
	tests := []struct {
		name     string
		config   any
		changing []string
		want     []string
	}{
		{name: "no configuration"},
		{name: "no MCPs", config: map[string]any{"model": "test"}, changing: []string{"tools"}},
		{name: "empty MCPs", config: map[string]any{"mcps": []any{}}, changing: []string{"tools"}},
		{
			name:     "unchanged reference",
			config:   map[string]any{"mcps": []any{"tools"}},
			changing: []string{"other"},
		},
		{name: "no changing MCPs", config: map[string]any{"mcps": []any{"tools"}}},
		{
			name:     "new MCP by name",
			config:   map[string]any{"mcps": []any{"tools"}},
			changing: []string{"tools"},
			want:     []string{"tools"},
		},
		{
			name:     "trimmed reference",
			config:   map[string]any{"mcps": []any{" tools "}},
			changing: []string{"tools"},
			want:     []string{"tools"},
		},
		{
			name: "reference uses MCP name not tool prefix",
			config: map[string]any{
				"mcps": []any{
					map[string]any{"ref": " tools ", "id": "prefix", "endpoint": "public"},
				},
			},
			changing: []string{"tools"},
			want:     []string{"tools"},
		},
		{
			name:     "id alone is not an MCP reference",
			config:   map[string]any{"mcps": []any{map[string]any{"id": "tools"}}},
			changing: []string{"tools"},
		},
		{
			name: "ref takes precedence over extra inline fields",
			config: map[string]any{"mcps": []any{map[string]any{
				"ref": "tools", "id": "prefix", "hostname": "https://example.com",
			}}},
			changing: []string{"tools"},
			want:     []string{"tools"},
		},
		{
			name:     "empty ref does not fall back to id",
			config:   map[string]any{"mcps": []any{map[string]any{"ref": "", "id": "tools"}}},
			changing: []string{"tools"},
		},
		{
			name: "inline server is not a reference",
			config: map[string]any{
				"mcps": []any{map[string]any{"id": "tools", "hostname": "https://example.com"}},
			},
			changing: []string{"tools"},
		},
		{
			name: "references are sorted and deduplicated",
			config: map[string]any{
				"mcps": []any{"web", "tools", map[string]any{"ref": "web"}, "unchanged"},
			},
			changing: []string{"tools", "web"},
			want:     []string{"tools", "web"},
		},
		{
			name:     "invalid entries remain for server validation",
			config:   map[string]any{"mcps": []any{nil, 3, " ", map[string]any{"ref": 3}}},
			changing: []string{"tools", ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, PendingMcpRefs(tt.config, tt.changing)); diff != "" {
				t.Fatalf("pending MCPs (-want +got):\n%s", diff)
			}
		})
	}
}
