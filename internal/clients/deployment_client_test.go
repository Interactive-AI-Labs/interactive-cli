package internal

import (
	"net/http"
	"testing"
	"time"
)

func TestNewDeploymentClient(t *testing.T) {
	t.Run("creates client with API key", func(t *testing.T) {
		client, err := NewDeploymentClient("https://deploy.example.com", 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.apiKey != "test-key" {
			t.Errorf("apiKey = %q, want %q", client.apiKey, "test-key")
		}
	})

	t.Run("creates client with cookies", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewDeploymentClient("https://deploy.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if len(client.cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(client.cookies))
		}
	})

	t.Run("returns error with no auth", func(t *testing.T) {
		_, err := NewDeploymentClient("https://deploy.example.com", 30*time.Second, "", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("stores hostname correctly", func(t *testing.T) {
		hostname := "https://deploy.example.com"
		client, err := NewDeploymentClient(hostname, 30*time.Second, "test-key", nil)
		if err != nil {
			t.Fatalf("NewDeploymentClient() error = %v", err)
		}
		if client.hostname != hostname {
			t.Errorf("hostname = %q, want %q", client.hostname, hostname)
		}
	})
}
