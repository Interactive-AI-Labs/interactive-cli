package deployment

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReplicaListPath(t *testing.T) {
	for _, tt := range []struct {
		name, resource, want string
	}{
		{"plain", "my-tool", "/v1/organizations/org%2F1/projects/project%202/replicas?resource=my-tool"},
		{"escaped", "a b&c", "/v1/organizations/org%2F1/projects/project%202/replicas?resource=a+b%26c"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := replicaListPath("org/1", "project 2", tt.resource)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
