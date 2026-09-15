package deployment

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReplicaWorkloadPath(t *testing.T) {
	for _, tt := range []struct {
		kind, want, wantErr string
	}{
		{"service", "/v1/organizations/org%2F1/projects/project%202/services", ""},
		{"mcp", "/v1/organizations/org%2F1/projects/project%202/mcps", ""},
		{"agent", "", `invalid replica type "agent": expected service or mcp`},
		{"", "", `invalid replica type "": expected service or mcp`},
	} {
		t.Run(tt.kind, func(t *testing.T) {
			got, err := replicaWorkloadPath("org/1", "project 2", tt.kind)
			message := ""
			if err != nil {
				message = err.Error()
			}
			if diff := cmp.Diff([]string{tt.want, tt.wantErr}, []string{got, message}); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
