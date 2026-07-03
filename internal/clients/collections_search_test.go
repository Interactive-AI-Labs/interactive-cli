package clients

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWithHybridMode(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		want    map[string]any // expected decoded body; nil when wantErr
		wantErr bool
	}{
		{
			name: "injects mode when absent, preserves the body",
			body: `{"queries":[{"query":"x"}],"limit":5}`,
			want: map[string]any{
				"mode":    "hybrid",
				"queries": []any{map[string]any{"query": "x"}},
				"limit":   float64(5),
			},
		},
		{
			name: "forces mode to hybrid over an existing value",
			body: `{"mode":"single","queries":[]}`,
			want: map[string]any{"mode": "hybrid", "queries": []any{}},
		},
		{
			name:    "non-object body is an error, not a silent pass-through",
			body:    `[1,2,3]`,
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := withHybridMode([]byte(c.body))
			if c.wantErr {
				if err == nil {
					t.Fatalf("withHybridMode(%s) expected error", c.body)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(out, &got); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(c.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
