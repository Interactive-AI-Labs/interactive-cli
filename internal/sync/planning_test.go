package sync

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSelectUpdate(t *testing.T) {
	tests := []struct {
		name          string
		dryRun        bool
		wantOperation string
		wantContext   string
	}{
		{name: "apply", wantOperation: "apply operation", wantContext: "update"},
		{
			name:          "dry run",
			dryRun:        true,
			wantOperation: "plan operation",
			wantContext:   "plan update for",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation, context := selectUpdate(tt.dryRun, "apply operation", "plan operation")
			if operation != tt.wantOperation || context != tt.wantContext {
				t.Fatalf(
					"operation/context = (%q, %q), want (%q, %q)",
					operation,
					context,
					tt.wantOperation,
					tt.wantContext,
				)
			}
		})
	}
}

func TestDryRunWithoutOperations(t *testing.T) {
	tests := []struct {
		name              string
		existing, desired map[string]string
		deferred          map[string][]string
		allowDelete       bool
		want              Result
		wantWarning       string
	}{
		{name: "empty stack"},
		{
			name:    "creates are sorted without calling create",
			desired: map[string]string{"zulu": "new", "alpha": "new"},
			want:    Result{Created: []string{"alpha", "zulu"}},
		},
		{
			name:     "omitted resources are protected",
			existing: map[string]string{"zulu": "live", "alpha": "live"},
			want:     Result{Protected: []string{"alpha", "zulu"}},
		},
		{
			name:        "allowed deletions never call delete",
			existing:    map[string]string{"zulu": "live", "alpha": "live"},
			allowDelete: true,
			want:        Result{Deleted: []string{"alpha", "zulu"}},
		},
		{
			name:     "deferred update calls neither plan nor apply",
			existing: map[string]string{"helper": "live"},
			desired:  map[string]string{"helper": "desired"},
			deferred: map[string][]string{"helper": {"tools"}},
			want:     Result{Deferred: map[string][]string{"helper": {"tools"}}},
		},
		{
			name:        "deferred update still allows create and delete predictions",
			existing:    map[string]string{"helper": "live", "old": "live"},
			desired:     map[string]string{"helper": "desired", "new": "desired"},
			deferred:    map[string][]string{"helper": {"tools"}},
			allowDelete: true,
			want: Result{
				Created:  []string{"new"},
				Deleted:  []string{"old"},
				Deferred: map[string][]string{"helper": {"tools"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var warnings bytes.Buffer
			// These dry-run paths must not invoke any operations; none are supplied.
			got, err := syncResources(
				&warnings,
				tt.existing,
				tt.desired,
				Options{DryRun: true, AllowDelete: tt.allowDelete},
				resourceOps[string, string]{
					resource:  "agent",
					allowFlag: "agents",
					deferred:  tt.deferred,
				},
			)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, *got); diff != "" {
				t.Fatalf("result mismatch (-want +got):\n%s", diff)
			}
			if warnings.String() != tt.wantWarning {
				t.Fatalf("warnings = %q, want %q", warnings.String(), tt.wantWarning)
			}
		})
	}
}

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
