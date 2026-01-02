package internal

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNewAPIClient(t *testing.T) {
	t.Run("creates client with cookies", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.isApiKeyMode {
			t.Error("expected isApiKeyMode to be false")
		}
	})

	t.Run("returns error with no auth", func(t *testing.T) {
		_, err := NewAPIClient("https://api.example.com", 30*time.Second, "", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no authentication method available") {
			t.Errorf("error should mention 'no authentication method available', got: %v", err)
		}
	})
}

func TestAPIClientGetOrgIdByName(t *testing.T) {
	t.Run("returns error for empty name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetOrgIdByName(ctx, "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "organization name cannot be empty") {
			t.Errorf("error should mention 'organization name cannot be empty', got: %v", err)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetOrgIdByName(ctx, "   ")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "organization name cannot be empty") {
			t.Errorf("error should mention 'organization name cannot be empty', got: %v", err)
		}
	})
}

func TestAPIClientGetProjectByName(t *testing.T) {
	t.Run("returns error for empty name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetProjectByName(ctx, "org-123", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "project name cannot be empty") {
			t.Errorf("error should mention 'project name cannot be empty', got: %v", err)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetProjectByName(ctx, "org-123", "   ")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "project name cannot be empty") {
			t.Errorf("error should mention 'project name cannot be empty', got: %v", err)
		}
	})
}

func TestAPIClientGetProjectId(t *testing.T) {
	t.Run("returns error for empty org name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, _, err = client.GetProjectId(ctx, "", "project")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "organization name cannot be empty") {
			t.Errorf("error should mention 'organization name cannot be empty', got: %v", err)
		}
	})

	t.Run("returns error for empty project name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, _, err = client.GetProjectId(ctx, "org", "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "project name cannot be empty") {
			t.Errorf("error should mention 'project name cannot be empty', got: %v", err)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		client, err := NewAPIClient("https://api.example.com", 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, _, err = client.GetProjectId(ctx, "  ", "  ")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
