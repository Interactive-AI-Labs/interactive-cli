package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintServiceList(t *testing.T) {
	tests := []struct {
		name     string
		services []clients.ServiceOutput
		want     string
	}{
		{
			name:     "empty list prints headers only",
			services: []clients.ServiceOutput{},
			want:     "NAME   REVISION   STATUS   ENDPOINT   UPDATED\n",
		},
		{
			name: "single service",
			services: []clients.ServiceOutput{
				{
					Name:     "web",
					Revision: 3,
					Status:   "Running",
					Endpoint: "web-abc123.interactive.ai",
					Updated:  "2024-01-15",
				},
			},
			want: "NAME   REVISION   STATUS    ENDPOINT                    UPDATED\n" +
				"web    3          Running   web-abc123.interactive.ai   2024-01-15\n",
		},
		{
			name: "multiple services with empty fields",
			services: []clients.ServiceOutput{
				{Name: "api", Revision: 1, Status: "Running"},
				{Name: "worker", Revision: 5, Status: "Deploying"},
			},
			want: "NAME     REVISION   STATUS      ENDPOINT   UPDATED\n" +
				"api      1          Running                \n" +
				"worker   5          Deploying              \n",
		},
		{
			name: "revision zero",
			services: []clients.ServiceOutput{
				{Name: "new-svc", Revision: 0, Status: "Pending"},
			},
			want: "NAME      REVISION   STATUS    ENDPOINT   UPDATED\n" +
				"new-svc   0          Pending              \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintServiceList(&buf, tt.services)
			if err != nil {
				t.Fatalf("PrintServiceList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintSyncResult(t *testing.T) {
	tests := []struct {
		name    string
		label   string
		created []string
		updated []string
		deleted []string
		skipped []string
		want    string
	}{
		{
			name:  "no changes",
			label: "services",
			want:  "No changes required; services already match config.\n",
		},
		{
			name:    "empty slices count as no changes",
			label:   "services",
			created: []string{},
			updated: []string{},
			deleted: []string{},
			skipped: []string{},
			want:    "No changes required; services already match config.\n",
		},
		{
			name:    "created only",
			label:   "services",
			created: []string{"api", "web"},
			want:    "Created services: api, web\n",
		},
		{
			name:    "updated only",
			label:   "services",
			updated: []string{"worker"},
			want:    "Updated services: worker\n",
		},
		{
			name:    "deleted only",
			label:   "services",
			deleted: []string{"old-svc"},
			want:    "Deleted services: old-svc\n",
		},
		{
			name:    "all three",
			label:   "services",
			created: []string{"new-svc"},
			updated: []string{"api"},
			deleted: []string{"legacy"},
			want: "Created services: new-svc\n" +
				"Updated services: api\n" +
				"Deleted services: legacy\n",
		},
		{
			name:    "vector stores label - created and deleted",
			label:   "vector stores",
			created: []string{"knowledge-base"},
			deleted: []string{"old-vs"},
			want: "Created vector stores: knowledge-base\n" +
				"Deleted vector stores: old-vs\n",
		},
		{
			name:  "no changes - vector stores",
			label: "vector stores",
			want:  "No changes required; vector stores already match config.\n",
		},
		{
			name:    "skipped only",
			label:   "vector stores",
			skipped: []string{"my-store"},
			want:    "Skipped vector stores (already exist, updates not supported): my-store\n",
		},
		{
			name:    "created and skipped",
			label:   "vector stores",
			created: []string{"new-vs"},
			skipped: []string{"existing-vs"},
			want: "Created vector stores: new-vs\n" +
				"Skipped vector stores (already exist, updates not supported): existing-vs\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintSyncResult(&buf, tt.label, tt.created, tt.updated, tt.deleted, tt.skipped)
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
