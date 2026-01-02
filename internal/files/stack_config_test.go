package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestLoadStackConfig(t *testing.T) {
	t.Run("loads valid stack config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "stack.yaml")
		content := `organization: test-org
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
      requests:
        memory: 256Mi
        cpu: 100m
      limits:
        memory: 512Mi
        cpu: 200m
    replicas: 2
    endpoint: true
`
		if err := os.WriteFile(configFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		cfg, err := LoadStackConfig(configFile)
		if err != nil {
			t.Fatalf("LoadStackConfig() error = %v", err)
		}

		if cfg.Organization != "test-org" {
			t.Errorf("Organization = %q, want %q", cfg.Organization, "test-org")
		}
		if cfg.Project != "test-project" {
			t.Errorf("Project = %q, want %q", cfg.Project, "test-project")
		}
		if cfg.StackId != "stack-123" {
			t.Errorf("StackId = %q, want %q", cfg.StackId, "stack-123")
		}

		if len(cfg.Services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(cfg.Services))
		}

		web, ok := cfg.Services["web"]
		if !ok {
			t.Fatal("service 'web' not found")
		}

		if web.ServicePort != 8080 {
			t.Errorf("web.ServicePort = %d, want 8080", web.ServicePort)
		}
		if web.Image.Type != "internal" {
			t.Errorf("web.Image.Type = %q, want 'internal'", web.Image.Type)
		}
		if web.Replicas != 2 {
			t.Errorf("web.Replicas = %d, want 2", web.Replicas)
		}
		if !web.Endpoint {
			t.Error("web.Endpoint = false, want true")
		}
	})

	t.Run("returns error when file does not exist", func(t *testing.T) {
		_, err := LoadStackConfig("/nonexistent/file.yaml")
		if err == nil {
			t.Fatal("LoadStackConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read config file") {
			t.Errorf("error should mention 'failed to read config file', got: %v", err)
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "stack.yaml")
		content := `organization: test-org
services: [invalid, yaml: syntax}`
		if err := os.WriteFile(configFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := LoadStackConfig(configFile)
		if err == nil {
			t.Fatal("LoadStackConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse YAML") {
			t.Errorf("error should mention 'failed to parse YAML', got: %v", err)
		}
	})

	t.Run("returns error when stack-id missing with services", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "stack.yaml")
		content := `organization: test-org
project: test-project
services:
  web:
    servicePort: 8080
    image:
      type: internal
      name: myapp
      tag: latest
    resources:
      requests:
        memory: 256Mi
        cpu: 100m
      limits:
        memory: 512Mi
        cpu: 200m
    replicas: 1
`
		if err := os.WriteFile(configFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := LoadStackConfig(configFile)
		if err == nil {
			t.Fatal("LoadStackConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "stack-id is required") {
			t.Errorf("error should mention 'stack-id is required', got: %v", err)
		}
	})

	t.Run("validates service port", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "stack.yaml")
		content := `organization: test-org
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
      requests:
        memory: 256Mi
        cpu: 100m
      limits:
        memory: 512Mi
        cpu: 200m
    replicas: 1
`
		if err := os.WriteFile(configFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := LoadStackConfig(configFile)
		if err == nil {
			t.Fatal("LoadStackConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "servicePort must be greater than zero") {
			t.Errorf("error should mention 'servicePort must be greater than zero', got: %v", err)
		}
	})

	t.Run("validates image type", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "stack.yaml")
		content := `organization: test-org
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
      requests:
        memory: 256Mi
        cpu: 100m
      limits:
        memory: 512Mi
        cpu: 200m
    replicas: 1
`
		if err := os.WriteFile(configFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := LoadStackConfig(configFile)
		if err == nil {
			t.Fatal("LoadStackConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "must be 'internal' or 'external'") {
			t.Errorf("error should mention image type validation, got: %v", err)
		}
	})

	t.Run("validates replicas", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "stack.yaml")
		content := `organization: test-org
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
      requests:
        memory: 256Mi
        cpu: 100m
      limits:
        memory: 512Mi
        cpu: 200m
    replicas: 0
`
		if err := os.WriteFile(configFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := LoadStackConfig(configFile)
		if err == nil {
			t.Fatal("LoadStackConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "replicas must be at least 1") {
			t.Errorf("error should mention 'replicas must be at least 1', got: %v", err)
		}
	})

	t.Run("validates external image repository", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "stack.yaml")
		content := `organization: test-org
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
      requests:
        memory: 256Mi
        cpu: 100m
      limits:
        memory: 512Mi
        cpu: 200m
    replicas: 1
`
		if err := os.WriteFile(configFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := LoadStackConfig(configFile)
		if err == nil {
			t.Fatal("LoadStackConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "image.repository is required for external images") {
			t.Errorf("error should mention repository requirement, got: %v", err)
		}
	})
}

func TestServiceConfigToCreateRequest(t *testing.T) {
	t.Run("converts to create request", func(t *testing.T) {
		svc := ServiceConfig{
			Version:     "v1",
			ServicePort: 8080,
			Image: clients.ImageSpec{
				Type: "internal",
				Name: "myapp",
				Tag:  "latest",
			},
			Resources: clients.Resources{
				Requests: clients.ResourceRequirements{
					Memory: "256Mi",
					CPU:    "100m",
				},
				Limits: clients.ResourceRequirements{
					Memory: "512Mi",
					CPU:    "200m",
				},
			},
			Env: []clients.EnvVar{
				{Name: "KEY1", Value: "value1"},
			},
			SecretRefs: []clients.SecretRef{
				{SecretName: "my-secret"},
			},
			Endpoint: true,
			Replicas: 3,
		}

		stackId := "stack-123"
		req := svc.ToCreateRequest(stackId)

		if req.ServicePort != 8080 {
			t.Errorf("ServicePort = %d, want 8080", req.ServicePort)
		}
		if req.Image.Type != "internal" {
			t.Errorf("Image.Type = %q, want 'internal'", req.Image.Type)
		}
		if req.Replicas != 3 {
			t.Errorf("Replicas = %d, want 3", req.Replicas)
		}
		if req.StackId != stackId {
			t.Errorf("StackId = %q, want %q", req.StackId, stackId)
		}
		if !req.Endpoint {
			t.Error("Endpoint = false, want true")
		}
		if len(req.Env) != 1 {
			t.Errorf("Env length = %d, want 1", len(req.Env))
		}
		if len(req.SecretRefs) != 1 {
			t.Errorf("SecretRefs length = %d, want 1", len(req.SecretRefs))
		}
	})
}
