package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestLoadStackConfig(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		want           *StackConfig
		wantErr        bool
		errContains    string
		useNonexistent bool
	}{
		{
			name: "valid config with replicas",
			content: `organization: test-org
project: test-project
stack-id: stack-123
services:
  web:
    version: v1
    servicePort: 8080
    image:
      type: internal
      name: myapp
      tag: latest
    resources:
      memory: 256Mi
      cpu: 100m
    replicas: 2
    endpoint: true
`,
			want: &StackConfig{
				Organization: "test-org",
				Project:      "test-project",
				StackId:      "stack-123",
				Services: map[string]ServiceConfig{
					"web": {
						Version:     "v1",
						ServicePort: 8080,
						Image: clients.ImageSpec{
							Type: "internal",
							Name: "myapp",
							Tag:  "latest",
						},
						Resources: clients.Resources{
							Memory: "256Mi",
							CPU:    "100m",
						},
						Replicas: 2,
						Endpoint: true,
					},
				},
			},
		},
		{
			name: "valid config with autoscaling",
			content: `organization: my-org
project: my-project
stack-id: stack-456
services:
  api:
    servicePort: 3000
    image:
      type: external
      repository: nginx
      name: nginx
      tag: alpine
    resources:
      memory: 128Mi
      cpu: 50m
    autoscaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 10
      cpuPercentage: 80
      memoryPercentage: 85
`,
			want: &StackConfig{
				Organization: "my-org",
				Project:      "my-project",
				StackId:      "stack-456",
				Services: map[string]ServiceConfig{
					"api": {
						ServicePort: 3000,
						Image: clients.ImageSpec{
							Type:       "external",
							Repository: "nginx",
							Name:       "nginx",
							Tag:        "alpine",
						},
						Resources: clients.Resources{
							Memory: "128Mi",
							CPU:    "50m",
						},
						Autoscaling: &clients.Autoscaling{
							Enabled:          true,
							MinReplicas:      2,
							MaxReplicas:      10,
							CPUPercentage:    80,
							MemoryPercentage: 85,
						},
					},
				},
			},
		},
		{
			name:           "file does not exist",
			useNonexistent: true,
			wantErr:        true,
			errContains:    "failed to read config file",
		},
		{
			name: "invalid YAML",
			content: `organization: test-org
services: [invalid, yaml: syntax}`,
			wantErr:     true,
			errContains: "failed to parse YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var configFile string

			if tt.useNonexistent {
				configFile = "/nonexistent/file.yaml"
			} else {
				tmpDir := t.TempDir()
				configFile = filepath.Join(tmpDir, "stack.yaml")
				if err := os.WriteFile(configFile, []byte(tt.content), 0o600); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			}

			got, err := LoadStackConfig(configFile)

			if tt.wantErr {
				if err == nil {
					t.Fatal("LoadStackConfig() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error should contain %q, got: %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("LoadStackConfig() unexpected error = %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("LoadStackConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}

	validationTests := []struct {
		name        string
		config      string
		errContains string
	}{
		{
			name: "missing stack-id with services",
			config: `organization: test-org
project: test-project
services:
  web:
    servicePort: 8080
    image:
      type: internal
      name: myapp
      tag: latest
    resources:
      memory: 256Mi
      cpu: 100m
    replicas: 1
`,
			errContains: "stack-id is required",
		},
		{
			name: "invalid service port",
			config: `organization: test-org
project: test-project
stack-id: stack-123
services:
  web:
    servicePort: 0
    image:
      type: internal
      name: myapp
      tag: latest
    resources:
      memory: 256Mi
      cpu: 100m
    replicas: 1
`,
			errContains: "servicePort must be greater than zero",
		},
		{
			name: "invalid image type",
			config: `organization: test-org
project: test-project
stack-id: stack-123
services:
  web:
    servicePort: 8080
    image:
      type: invalid
      name: myapp
      tag: latest
    resources:
      memory: 256Mi
      cpu: 100m
    replicas: 1
`,
			errContains: "must be 'internal' or 'external'",
		},
		{
			name: "both replicas and autoscaling",
			config: `organization: test-org
project: test-project
stack-id: stack-123
services:
  web:
    servicePort: 8080
    image:
      type: internal
      name: myapp
      tag: latest
    resources:
      memory: 256Mi
      cpu: 100m
    replicas: 2
    autoscaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 10
      cpuPercentage: 80
`,
			errContains: "cannot set both replicas and autoscaling.enabled",
		},
		{
			name: "external image without repository",
			config: `organization: test-org
project: test-project
stack-id: stack-123
services:
  web:
    servicePort: 8080
    image:
      type: external
      name: myapp
      tag: latest
    resources:
      memory: 256Mi
      cpu: 100m
    replicas: 1
`,
			errContains: "image.repository is required for external images",
		},
		{
			name: "neither replicas nor autoscaling",
			config: `organization: test-org
project: test-project
stack-id: stack-123
services:
  web:
    servicePort: 8080
    image:
      type: internal
      name: myapp
      tag: latest
    resources:
      memory: 256Mi
      cpu: 100m
`,
			errContains: "must specify either replicas or autoscaling",
		},
	}

	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "stack.yaml")

			if err := os.WriteFile(configFile, []byte(tt.config), 0o600); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			_, err := LoadStackConfig(configFile)
			if err == nil {
				t.Fatal("LoadStackConfig() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error containing %q, got: %v", tt.errContains, err)
			}
		})
	}
}

func TestServiceConfigToCreateRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   ServiceConfig
		stackId string
		want    clients.CreateServiceBody
	}{
		{
			name: "with fixed replicas",
			input: ServiceConfig{
				ServicePort: 8080,
				Image: clients.ImageSpec{
					Type: "internal",
					Name: "myapp",
					Tag:  "latest",
				},
				Resources: clients.Resources{
					Memory: "256Mi",
					CPU:    "100m",
				},
				Env: []clients.EnvVar{
					{Name: "KEY1", Value: "value1"},
				},
				SecretRefs: []clients.SecretRef{
					{SecretName: "my-secret"},
				},
				Endpoint: true,
				Replicas: 3,
			},
			stackId: "stack-123",
			want: clients.CreateServiceBody{
				ServicePort: 8080,
				Image: clients.ImageSpec{
					Type: "internal",
					Name: "myapp",
					Tag:  "latest",
				},
				Resources: clients.Resources{
					Memory: "256Mi",
					CPU:    "100m",
				},
				Env: []clients.EnvVar{
					{Name: "KEY1", Value: "value1"},
				},
				SecretRefs: []clients.SecretRef{
					{SecretName: "my-secret"},
				},
				Endpoint:    true,
				Replicas:    3,
				Autoscaling: nil,
				StackId:     "stack-123",
			},
		},
		{
			name: "with autoscaling",
			input: ServiceConfig{
				ServicePort: 8080,
				Image: clients.ImageSpec{
					Type:       "external",
					Repository: "nginx",
					Name:       "nginx",
					Tag:        "latest",
				},
				Resources: clients.Resources{
					Memory: "128Mi",
					CPU:    "50m",
				},
				Autoscaling: &clients.Autoscaling{
					Enabled:          true,
					MinReplicas:      2,
					MaxReplicas:      10,
					CPUPercentage:    80,
					MemoryPercentage: 85,
				},
			},
			stackId: "stack-456",
			want: clients.CreateServiceBody{
				ServicePort: 8080,
				Image: clients.ImageSpec{
					Type:       "external",
					Repository: "nginx",
					Name:       "nginx",
					Tag:        "latest",
				},
				Resources: clients.Resources{
					Memory: "128Mi",
					CPU:    "50m",
				},
				Replicas: 0,
				Autoscaling: &clients.Autoscaling{
					Enabled:          true,
					MinReplicas:      2,
					MaxReplicas:      10,
					CPUPercentage:    80,
					MemoryPercentage: 85,
				},
				StackId: "stack-456",
			},
		},
		{
			name: "autoscaling disabled with replicas",
			input: ServiceConfig{
				ServicePort: 3000,
				Image: clients.ImageSpec{
					Type: "internal",
					Name: "app",
					Tag:  "v1",
				},
				Resources: clients.Resources{
					Memory: "512Mi",
					CPU:    "250m",
				},
				Autoscaling: &clients.Autoscaling{
					Enabled: false,
				},
				Replicas: 5,
			},
			stackId: "stack-789",
			want: clients.CreateServiceBody{
				ServicePort: 3000,
				Image: clients.ImageSpec{
					Type: "internal",
					Name: "app",
					Tag:  "v1",
				},
				Resources: clients.Resources{
					Memory: "512Mi",
					CPU:    "250m",
				},
				Replicas:    5,
				Autoscaling: nil,
				StackId:     "stack-789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.ToCreateRequest(tt.stackId)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToCreateRequest() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
