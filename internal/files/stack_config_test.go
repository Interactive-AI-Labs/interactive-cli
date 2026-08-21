package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
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
      memory: "256M"
      cpu: "1"
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
						Image: deployment.ImageSpec{
							Type: "internal",
							Name: "myapp",
							Tag:  "latest",
						},
						Resources: deployment.Resources{
							Memory: "256M",
							CPU:    "1",
						},
						Replicas: 2,
						Endpoint: true,
					},
				},
				Agents:    map[string]AgentConfig{},
				Databases: map[string]DatabaseConfig{},
				Mcps:      map[string]McpConfig{},
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
      memory: "128M"
      cpu: "1"
    autoscaling:
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
						Image: deployment.ImageSpec{
							Type:       "external",
							Repository: "nginx",
							Name:       "nginx",
							Tag:        "alpine",
						},
						Resources: deployment.Resources{
							Memory: "128M",
							CPU:    "1",
						},
						Autoscaling: &deployment.Autoscaling{
							MinReplicas:      2,
							MaxReplicas:      10,
							CPUPercentage:    utils.ToPtr(80),
							MemoryPercentage: utils.ToPtr(85),
						},
					},
				},
				Agents:    map[string]AgentConfig{},
				Databases: map[string]DatabaseConfig{},
				Mcps:      map[string]McpConfig{},
			},
		},
		{
			name: "valid config with databases",
			content: `organization: test-org
project: test-project
stack-id: stack-123
databases:
  my-db:
    instances: 2
    postgresVersion: "16"
    resources:
      cpu: "1"
      memory: "2G"
    storage:
      size: "20G"
    extensions:
      - vector
      - pg_trgm
    backup:
      schedule: "0 0 2 * * *"
      retentionpolicy: "30d"
`,
			want: &StackConfig{
				Organization: "test-org",
				Project:      "test-project",
				StackId:      "stack-123",
				Services:     map[string]ServiceConfig{},
				Agents:       map[string]AgentConfig{},
				Databases: map[string]DatabaseConfig{
					"my-db": {
						Instances:       2,
						PostgresVersion: "16",
						Resources:       deployment.Resources{CPU: "1", Memory: "2G"},
						Storage:         deployment.DatabaseStorageConfig{Size: "20G"},
						Extensions:      []string{"vector", "pg_trgm"},
						Backup: &deployment.DatabaseBackupConfig{
							Schedule:        "0 0 2 * * *",
							RetentionPolicy: "30d",
						},
					},
				},
				Mcps: map[string]McpConfig{},
			},
		},
		{
			name: "valid config with mcps",
			content: `organization: test-org
project: test-project
stack-id: stack-123
mcps:
  tools:
    type: internal
    port: 8080
    path: /mcp
    image:
      type: internal
      name: my-mcp
      tag: v1
    resources:
      cpu: "250m"
      memory: "512M"
    auth:
      type: none
`,
			want: &StackConfig{
				Organization: "test-org",
				Project:      "test-project",
				StackId:      "stack-123",
				Services:     map[string]ServiceConfig{},
				Agents:       map[string]AgentConfig{},
				Databases:    map[string]DatabaseConfig{},
				Mcps: map[string]McpConfig{
					"tools": {
						Type: "internal",
						Port: 8080,
						Path: "/mcp",
						Image: deployment.ImageSpec{
							Type: "internal",
							Name: "my-mcp",
							Tag:  "v1",
						},
						Resources: deployment.Resources{CPU: "250m", Memory: "512M"},
						Auth:      deployment.McpAuthBody{Type: "none"},
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
      memory: "256M"
      cpu: "1"
    replicas: 1
`,
			errContains: "stack-id is required",
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
		want    deployment.CreateServiceBody
	}{
		{
			name: "with fixed replicas",
			input: ServiceConfig{
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type: "internal",
					Name: "myapp",
					Tag:  "latest",
				},
				Resources: deployment.Resources{
					Memory: "256M",
					CPU:    "1",
				},
				Env: []deployment.EnvVar{
					{Name: "KEY1", Value: "value1"},
				},
				SecretRefs: []deployment.SecretRef{
					{SecretName: "my-secret"},
				},
				Endpoint: true,
				Replicas: 3,
			},
			stackId: "stack-123",
			want: deployment.CreateServiceBody{
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type: "internal",
					Name: "myapp",
					Tag:  "latest",
				},
				Resources: deployment.Resources{
					Memory: "256M",
					CPU:    "1",
				},
				Env: []deployment.EnvVar{
					{Name: "KEY1", Value: "value1"},
				},
				SecretRefs: []deployment.SecretRef{
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
				Image: deployment.ImageSpec{
					Type:       "external",
					Repository: "nginx",
					Name:       "nginx",
					Tag:        "latest",
				},
				Resources: deployment.Resources{
					Memory: "128M",
					CPU:    "1",
				},
				Autoscaling: &deployment.Autoscaling{
					MinReplicas:      2,
					MaxReplicas:      10,
					CPUPercentage:    utils.ToPtr(80),
					MemoryPercentage: utils.ToPtr(85),
				},
			},
			stackId: "stack-456",
			want: deployment.CreateServiceBody{
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type:       "external",
					Repository: "nginx",
					Name:       "nginx",
					Tag:        "latest",
				},
				Resources: deployment.Resources{
					Memory: "128M",
					CPU:    "1",
				},
				Replicas: 0,
				Autoscaling: &deployment.Autoscaling{
					MinReplicas:      2,
					MaxReplicas:      10,
					CPUPercentage:    utils.ToPtr(80),
					MemoryPercentage: utils.ToPtr(85),
				},
				StackId: "stack-456",
			},
		},
		{
			name: "nil autoscaling with replicas",
			input: ServiceConfig{
				ServicePort: 3000,
				Image: deployment.ImageSpec{
					Type: "internal",
					Name: "app",
					Tag:  "v1",
				},
				Resources: deployment.Resources{
					Memory: "512M",
					CPU:    "2",
				},
				Replicas: 5,
			},
			stackId: "stack-789",
			want: deployment.CreateServiceBody{
				ServicePort: 3000,
				Image: deployment.ImageSpec{
					Type: "internal",
					Name: "app",
					Tag:  "v1",
				},
				Resources: deployment.Resources{
					Memory: "512M",
					CPU:    "2",
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

func TestDatabaseConfigToCreateRequest(t *testing.T) {
	tests := []struct {
		name    string
		input   DatabaseConfig
		stackId string
		want    deployment.CreateDatabaseBody
	}{
		{
			name: "with backup",
			input: DatabaseConfig{
				Instances:       2,
				PostgresVersion: "16",
				Resources:       deployment.Resources{CPU: "1", Memory: "2G"},
				Storage:         deployment.DatabaseStorageConfig{Size: "20G"},
				Extensions:      []string{"vector", "pg_trgm"},
				Backup: &deployment.DatabaseBackupConfig{
					Schedule:        "0 0 2 * * *",
					RetentionPolicy: "30d",
				},
			},
			stackId: "stack-123",
			want: deployment.CreateDatabaseBody{
				Instances:       2,
				PostgresVersion: "16",
				Resources:       deployment.Resources{CPU: "1", Memory: "2G"},
				Storage:         deployment.DatabaseStorageConfig{Size: "20G"},
				Extensions:      []string{"vector", "pg_trgm"},
				Backup: &deployment.DatabaseBackupConfig{
					Schedule:        "0 0 2 * * *",
					RetentionPolicy: "30d",
				},
				StackId: "stack-123",
			},
		},
		{
			name: "without backup",
			input: DatabaseConfig{
				Instances: 1,
				Resources: deployment.Resources{CPU: "0.5", Memory: "1G"},
				Storage:   deployment.DatabaseStorageConfig{Size: "10G"},
			},
			stackId: "stack-456",
			want: deployment.CreateDatabaseBody{
				Instances: 1,
				Resources: deployment.Resources{CPU: "0.5", Memory: "1G"},
				Storage:   deployment.DatabaseStorageConfig{Size: "10G"},
				StackId:   "stack-456",
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

func TestMcpConfigToCreateRequest(t *testing.T) {
	tests := []struct {
		name  string
		input McpConfig
		want  deployment.CreateMcpBody
	}{
		{
			name: "internal",
			input: McpConfig{
				Type:      "internal",
				Port:      8080,
				Path:      "/mcp",
				Image:     deployment.ImageSpec{Type: "internal", Name: "my-mcp", Tag: "v1"},
				Resources: deployment.Resources{CPU: "250m", Memory: "512M"},
				Auth:      deployment.McpAuthBody{Type: "none"},
			},
			want: deployment.CreateMcpBody{
				Type:      "internal",
				Port:      8080,
				Path:      "/mcp",
				Image:     deployment.ImageSpec{Type: "internal", Name: "my-mcp", Tag: "v1"},
				Resources: deployment.Resources{CPU: "250m", Memory: "512M"},
				Auth:      deployment.McpAuthBody{Type: "none"},
				StackId:   "stack-123",
			},
		},
		{
			name: "external with bearer auth",
			input: McpConfig{
				Type:        "external",
				EndpointURL: "https://example.com/mcp",
				Auth:        deployment.McpAuthBody{Type: "bearer", Credential: "token"},
			},
			want: deployment.CreateMcpBody{
				Type:        "external",
				EndpointURL: "https://example.com/mcp",
				Auth:        deployment.McpAuthBody{Type: "bearer", Credential: "token"},
				StackId:     "stack-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.ToCreateRequest("stack-123")
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ToCreateRequest() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestServiceConfigFromDescribe(t *testing.T) {
	tests := []struct {
		name string
		desc *deployment.DescribeServiceResponse
		want ServiceConfig
	}{
		{
			name: "with endpoint",
			desc: &deployment.DescribeServiceResponse{
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type:       "external",
					Repository: "docker.io",
					Name:       "nginx",
					Tag:        "latest",
				},
				Resources:  deployment.Resources{Memory: "512M", CPU: "1"},
				Env:        []deployment.EnvVar{{Name: "K", Value: "V"}},
				SecretRefs: []deployment.SecretRef{{SecretName: "s"}},
				Endpoint:   "example.com",
				Replicas:   3,
			},
			want: ServiceConfig{
				ServicePort: 8080,
				Image: deployment.ImageSpec{
					Type:       "external",
					Repository: "docker.io",
					Name:       "nginx",
					Tag:        "latest",
				},
				Resources:  deployment.Resources{Memory: "512M", CPU: "1"},
				Env:        []deployment.EnvVar{{Name: "K", Value: "V"}},
				SecretRefs: []deployment.SecretRef{{SecretName: "s"}},
				Endpoint:   true,
				Replicas:   3,
			},
		},
		{
			name: "without endpoint",
			desc: &deployment.DescribeServiceResponse{ServicePort: 8080},
			want: ServiceConfig{ServicePort: 8080},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ServiceConfigFromDescribe(tt.desc)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ServiceConfigFromDescribe() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDatabaseConfigFromDescribe(t *testing.T) {
	tests := []struct {
		name string
		desc *deployment.DescribeDatabaseResponse
		want DatabaseConfig
	}{
		{
			name: "backup enabled",
			desc: &deployment.DescribeDatabaseResponse{
				Replicas:        2,
				PostgresVersion: "16",
				Resources:       deployment.Resources{CPU: "1", Memory: "2G"},
				Storage:         deployment.DatabaseStorageConfig{Size: "20G"},
				Extensions:      []string{"vector"},
				Backup: deployment.DatabaseBackupStatus{
					Enabled: true,
					DatabaseBackupConfig: deployment.DatabaseBackupConfig{
						Schedule:        "0 0 2 * * *",
						RetentionPolicy: "30d",
					},
				},
			},
			want: DatabaseConfig{
				Instances:       2,
				PostgresVersion: "16",
				Resources:       deployment.Resources{CPU: "1", Memory: "2G"},
				Storage:         deployment.DatabaseStorageConfig{Size: "20G"},
				Extensions:      []string{"vector"},
				Backup: &deployment.DatabaseBackupConfig{
					Schedule:        "0 0 2 * * *",
					RetentionPolicy: "30d",
				},
			},
		},
		{
			name: "backup disabled",
			desc: &deployment.DescribeDatabaseResponse{
				Replicas: 1,
				Backup:   deployment.DatabaseBackupStatus{Enabled: false},
			},
			want: DatabaseConfig{Instances: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DatabaseConfigFromDescribe(tt.desc)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("DatabaseConfigFromDescribe() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMcpConfigFromDescribe(t *testing.T) {
	tests := []struct {
		name string
		desc *deployment.DescribeMcpResponse
		want McpConfig
	}{
		{
			name: "internal hides generated endpoint",
			desc: &deployment.DescribeMcpResponse{
				McpOutput: deployment.McpOutput{
					Type:        "internal",
					EndpointURL: "http://tools.p1.svc.cluster.local:8080/mcp",
					Auth:        deployment.McpAuthInfo{Type: "none"},
				},
				Port:      8080,
				Path:      "/mcp",
				Image:     deployment.ImageSpec{Type: "internal", Name: "my-mcp", Tag: "v1"},
				Resources: deployment.Resources{CPU: "250m", Memory: "512M"},
				Env:       []deployment.EnvVar{{Name: "K", Value: "V"}},
			},
			want: McpConfig{
				Type:      "internal",
				Port:      8080,
				Path:      "/mcp",
				Image:     deployment.ImageSpec{Type: "internal", Name: "my-mcp", Tag: "v1"},
				Resources: deployment.Resources{CPU: "250m", Memory: "512M"},
				Env:       []deployment.EnvVar{{Name: "K", Value: "V"}},
				Auth:      deployment.McpAuthBody{Type: "none"},
			},
		},
		{
			name: "external keeps endpoint and auth routing",
			desc: &deployment.DescribeMcpResponse{
				McpOutput: deployment.McpOutput{
					Type:        "external",
					EndpointURL: "https://example.com/mcp",
					CatalogID:   "github",
					Auth: deployment.McpAuthInfo{
						Type:         "custom",
						Header:       "X-Token",
						HeaderPrefix: "Token ",
					},
				},
				Headers: map[string]string{"X-Team": "dev"},
			},
			want: McpConfig{
				Type:        "external",
				EndpointURL: "https://example.com/mcp",
				CatalogID:   "github",
				Auth: deployment.McpAuthBody{
					Type:         "custom",
					Header:       "X-Token",
					HeaderPrefix: "Token ",
				},
				Headers: map[string]string{"X-Team": "dev"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := McpConfigFromDescribe(tt.desc)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("McpConfigFromDescribe() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMcpAuthYAMLOmitsEmptyFields(t *testing.T) {
	config := StackConfig{
		Organization: "test-org",
		Project:      "test-project",
		StackId:      "stack-123",
		Mcps: map[string]McpConfig{
			"tools": {Auth: deployment.McpAuthBody{Type: "none"}},
		},
	}

	got, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("yaml.Marshal() unexpected error = %v", err)
	}

	want := `organization: test-org
project: test-project
stack-id: stack-123
services: {}
agents: {}
databases: {}
mcps:
    tools:
        auth:
            type: none
`
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("yaml.Marshal() mismatch (-want +got):\n%s", diff)
	}
}
