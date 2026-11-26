package platform

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAgentAnalyticsURIs(t *testing.T) {
	const base = "/api/platform/v1/organizations/org-1/projects/proj-1/agents/"
	tests := []struct {
		agent string
		opts  AgentAnalyticsOptions
		want  string
	}{
		{"bot", AgentAnalyticsOptions{}, "bot/%s?"},
		{"bot", AgentAnalyticsOptions{TimeRange: "6h", Step: "5m"}, "bot/%s?step=5m&time_range=6h"},
		{"my bot", AgentAnalyticsOptions{TimeRange: "7d"}, "my%%20bot/%s?time_range=7d"},
	}
	for _, tt := range tests {
		t.Run(tt.agent+" "+tt.opts.TimeRange+" "+tt.opts.Step, func(t *testing.T) {
			got, err := agentAnalyticsURIs("org-1", "proj-1", tt.agent, tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]string{}
			for key, route := range map[string]string{
				"cpu_usage": "cpu-usage", "memory_usage": "memory-usage",
				"status_codes": "status-codes", "requests_per_domain": "requests-per-domain",
				"latency": "latency",
			} {
				want[key] = base + fmt.Sprintf(tt.want, route)
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("URIs (-want +got):\n%s", diff)
			}
		})
	}
}
