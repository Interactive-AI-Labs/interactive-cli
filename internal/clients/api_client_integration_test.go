//go:build integration
// +build integration

package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAPIClientIntegrationListOrganizations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/session/organizations" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		cookies := r.Cookies()
		if len(cookies) == 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"organizations": [
				{"id": "org-1", "name": "Org One", "project_count": 5, "role": "admin"},
				{"id": "org-2", "name": "Org Two", "project_count": 3, "role": "member"}
			]
		}`))
	}))
	defer server.Close()

	t.Run("successfully lists organizations with cookies", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		orgs, err := client.ListOrganizations(ctx)
		if err != nil {
			t.Fatalf("ListOrganizations() error = %v", err)
		}

		if len(orgs) != 2 {
			t.Fatalf("expected 2 organizations, got %d", len(orgs))
		}

		if orgs[0].Id != "org-1" {
			t.Errorf("orgs[0].Id = %q, want %q", orgs[0].Id, "org-1")
		}
		if orgs[0].Name != "Org One" {
			t.Errorf("orgs[0].Name = %q, want %q", orgs[0].Name, "Org One")
		}
		if orgs[0].ProjectCount != 5 {
			t.Errorf("orgs[0].ProjectCount = %d, want %d", orgs[0].ProjectCount, 5)
		}
	})
}

func TestAPIClientIntegrationListProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/session/organizations/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		cookies := r.Cookies()
		if len(cookies) == 0 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"projects": [
				{"id": "proj-1", "name": "Project One", "role": "admin"},
				{"id": "proj-2", "name": "Project Two", "role": "member"}
			]
		}`))
	}))
	defer server.Close()

	t.Run("successfully lists projects with cookies", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		projects, err := client.ListProjects(ctx, "org-123")
		if err != nil {
			t.Fatalf("ListProjects() error = %v", err)
		}

		if len(projects) != 2 {
			t.Fatalf("expected 2 projects, got %d", len(projects))
		}

		if projects[0].Id != "proj-1" {
			t.Errorf("projects[0].Id = %q, want %q", projects[0].Id, "proj-1")
		}
		if projects[0].Name != "Project One" {
			t.Errorf("projects[0].Name = %q, want %q", projects[0].Name, "Project One")
		}
	})
}

func TestAPIClientIntegrationGetOrgIdByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/session/organizations" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"organizations": [
				{"id": "org-1", "name": "Test Org", "project_count": 5, "role": "admin"},
				{"id": "org-2", "name": "Another Org", "project_count": 3, "role": "member"}
			]
		}`))
	}))
	defer server.Close()

	t.Run("finds organization by name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		orgId, err := client.GetOrgIdByName(ctx, "Test Org")
		if err != nil {
			t.Fatalf("GetOrgIdByName() error = %v", err)
		}

		if orgId != "org-1" {
			t.Errorf("orgId = %q, want %q", orgId, "org-1")
		}
	})

	t.Run("case insensitive search", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		orgId, err := client.GetOrgIdByName(ctx, "test org")
		if err != nil {
			t.Fatalf("GetOrgIdByName() error = %v", err)
		}

		if orgId != "org-1" {
			t.Errorf("orgId = %q, want %q", orgId, "org-1")
		}
	})

	t.Run("returns error for not found", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetOrgIdByName(ctx, "Nonexistent Org")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error should mention 'not found', got: %v", err)
		}
	})
}

func TestAPIClientIntegrationGetProjectByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/session/organizations/") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"projects": [
				{"id": "proj-1", "name": "Test Project", "role": "admin"},
				{"id": "proj-2", "name": "Another Project", "role": "member"}
			]
		}`))
	}))
	defer server.Close()

	t.Run("finds project by name", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		projectId, err := client.GetProjectByName(ctx, "org-123", "Test Project")
		if err != nil {
			t.Fatalf("GetProjectByName() error = %v", err)
		}

		if projectId != "proj-1" {
			t.Errorf("projectId = %q, want %q", projectId, "proj-1")
		}
	})

	t.Run("case insensitive search", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		projectId, err := client.GetProjectByName(ctx, "org-123", "test project")
		if err != nil {
			t.Fatalf("GetProjectByName() error = %v", err)
		}

		if projectId != "proj-1" {
			t.Errorf("projectId = %q, want %q", projectId, "proj-1")
		}
	})

	t.Run("returns error for not found", func(t *testing.T) {
		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.GetProjectByName(ctx, "org-123", "Nonexistent Project")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error should mention 'not found', got: %v", err)
		}
	})
}

func TestAPIClientIntegrationServerErrors(t *testing.T) {
	t.Run("handles server error with message field", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Internal server error occurred"}`))
		}))
		defer server.Close()

		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.ListOrganizations(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "Internal server error occurred") {
			t.Errorf("error should contain server message, got: %v", err)
		}
	})

	t.Run("handles server error with detail field", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"detail": "Invalid request parameters"}`))
		}))
		defer server.Close()

		cookies := []*http.Cookie{{Name: "session", Value: "test-session"}}
		client, err := NewAPIClient(server.URL, 30*time.Second, "", cookies)
		if err != nil {
			t.Fatalf("NewAPIClient() error = %v", err)
		}

		ctx := context.Background()
		_, err = client.ListOrganizations(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "Invalid request parameters") {
			t.Errorf("error should contain server message, got: %v", err)
		}
	})
}
