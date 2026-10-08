package sync

import (
	"bytes"
	"io"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/google/go-cmp/cmp"
)

func TestSyncResources(t *testing.T) {
	unchanged := deployment.CreateServiceBody{Replicas: 1}
	changed := deployment.CreateServiceBody{Replicas: 2}
	tests := []struct {
		name              string
		existing, desired map[string]deployment.CreateServiceBody
		options           Options
		want              Result
		out               string
	}{
		{
			name:     "dry-run reports no updates when the server would change nothing",
			existing: map[string]deployment.CreateServiceBody{"api": unchanged},
			desired:  map[string]deployment.CreateServiceBody{"api": unchanged},
			options:  Options{DryRun: true},
			out:      "No changes required; services already match config.\n",
		},
		{
			name: "dry-run reports creates and what the server would change",
			existing: map[string]deployment.CreateServiceBody{
				"keep":   unchanged,
				"change": unchanged,
			},
			desired: map[string]deployment.CreateServiceBody{
				"keep":   unchanged,
				"change": changed,
				"new":    unchanged,
			},
			options: Options{DryRun: true},
			want:    Result{Created: []string{"new"}, Updated: []string{"change"}},
			out:     "Would create services: new\nWould update services: change\n",
		},
		{
			name:     "dry-run deletes omitted resources when allowed",
			existing: map[string]deployment.CreateServiceBody{"keep": unchanged, "old": unchanged},
			desired:  map[string]deployment.CreateServiceBody{"keep": unchanged},
			options:  Options{DryRun: true, AllowDelete: true},
			want:     Result{Deleted: []string{"old"}},
			out:      "Would delete services: old\n",
		},
		{
			name:     "dry-run protects omitted resources",
			existing: map[string]deployment.CreateServiceBody{"keep": unchanged, "old": unchanged},
			desired:  map[string]deployment.CreateServiceBody{"keep": unchanged},
			options:  Options{DryRun: true},
			want:     Result{Protected: []string{"old"}},
			out:      "Would refuse to delete services: old (a config that omits a resource looks identical to a stale one — pass --allow-delete=services to delete)\n",
		},
		{
			name:     "apply leaves unchanged resources out of the result",
			existing: map[string]deployment.CreateServiceBody{"api": unchanged},
			desired:  map[string]deployment.CreateServiceBody{"api": unchanged},
			out:      "No changes required; services already match config.\n",
		},
		{
			name: "apply reports what the server changed",
			existing: map[string]deployment.CreateServiceBody{
				"keep":   unchanged,
				"change": unchanged,
			},
			desired: map[string]deployment.CreateServiceBody{
				"keep":   unchanged,
				"change": changed,
				"new":    unchanged,
			},
			want: Result{Created: []string{"new"}, Updated: []string{"change"}},
			out:  "Created services: new\nUpdated services: change\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := func(name string, body deployment.CreateServiceBody) (bool, error) {
				return body.Replicas != tt.existing[name].Replicas, nil
			}
			got, err := syncResources(io.Discard, tt.existing, tt.desired, tt.options,
				resourceOps[deployment.CreateServiceBody, deployment.CreateServiceBody]{
					resource:  "service",
					allowFlag: "services",
					create:    func(string, deployment.CreateServiceBody) error { return nil },
					update:    changed,
					plan:      changed,
					delete:    func(string) error { return nil },
				})
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, *got); diff != "" {
				t.Fatal(diff)
			}
			var out bytes.Buffer
			if tt.options.DryRun {
				PrintPlan(&out, "services", got)
			} else if err := PrintResult(&out, "services", got, nil); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.out {
				t.Fatalf("output = %q, want %q", out.String(), tt.out)
			}
		})
	}
}
