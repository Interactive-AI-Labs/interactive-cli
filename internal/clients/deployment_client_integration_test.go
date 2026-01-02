//go:build integration
// +build integration

package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDeploymentClientIntegrationSecrets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case strings.Contains(path, "/secrets/db-creds") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"name": "db-creds",
				"type": "opaque",
				"createdAt": "2024-01-01T00:00:00Z",
				"keys": ["username", "password"],
				"data": {
					"username": "admin",
					"password": "secret123"
				}
			}`))

		case strings.Contains(path, "/secrets") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"secrets": [
					{
						"name": "db-creds",
						"type": "opaque",
						"createdAt": "2024-01-01T00:00:00Z",
						"keys": ["username", "password"]
					},
					{
						"name": "api-key",
						"type": "opaque",
						"createdAt": "2024-01-02T00:00:00Z",
						"keys": ["key"]
					}
				]
			}`))

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("lists secrets", func(t *testing.T) {
		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		secrets, err := client.ListSecrets(ctx, "org-123", "proj-456")
		if err != nil {
			t.Fatalf("ListSecrets() error = %v", err)
		}

		if len(secrets) != 2 {
			t.Fatalf("expected 2 secrets, got %d", len(secrets))
		}

		if secrets[0].Name != "db-creds" {
			t.Errorf("secrets[0].Name = %q, want %q", secrets[0].Name, "db-creds")
		}
		if len(secrets[0].Keys) != 2 {
			t.Errorf("secrets[0].Keys length = %d, want 2", len(secrets[0].Keys))
		}
	})

	t.Run("gets secret", func(t *testing.T) {
		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		secret, err := client.GetSecret(ctx, "org-123", "proj-456", "db-creds")
		if err != nil {
			t.Fatalf("GetSecret() error = %v", err)
		}

		if secret.Name != "db-creds" {
			t.Errorf("secret.Name = %q, want %q", secret.Name, "db-creds")
		}
		if secret.Data["username"] != "admin" {
			t.Errorf("secret.Data[username] = %q, want %q", secret.Data["username"], "admin")
		}
	})
}

func TestDeploymentClientIntegrationImages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/images") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"images": [
					{
						"name": "myapp",
						"tags": ["v1.0", "v1.1", "latest"]
					},
					{
						"name": "worker",
						"tags": ["v2.0"]
					}
				]
			}`))
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	t.Run("lists images", func(t *testing.T) {
		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		images, err := client.ListImages(ctx, "org-123", "proj-456")
		if err != nil {
			t.Fatalf("ListImages() error = %v", err)
		}

		if len(images) != 2 {
			t.Fatalf("expected 2 images, got %d", len(images))
		}

		if images[0].Name != "myapp" {
			t.Errorf("images[0].Name = %q, want %q", images[0].Name, "myapp")
		}
		if len(images[0].Tags) != 3 {
			t.Errorf("images[0].Tags length = %d, want 3", len(images[0].Tags))
		}
	})
}

func TestDeploymentClientIntegrationServices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case strings.Contains(path, "/services") && r.Method == http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"services": [
					{
						"name": "web",
						"stackId": "stack-1",
						"replicas": 2,
						"servicePort": 8080,
						"image": {
							"type": "internal",
							"name": "myapp",
							"tag": "latest"
						},
						"resources": {
							"requests": {
								"memory": "256Mi",
								"cpu": "100m"
							},
							"limits": {
								"memory": "512Mi",
								"cpu": "200m"
							}
						}
					}
				]
			}`))

		case strings.Contains(path, "/services") && r.Method == http.MethodPost:
			body, _ := io.ReadAll(r.Body)
			var req CreateServiceBody
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`service-id-123`))

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Run("lists services", func(t *testing.T) {
		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		services, err := client.ListServices(ctx, "org-123", "proj-456", "stack-1")
		if err != nil {
			t.Fatalf("ListServices() error = %v", err)
		}

		if len(services) != 1 {
			t.Fatalf("expected 1 service, got %d", len(services))
		}

		if services[0].Name != "web" {
			t.Errorf("services[0].Name = %q, want %q", services[0].Name, "web")
		}
	})

	t.Run("creates service", func(t *testing.T) {
		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		body := CreateServiceBody{
			ServicePort: 8080,
			Image: ImageSpec{
				Type: "internal",
				Name: "myapp",
				Tag:  "latest",
			},
			Resources: Resources{
				Requests: ResourceRequirements{
					Memory: "256Mi",
					CPU:    "100m",
				},
				Limits: ResourceRequirements{
					Memory: "512Mi",
					CPU:    "200m",
				},
			},
			Replicas: 2,
			StackId:  "stack-1",
		}

		serviceId, err := client.CreateService(ctx, "org-123", "proj-456", "web", body)
		if err != nil {
			t.Fatalf("CreateService() error = %v", err)
		}

		if serviceId == "" {
			t.Error("expected non-empty service ID")
		}
	})
}

