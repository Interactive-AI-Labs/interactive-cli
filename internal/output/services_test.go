package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
)

func TestPrintServiceList(t *testing.T) {
	tests := []struct {
		name     string
		services []deployment.ServiceOutput
		want     string
	}{
		{
			name:     "empty list prints message",
			services: []deployment.ServiceOutput{},
			want:     "No services found.\n",
		},
		{
			name: "single service",
			services: []deployment.ServiceOutput{
				{
					Name:     "web",
					Revision: 3,
					Status:   "Running",
					Updated:  "2024-01-15",
				},
			},
			want: "NAME   REVISION   STATUS    UPDATED\n" +
				"web    3          Running   2024-01-15\n",
		},
		{
			name: "multiple services with empty fields",
			services: []deployment.ServiceOutput{
				{Name: "api", Revision: 1, Status: "Running"},
				{Name: "worker", Revision: 5, Status: "Deploying"},
			},
			want: "NAME     REVISION   STATUS      UPDATED\n" +
				"api      1          Running     \n" +
				"worker   5          Deploying   \n",
		},
		{
			name: "revision zero",
			services: []deployment.ServiceOutput{
				{Name: "new-svc", Revision: 0, Status: "Pending"},
			},
			want: "NAME      REVISION   STATUS    UPDATED\n" +
				"new-svc   0          Pending   \n",
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

func TestPrintServiceDescribe(t *testing.T) {
	tests := []struct {
		name string
		svc  *deployment.DescribeServiceResponse
		want string
	}{
		{
			name: "minimal service with replicas",
			svc: &deployment.DescribeServiceResponse{
				ServiceOutput: deployment.ServiceOutput{
					Name:      "minimal-svc",
					ProjectId: "proj-123",
					Revision:  1,
					Status:    "deployed",
				},
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type: "external",
					Name: "nginx",
					Tag:  "latest",
				},
				Resources: deployment.Resources{Memory: "128M", CPU: "100m"},
				Replicas:  1,
			},
			want: "Name:       minimal-svc\n" +
				"Revision:   1\n" +
				"Status:     deployed\n" +
				"Port:       8080\n" +
				"Image:\n" +
				"  Type:   external\n" +
				"  Name:   nginx\n" +
				"  Tag:    latest\n" +
				"Resources:\n" +
				"  CPU:      100m\n" +
				"  Memory:   128M\n" +
				"Replicas:   1\n",
		},
		{
			name: "zero replicas prints zero",
			svc: &deployment.DescribeServiceResponse{
				ServiceOutput: deployment.ServiceOutput{
					Name:   "scaled-svc",
					Status: "deployed",
				},
				ServicePort: 80,
				Image: deployment.ImageSpec{
					Type: "external",
					Name: "app",
					Tag:  "1.0",
				},
				Resources: deployment.Resources{Memory: "256M", CPU: "200m"},
				Replicas:  0,
			},
			want: "Name:       scaled-svc\n" +
				"Revision:   0\n" +
				"Status:     deployed\n" +
				"Port:       80\n" +
				"Image:\n" +
				"  Type:   external\n" +
				"  Name:   app\n" +
				"  Tag:    1.0\n" +
				"Resources:\n" +
				"  CPU:      200m\n" +
				"  Memory:   256M\n" +
				"Replicas:   0\n",
		},
		{
			name: "autoscaling with CPU only",
			svc: &deployment.DescribeServiceResponse{
				ServiceOutput: deployment.ServiceOutput{
					Name:   "autoscaling-svc",
					Status: "deployed",
				},
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type: "external",
					Name: "redis",
					Tag:  "7",
				},
				Resources: deployment.Resources{Memory: "256M", CPU: "200m"},
				Autoscaling: &deployment.Autoscaling{
					MinReplicas:   2,
					MaxReplicas:   5,
					CPUPercentage: utils.ToPtr(70),
				},
			},
			want: "Name:       autoscaling-svc\n" +
				"Revision:   0\n" +
				"Status:     deployed\n" +
				"Port:       8080\n" +
				"Image:\n" +
				"  Type:   external\n" +
				"  Name:   redis\n" +
				"  Tag:    7\n" +
				"Resources:\n" +
				"  CPU:      200m\n" +
				"  Memory:   256M\n" +
				"\n" +
				"Autoscaling:\n" +
				"  Min Replicas:   2\n" +
				"  Max Replicas:   5\n" +
				"  CPU%:           70\n",
		},
		{
			name: "full-featured service",
			svc: &deployment.DescribeServiceResponse{
				ServiceOutput: deployment.ServiceOutput{
					Name:      "full-svc",
					ProjectId: "proj-456",
					Revision:  10,
					Status:    "deployed",
				},
				Endpoint:    "full-svc-abc.dev.interactive.ai",
				ServicePort: 443,
				Image: deployment.ImageSpec{
					Type:       "platform",
					Repository: "agents",
					Name:       "api-gateway",
					Tag:        "1.0.0",
				},
				Resources: deployment.Resources{Memory: "1G", CPU: "1"},
				Env: []deployment.EnvVar{
					{Name: "LOG_LEVEL", Value: "info"},
				},
				SecretRefs: []deployment.SecretRef{
					{SecretName: "api-keys"},
					{SecretName: "db-creds"},
				},
				StackId: "my-stack",
				Autoscaling: &deployment.Autoscaling{
					MinReplicas:      2,
					MaxReplicas:      8,
					CPUPercentage:    utils.ToPtr(65),
					MemoryPercentage: utils.ToPtr(80),
				},
				Healthcheck: &deployment.Healthcheck{
					Path:                "/healthz",
					InitialDelaySeconds: utils.ToPtr(20),
				},
				Schedule: &deployment.Schedule{
					Uptime:   "Mon-Fri 08:00-20:00",
					Timezone: "UTC",
				},
			},
			want: "Name:       full-svc\n" +
				"Stack Id:   my-stack\n" +
				"Revision:   10\n" +
				"Status:     deployed\n" +
				"Port:       443\n" +
				"Image:\n" +
				"  Type:         platform\n" +
				"  Name:         api-gateway\n" +
				"  Tag:          1.0.0\n" +
				"  Repository:   agents\n" +
				"Resources:\n" +
				"  CPU:      1\n" +
				"  Memory:   1G\n" +
				"\n" +
				"Autoscaling:\n" +
				"  Min Replicas:   2\n" +
				"  Max Replicas:   8\n" +
				"  CPU%:           65\n" +
				"  Memory%:        80\n" +
				"Endpoint:         full-svc-abc.dev.interactive.ai\n" +
				"\n" +
				"Healthcheck:\n" +
				"  Path:            /healthz\n" +
				"  Initial Delay:   20s\n" +
				"\n" +
				"Environment:\n" +
				"  LOG_LEVEL=info\n" +
				"\n" +
				"Secrets:   api-keys, db-creds\n" +
				"\n" +
				"Schedule:\n" +
				"  Uptime:     Mon-Fri 08:00-20:00\n" +
				"  Timezone:   UTC\n",
		},
		{
			name: "service with message",
			svc: &deployment.DescribeServiceResponse{
				ServiceOutput: deployment.ServiceOutput{
					Name:     "msg-svc",
					Revision: 2,
					Status:   "deployed",
				},
				Message:     "rollout completed",
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type: "external",
					Name: "nginx",
					Tag:  "latest",
				},
				Resources: deployment.Resources{Memory: "128M", CPU: "100m"},
				Replicas:  1,
			},
			want: "Name:       msg-svc\n" +
				"Revision:   2\n" +
				"Status:     deployed\n" +
				"Message:    rollout completed\n" +
				"Port:       8080\n" +
				"Image:\n" +
				"  Type:   external\n" +
				"  Name:   nginx\n" +
				"  Tag:    latest\n" +
				"Resources:\n" +
				"  CPU:      100m\n" +
				"  Memory:   128M\n" +
				"Replicas:   1\n",
		},
		{
			name: "schedule with downtime",
			svc: &deployment.DescribeServiceResponse{
				ServiceOutput: deployment.ServiceOutput{
					Name:   "downtime-svc",
					Status: "deployed",
				},
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type: "external",
					Name: "app",
					Tag:  "v2",
				},
				Resources: deployment.Resources{Memory: "128M", CPU: "50m"},
				Replicas:  1,
				Schedule: &deployment.Schedule{
					Downtime: "Sat-Sun 00:00-24:00",
					Timezone: "US/Eastern",
				},
			},
			want: "Name:       downtime-svc\n" +
				"Revision:   0\n" +
				"Status:     deployed\n" +
				"Port:       8080\n" +
				"Image:\n" +
				"  Type:   external\n" +
				"  Name:   app\n" +
				"  Tag:    v2\n" +
				"Resources:\n" +
				"  CPU:      50m\n" +
				"  Memory:   128M\n" +
				"Replicas:   1\n" +
				"\n" +
				"Schedule:\n" +
				"  Downtime:   Sat-Sun 00:00-24:00\n" +
				"  Timezone:   US/Eastern\n",
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

func TestPrintServiceRevision(t *testing.T) {
	tests := []struct {
		name string
		rev  *deployment.ServiceRevisionResponse
		want string
	}{
		{
			name: "minimal revision",
			rev: &deployment.ServiceRevisionResponse{
				RevisionMeta: deployment.RevisionMeta{Revision: 2, Status: "deployed"},
				ServicePort:  8080,
				Image:        deployment.ImageSpec{Type: "external", Name: "nginx", Tag: "latest"},
				Resources:    deployment.Resources{Memory: "128M", CPU: "100m"},
				Replicas:     1,
			},
			want: "Revision:   2\n" +
				"Status:     deployed\n" +
				"By:         —\n" +
				"Source:     —\n" +
				"Port:       8080\n" +
				"Image:\n" +
				"  Type:   external\n" +
				"  Name:   nginx\n" +
				"  Tag:    latest\n" +
				"Resources:\n" +
				"  CPU:      100m\n" +
				"  Memory:   128M\n" +
				"Replicas:   1\n",
		},
		{
			name: "revision with endpoint and env",
			rev: &deployment.ServiceRevisionResponse{
				RevisionMeta: deployment.RevisionMeta{
					Revision: 5,
					Status:   "deployed",
					Updated:  "2024-06-01",
					Actor: &deployment.RevisionActor{
						Type:        "service",
						DisplayName: "deployment-controller",
					},
					Source: &deployment.RevisionSource{Type: "controller"},
				},
				ServicePort: 443,
				Image: deployment.ImageSpec{
					Type:       "platform",
					Name:       "api",
					Tag:        "2.0",
					Repository: "apps",
				},
				Resources: deployment.Resources{Memory: "512M", CPU: "500m"},
				Endpoint:  "api.interactive.ai",
				Env: []deployment.EnvVar{
					{Name: "LOG_LEVEL", Value: "info"},
				},
			},
			want: "Revision:   5\n" +
				"Status:     deployed\n" +
				"Updated:    2024-06-01\n" +
				"By:         deployment-controller (service)\n" +
				"Source:     controller\n" +
				"Port:       443\n" +
				"Image:\n" +
				"  Type:         platform\n" +
				"  Name:         api\n" +
				"  Tag:          2.0\n" +
				"  Repository:   apps\n" +
				"Resources:\n" +
				"  CPU:      500m\n" +
				"  Memory:   512M\n" +
				"Replicas:   0\n" +
				"Endpoint:   api.interactive.ai\n" +
				"\n" +
				"Environment:\n" +
				"  LOG_LEVEL=info\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintServiceRevision(&buf, tt.rev)
			if err != nil {
				t.Fatalf("PrintServiceRevision() error = %v", err)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintSyncResult(&buf, tt.label, tt.created, tt.updated, tt.deleted)
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
