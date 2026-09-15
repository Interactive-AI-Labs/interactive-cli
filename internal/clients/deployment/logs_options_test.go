package deployment

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLogsQuery(t *testing.T) {
	for _, tt := range []struct {
		name string
		opts LogsOptions
		want url.Values
	}{
		{"defaults stay on server", LogsOptions{}, url.Values{}},
		{"follow", LogsOptions{Follow: true, Since: "30m", Limit: 10}, url.Values{"follow": {"true"}, "since": {"30m"}, "limit": {"10"}}},
		{"absolute window and filters", LogsOptions{StartTime: "2026-01-01T00:00:00Z", EndTime: "2026-01-01T01:00:00Z", Message: "failed|timeout & retry", Level: "error"}, url.Values{"start-time": {"2026-01-01T00:00:00Z"}, "end-time": {"2026-01-01T01:00:00Z"}, "message": {"failed|timeout & retry"}, "level": {"error"}}},
		{"invalid limit reaches validation", LogsOptions{Limit: -1}, url.Values{"limit": {"-1"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, tt.opts.query()); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