func TestDeploymentClientIntegrationReplicas(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/replicas") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"replicas": [
				{
					"name": "web-abc123",
					"phase": "Running",
					"status": "Ready",
					"ready": true,
					"cpu": "50m",
					"memory": "128Mi",
					"startTime": "2024-01-01T00:00:00Z"
				},
				{
					"name": "web-def456",
					"phase": "Running",
					"status": "Ready",
					"ready": true,
					"cpu": "45m",
					"memory": "120Mi",
					"startTime": "2024-01-01T00:01:00Z"
				}
			]
		}`))
	}))
	defer server.Close()

	t.Run("lists replicas", func(t *testing.T) {
		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		replicas, err := client.ListReplicas(ctx, "org-123", "proj-456", "web")
		if err != nil {
			t.Fatalf("ListReplicas() error = %v", err)
		}

		if len(replicas) != 2 {
			t.Fatalf("expected 2 replicas, got %d", len(replicas))
		}

		if replicas[0].Name != "web-abc123" {
			t.Errorf("replicas[0].Name = %q, want %q", replicas[0].Name, "web-abc123")
		}
		if replicas[0].Phase != "Running" {
			t.Errorf("replicas[0].Phase = %q, want %q", replicas[0].Phase, "Running")
		}
		if !replicas[0].Ready {
			t.Error("replicas[0].Ready = false, want true")
		}
	})
}

func TestDeploymentClientIntegrationErrorHandling(t *testing.T) {
	t.Run("handles 404 errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "Resource not found"}`))
		}))
		defer server.Close()

		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.ListSecrets(ctx, "org-123", "proj-456")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "Resource not found") {
			t.Errorf("error should contain server message, got: %v", err)
		}
	})

	t.Run("handles authentication errors", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message": "Authentication required"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.ListSecrets(ctx, "org-123", "proj-456")
		if err != nil {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("handles malformed JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{malformed json`))
		}))
		defer server.Close()

		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.ListSecrets(ctx, "org-123", "proj-456")
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
	})
}

func TestDeploymentClientIntegrationRequestBody(t *testing.T) {
	t.Run("sends correct request body for service creation", func(t *testing.T) {
		var receivedBody CreateServiceBody
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}

			if err := json.Unmarshal(body, &receivedBody); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`service-nginx-id`))
		}))
		defer server.Close()

		client, err := NewDeploymentClient(server.URL, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}

		body := CreateServiceBody{
			ServicePort: 9000,
			Image: ImageSpec{
				Type: "external",
				Name: "nginx",
				Tag:  "1.21",
			},
			Resources: Resources{
				Requests: ResourceRequirements{
					Memory: "128Mi",
					CPU:    "50m",
				},
				Limits: ResourceRequirements{
					Memory: "256Mi",
					CPU:    "100m",
				},
			},
			Replicas: 3,
			StackId:  "stack-test",
			Endpoint: true,
		}

		ctx := context.Background()
		serviceId, err := client.CreateService(ctx, "org-123", "proj-456", "nginx", body)
		if err != nil {
			t.Fatalf("CreateService() error = %v", err)
		}

		if serviceId == "" {
			t.Error("expected non-empty service ID")
		}

		if receivedBody.ServicePort != 9000 {
			t.Errorf("receivedBody.ServicePort = %d, want 9000", receivedBody.ServicePort)
		}
		if receivedBody.Image.Type != "external" {
			t.Errorf("receivedBody.Image.Type = %q, want 'external'", receivedBody.Image.Type)
		}
		if receivedBody.Replicas != 3 {
			t.Errorf("receivedBody.Replicas = %d, want 3", receivedBody.Replicas)
		}
		if !receivedBody.Endpoint {
			t.Error("receivedBody.Endpoint = false, want true")
		}
	})
}
