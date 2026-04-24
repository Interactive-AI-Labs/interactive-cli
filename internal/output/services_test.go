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
			name:     "empty list prints message",
			services: []clients.ServiceOutput{},
			want:     "No services found.\n",
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

func intPtr(i int) *int { return &i }

func TestPrintServiceDescribe(t *testing.T) {
	tests := []struct {
		name string
		svc  *clients.DescribeServiceResponse
		want string
	}{
		{
			name: "minimal service with replicas",
			svc: &clients.DescribeServiceResponse{
				ServiceOutput: clients.ServiceOutput{
					Name:      "minimal-svc",
					ProjectId: "proj-123",
					Revision:  1,
					Status:    "deployed",
				},
				ServicePort: 8080,
				Image: clients.ImageSpec{
					Type: "external",
					Name: "nginx",
					Tag:  "latest",
				},
				Resources: clients.Resources{Memory: "128M", CPU: "100m"},
				Replicas:  1,
			},
			want: "Name:        minimal-svc\n" +
				"Project Id:  proj-123\n" +
				"Revision:    1\n" +
				"Status:      deployed\n" +
				"Port:        8080\n" +
				"Image:\n" +
				"  Type:       external\n" +
				"  Name:       nginx\n" +
				"  Tag:        latest\n" +
				"Resources:\n" +
				"  CPU:     100m\n" +
				"  Memory:  128M\n" +
				"Replicas:    1\n",
		},
		{
			name: "zero replicas prints zero",
			svc: &clients.DescribeServiceResponse{
				ServiceOutput: clients.ServiceOutput{
					Name:   "scaled-svc",
					Status: "deployed",
				},
				ServicePort: 80,
				Image: clients.ImageSpec{
					Type: "external",
					Name: "app",
					Tag:  "1.0",
				},
				Resources: clients.Resources{Memory: "256M", CPU: "200m"},
				Replicas:  0,
			},
			want: "Name:        scaled-svc\n" +
				"Project Id:  \n" +
				"Revision:    0\n" +
				"Status:      deployed\n" +
				"Port:        80\n" +
				"Image:\n" +
				"  Type:       external\n" +
				"  Name:       app\n" +
				"  Tag:        1.0\n" +
				"Resources:\n" +
				"  CPU:     200m\n" +
				"  Memory:  256M\n" +
				"Replicas:    0\n",
		},
		{
			name: "autoscaling with CPU only",
			svc: &clients.DescribeServiceResponse{
				ServiceOutput: clients.ServiceOutput{
					Name:   "autoscaling-svc",
					Status: "deployed",
				},
				ServicePort: 8080,
				Image: clients.ImageSpec{
					Type: "external",
					Name: "redis",
					Tag:  "7",
				},
				Resources: clients.Resources{Memory: "256M", CPU: "200m"},
				Autoscaling: &clients.Autoscaling{
					MinReplicas:   2,
					MaxReplicas:   5,
					CPUPercentage: intPtr(70),
				},
			},
			want: "Name:        autoscaling-svc\n" +
				"Project Id:  \n" +
				"Revision:    0\n" +
				"Status:      deployed\n" +
				"Port:        8080\n" +
				"Image:\n" +
				"  Type:       external\n" +
				"  Name:       redis\n" +
				"  Tag:        7\n" +
				"Resources:\n" +
				"  CPU:     200m\n" +
				"  Memory:  256M\n" +
				"\n" +
				"Autoscaling:\n" +
				"  Min Replicas: 2\n" +
				"  Max Replicas: 5\n" +
				"  CPU%:         70\n",
		},
		{
			name: "full-featured service",
			svc: &clients.DescribeServiceResponse{
				ServiceOutput: clients.ServiceOutput{
					Name:      "full-svc",
					ProjectId: "proj-456",
					Revision:  10,
					Status:    "deployed",
					Endpoint:  "full-svc-abc.dev.interactive.ai",
				},
				ServicePort: 443,
				Image: clients.ImageSpec{
					Type:       "platform",
					Repository: "agents",
					Name:       "api-gateway",
					Tag:        "1.0.0",
				},
				Resources: clients.Resources{Memory: "1G", CPU: "1"},
				Env: []clients.EnvVar{
					{Name: "LOG_LEVEL", Value: "info"},
				},
				SecretRefs: []clients.SecretRef{
					{SecretName: "api-keys"},
					{SecretName: "db-creds"},
				},
				StackId: "my-stack",
				Autoscaling: &clients.Autoscaling{
					MinReplicas:      2,
					MaxReplicas:      8,
					CPUPercentage:    intPtr(65),
					MemoryPercentage: intPtr(80),
				},
				Healthcheck: &clients.Healthcheck{
					Path:                "/healthz",
					InitialDelaySeconds: intPtr(20),
				},
				Schedule: &clients.Schedule{
					Uptime:   "Mon-Fri 08:00-20:00",
					Timezone: "UTC",
				},
			},
			want: "Name:        full-svc\n" +
				"Project Id:  proj-456\n" +
				"Revision:    10\n" +
				"Status:      deployed\n" +
				"Port:        443\n" +
				"Image:\n" +
				"  Type:       platform\n" +
				"  Name:       api-gateway\n" +
				"  Tag:        1.0.0\n" +
				"  Repository: agents\n" +
				"Resources:\n" +
				"  CPU:     1\n" +
				"  Memory:  1G\n" +
				"\n" +
				"Autoscaling:\n" +
				"  Min Replicas: 2\n" +
				"  Max Replicas: 8\n" +
				"  CPU%:         65\n" +
				"  Memory%:      80\n" +
				"Endpoint:    full-svc-abc.dev.interactive.ai\n" +
				"Stack Id:    my-stack\n" +
				"\n" +
				"Healthcheck:\n" +
				"  Path:           /healthz\n" +
				"  Initial Delay:  20s\n" +
				"\n" +
				"Environment:\n" +
				"  LOG_LEVEL=info\n" +
				"\n" +
				"Secrets:     api-keys, db-creds\n" +
				"\n" +
				"Schedule:\n" +
				"  Uptime:   Mon-Fri 08:00-20:00\n" +
				"  Timezone: UTC\n",
		},
		{
			name: "schedule with downtime",
			svc: &clients.DescribeServiceResponse{
				ServiceOutput: clients.ServiceOutput{
					Name:   "downtime-svc",
					Status: "deployed",
				},
				ServicePort: 8080,
				Image: clients.ImageSpec{
					Type: "external",
					Name: "app",
					Tag:  "v2",
				},
				Resources: clients.Resources{Memory: "128M", CPU: "50m"},
				Replicas:  1,
				Schedule: &clients.Schedule{
					Downtime: "Sat-Sun 00:00-24:00",
					Timezone: "US/Eastern",
				},
			},
			want: "Name:        downtime-svc\n" +
				"Project Id:  \n" +
				"Revision:    0\n" +
				"Status:      deployed\n" +
				"Port:        8080\n" +
				"Image:\n" +
				"  Type:       external\n" +
				"  Name:       app\n" +
				"  Tag:        v2\n" +
				"Resources:\n" +
				"  CPU:     50m\n" +
				"  Memory:  128M\n" +
				"Replicas:    1\n" +
				"\n" +
				"Schedule:\n" +
				"  Downtime: Sat-Sun 00:00-24:00\n" +
				"  Timezone: US/Eastern\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintServiceDescribe(&buf, tt.svc)
			if err != nil {
				t.Fatalf("PrintServiceDescribe() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", got, tt.want)
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
