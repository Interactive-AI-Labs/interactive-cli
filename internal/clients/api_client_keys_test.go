package clients

import (
	"net/http"
	"testing"
)

func TestProjectAPIKeysPath(t *testing.T) {
	tests := []struct {
		name      string
		orgID     string
		projectID string
		want      string
	}{
		{
			name:      "plain ids",
			orgID:     "org-1",
			projectID: "project-1",
			want:      "/api/platform/v1/organizations/org-1/projects/project-1/api-keys",
		},
		{
			name:      "escapes path segments",
			orgID:     "org/1",
			projectID: "project 1",
			want:      "/api/platform/v1/organizations/org%2F1/projects/project%201/api-keys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := projectAPIKeysPath(tt.orgID, tt.projectID); got != tt.want {
				t.Fatalf("projectAPIKeysPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProjectAPIKeyPath(t *testing.T) {
	tests := []struct {
		name      string
		orgID     string
		projectID string
		keyID     string
		want      string
	}{
		{
			name:      "plain ids",
			orgID:     "org-1",
			projectID: "project-1",
			keyID:     "key-1",
			want:      "/api/platform/v1/organizations/org-1/projects/project-1/api-keys/key-1",
		},
		{
			name:      "escapes path segments",
			orgID:     "org/1",
			projectID: "project 1",
			keyID:     "key/1",
			want:      "/api/platform/v1/organizations/org%2F1/projects/project%201/api-keys/key%2F1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := projectAPIKeyPath(tt.orgID, tt.projectID, tt.keyID); got != tt.want {
				t.Fatalf("projectAPIKeyPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRouterAPIKeyPath(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		keyID     string
		want      string
	}{
		{
			name:      "plain ids",
			projectID: "project-1",
			keyID:     "key-1",
			want:      "/api/v1/projects/project-1/openrouter-keys/key-1",
		},
		{
			name:      "escapes path segments",
			projectID: "project 1",
			keyID:     "key/1",
			want:      "/api/v1/projects/project%201/openrouter-keys/key%2F1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := routerAPIKeyPath(tt.projectID, tt.keyID); got != tt.want {
				t.Fatalf("routerAPIKeyPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequireKeyManagementAuth(t *testing.T) {
	tests := []struct {
		name    string
		client  *APIClient
		wantErr bool
	}{
		{name: "rejects api key", client: &APIClient{apiKey: "pk:sk"}, wantErr: true},
		{name: "allows jwt", client: &APIClient{token: "jwt"}},
		{name: "allows cookies", client: &APIClient{cookies: nonEmptyCookies()}},
		{name: "rejects missing auth", client: &APIClient{}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.requireKeyManagementAuth()
			if tt.wantErr && err == nil {
				t.Fatal("requireKeyManagementAuth() expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("requireKeyManagementAuth() error = %v", err)
			}
		})
	}
}

func nonEmptyCookies() []*http.Cookie {
	return []*http.Cookie{{Name: "next-auth.session-token", Value: "cookie"}}
}
