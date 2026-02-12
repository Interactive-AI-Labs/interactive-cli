package output

import (
	"bytes"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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
		created []string
		updated []string
		deleted []string
		want    string
	}{
		{
			name: "no changes",
			want: "No changes required; services already match config.\n",
		},
		{
			name:    "empty slices count as no changes",
			created: []string{},
			updated: []string{},
			deleted: []string{},
			want:    "No changes required; services already match config.\n",
		},
		{
			name:    "created only",
			created: []string{"api", "web"},
			want:    "Created services: api, web\n",
		},
		{
			name:    "updated only",
			updated: []string{"worker"},
			want:    "Updated services: worker\n",
		},
		{
			name:    "deleted only",
			deleted: []string{"old-svc"},
			want:    "Deleted services: old-svc\n",
		},
		{
			name:    "all three",
			created: []string{"new-svc"},
			updated: []string{"api"},
			deleted: []string{"legacy"},
			want: "Created services: new-svc\n" +
				"Updated services: api\n" +
				"Deleted services: legacy\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintSyncResult(&buf, tt.created, tt.updated, tt.deleted)
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
