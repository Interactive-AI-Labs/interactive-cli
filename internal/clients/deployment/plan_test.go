package deployment

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDryRunPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "service",
			path: "/v1/organizations/o/projects/p/services/api",
			want: "/v1/organizations/o/projects/p/services/api?dryRun=true",
		},
		{
			name: "agent",
			path: "/v1/organizations/o/projects/p/agents/helper",
			want: "/v1/organizations/o/projects/p/agents/helper?dryRun=true",
		},
		{
			name: "database",
			path: "/v1/organizations/o/projects/p/databases/store",
			want: "/v1/organizations/o/projects/p/databases/store?dryRun=true",
		},
		{
			name: "mcp",
			path: "/v1/organizations/o/projects/p/mcps/tools",
			want: "/v1/organizations/o/projects/p/mcps/tools?dryRun=true",
		},
		{
			name: "job",
			path: "/v1/organizations/o/projects/p/jobs/report",
			want: "/v1/organizations/o/projects/p/jobs/report?dryRun=true",
		},
		{
			name: "escaped query characters stay in the resource name",
			path: "/v1/organizations/o/projects/p/services/api%3FdryRun=false",
			want: "/v1/organizations/o/projects/p/services/api%3FdryRun=false?dryRun=true",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dryRunPath(tt.path); got != tt.want {
				t.Fatalf("dryRunPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecodeChanged(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{name: "changed", body: `{"changed":true}`, want: true},
		{name: "unchanged", body: `{"changed":false}`, want: false},
		{name: "flag absent", body: `{"message":"ok"}`, want: true},
		{name: "empty object", body: `{}`, want: true},
		{name: "empty body", body: ``, want: true},
		{name: "not json", body: `replaced`, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeChanged([]byte(tt.body)); got != tt.want {
				t.Fatalf("decodeChanged(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}

func TestDecodePlan(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    *Plan[CreateJobBody]
		wantErr string
	}{
		{
			name: "changed with configuration",
			body: `{"changed":true,"type":"script","script":"print(1)"}`,
			want: &Plan[CreateJobBody]{
				Changed: true,
				Config:  CreateJobBody{Type: "script", Script: "print(1)"},
			},
		},
		{
			name: "unchanged",
			body: `{"changed":false,"type":"image"}`,
			want: &Plan[CreateJobBody]{Config: CreateJobBody{Type: "image"}},
		},
		{
			name:    "not json",
			body:    `replaced`,
			wantErr: "failed to decode dry-run response: invalid character 'r' looking for beginning of value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodePlan[CreateJobBody]([]byte(tt.body))
			if (err != nil && err.Error() != tt.wantErr) || (err == nil && tt.wantErr != "") {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("plan mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
