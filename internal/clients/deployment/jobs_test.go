package deployment

import (
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/google/go-cmp/cmp"
)

func TestJobsPath(t *testing.T) {
	tests := []struct {
		name                      string
		orgId, projectId, jobName string
		want                      string
	}{
		{"list", "org", "project", "", "/v1/organizations/org/projects/project/jobs"},
		{"job", "org", "project", "report", "/v1/organizations/org/projects/project/jobs/report"},
		{
			"escape path segments",
			"org/a",
			"project?x",
			"job#x",
			"/v1/organizations/org%2Fa/projects/project%3Fx/jobs/job%23x",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jobsPath(tt.orgId, tt.projectId, tt.jobName); got != tt.want {
				t.Fatalf("path = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJobRunPath(t *testing.T) {
	tests := []struct {
		name                    string
		orgId, projectId, runId string
		want                    string
	}{
		{
			"specific run",
			"org",
			"project",
			"run-1",
			"/v1/organizations/org/projects/project/job-runs/run-1",
		},
		{
			"escape path segments",
			"org/a",
			"project?x",
			"run/1",
			"/v1/organizations/org%2Fa/projects/project%3Fx/job-runs/run%2F1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jobRunPath(tt.orgId, tt.projectId, tt.runId); got != tt.want {
				t.Fatalf("path = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreateJobBodyJSON(t *testing.T) {
	tests := []struct {
		name string
		body CreateJobBody
		want map[string]any
	}{
		{
			name: "zero values are not omitted",
			body: CreateJobBody{
				Type:      "image",
				Image:     &ImageSpec{Type: "internal", Name: "app", Tag: "v1"},
				Resources: Resources{CPU: "1", Memory: "1G"},
				Retries:   utils.ToPtr(int32(0)),
				Retention: &JobRetention{
					TTL:            utils.ToPtr(int32(0)),
					SuccessfulRuns: utils.ToPtr(int32(0)),
					FailedRuns:     utils.ToPtr(int32(0)),
				},
			},
			want: map[string]any{
				"type":      "image",
				"image":     map[string]any{"type": "internal", "name": "app", "tag": "v1"},
				"resources": map[string]any{"cpu": "1", "memory": "1G"},
				"retries":   float64(0),
				"retention": map[string]any{
					"ttl":            float64(0),
					"successfulRuns": float64(0),
					"failedRuns":     float64(0),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("JSON mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
