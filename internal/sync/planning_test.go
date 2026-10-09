package sync

import (
	"bytes"
	"testing"
)

func TestPrintDeferredPlans(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   string
	}{
		{
			name: "no changes",
			want: "No changes required; agents already match config.\n",
		},
		{
			name:   "deferred is not unchanged",
			result: Result{Deferred: map[string][]string{"helper": {"tools", "web"}}},
			want:   "Plan deferred for agents helper: changing mcps: tools, web; check again after those changes.\n",
		},
		{
			name: "deferred agents are sorted alongside known operations",
			result: Result{
				Created: []string{"new"}, Updated: []string{"changed"}, Deleted: []string{"old"},
				Deferred: map[string][]string{"zulu": {"web"}, "alpha": {"tools"}},
			},
			want: "Would create agents: new\nWould update agents: changed\nWould delete agents: old\n" +
				"Plan deferred for agents alpha: changing mcps: tools; check again after those changes.\n" +
				"Plan deferred for agents zulu: changing mcps: web; check again after those changes.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			PrintPlan(&out, "agents", &tt.result)
			if got := out.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
