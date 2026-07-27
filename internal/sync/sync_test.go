package sync

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestAllowDeleteResource(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		resource string
		want     bool
	}{
		{
			name:     "nil list",
			allowed:  nil,
			resource: "databases",
			want:     false,
		},
		{
			name:     "empty list",
			allowed:  []string{},
			resource: "databases",
			want:     false,
		},
		{
			name:     "exact match",
			allowed:  []string{"databases"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "case insensitive match",
			allowed:  []string{"Databases"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "all keyword",
			allowed:  []string{"all"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "ALL keyword uppercase",
			allowed:  []string{"ALL"},
			resource: "databases",
			want:     true,
		},
		{
			name:     "no match",
			allowed:  []string{"services"},
			resource: "databases",
			want:     false,
		},
		{
			name:     "multiple entries with match",
			allowed:  []string{"services", "databases"},
			resource: "databases",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllowDeleteResource(tt.allowed, tt.resource)
			if got != tt.want {
				t.Errorf(
					"AllowDeleteResource(%v, %q) = %v, want %v",
					tt.allowed, tt.resource, got, tt.want,
				)
			}
		})
	}
}

func TestPrintResult(t *testing.T) {
	tests := []struct {
		name    string
		label   string
		result  *Result
		syncErr error
		want    string
		wantErr bool
	}{
		{
			name:  "created and deleted",
			label: "databases",
			result: &Result{
				Created: []string{"new-db"},
				Deleted: []string{"old-db"},
			},
			want: "Created databases: new-db\n" +
				"Deleted databases: old-db\n",
		},
		{
			name:  "updated items",
			label: "services",
			result: &Result{
				Updated: []string{"svc-a"},
			},
			want: "Updated services: svc-a\n",
		},
		{
			name:   "no changes",
			label:  "services",
			result: &Result{},
			want:   "No changes required; services already match config.\n",
		},
		{
			name:  "multiple items joined with comma",
			label: "services",
			result: &Result{
				Created: []string{"svc-a", "svc-b"},
				Deleted: []string{"svc-c", "svc-d"},
			},
			want: "Created services: svc-a, svc-b\n" +
				"Deleted services: svc-c, svc-d\n",
		},
		{
			name:  "protected items print warning",
			label: "databases",
			result: &Result{
				Created:   []string{"new-db"},
				Protected: []string{"old-db"},
			},
			want: "Created databases: new-db\n" +
				"\nProtected databases (not deleted): old-db\n" +
				"Use --allow-delete=databases to delete them.\n",
		},
		{
			name:    "error with partial result",
			label:   "services",
			result:  &Result{Created: []string{"svc-a"}},
			syncErr: fmt.Errorf("failed to create service \"svc-b\""),
			wantErr: true,
			want:    "Created services (partial): svc-a\n",
		},
		{
			name:    "error with nil result",
			label:   "services",
			syncErr: fmt.Errorf("failed to list services"),
			wantErr: true,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintResult(&buf, tt.label, tt.result, tt.syncErr)
			if tt.wantErr && err != tt.syncErr {
				t.Fatalf("expected original error, got: %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

// newTestDeployClient returns a client pointed at a stub deployment server.
func newTestDeployClient(t *testing.T, handler http.HandlerFunc) *clients.DeploymentClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := clients.NewDeploymentClient(server.URL, 5*time.Second, "test-token", "", nil)
	if err != nil {
		t.Fatalf("NewDeploymentClient() error = %v", err)
	}
	return client
}

func TestServicesPrintsUpdateBanner(t *testing.T) {
	client := newTestDeployClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/organizations/o1/projects/p1/services":
			fmt.Fprint(
				w,
				`{"services":[{"name":"svc-a","projectId":"p1","revision":3,"status":"ready","updated":"2026-07-24T11:20:00Z"}]}`,
			)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/organizations/o1/projects/p1/services/svc-a":
			fmt.Fprint(w, `{}`)
		case r.Method == http.MethodPost && r.URL.Path == "/v1/organizations/o1/projects/p1/services/svc-new":
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	var warn bytes.Buffer
	desired := map[string]clients.CreateServiceBody{
		"svc-a":   {},
		"svc-new": {},
	}
	result, err := Services(context.Background(), &warn, client, "o1", "p1", "stack-1", desired)
	if err != nil {
		t.Fatalf("Services() error = %v", err)
	}

	wantWarn := "Live: service svc-a revision 3, last updated 2026-07-24 11:20 UTC — this update creates revision 4\n"
	if got := warn.String(); got != wantWarn {
		t.Errorf("banner = %q, want %q", got, wantWarn)
	}
	if len(result.Updated) != 1 || result.Updated[0] != "svc-a" {
		t.Errorf("Updated = %v, want [svc-a]", result.Updated)
	}
	if len(result.Created) != 1 || result.Created[0] != "svc-new" {
		t.Errorf("Created = %v, want [svc-new]", result.Created)
	}
}

func TestAgentsPrintsUpdateBanner(t *testing.T) {
	client := newTestDeployClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/organizations/o1/projects/p1/agents":
			fmt.Fprint(
				w,
				`{"agents":[{"name":"agent-a","projectId":"p1","revision":13,"status":"ready","updated":"2026-07-24T11:20:00Z"}]}`,
			)
		case r.Method == http.MethodPut && r.URL.Path == "/v1/organizations/o1/projects/p1/agents/agent-a":
			fmt.Fprint(w, `{}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	})

	var warn bytes.Buffer
	desired := map[string]clients.CreateAgentBody{"agent-a": {}}
	result, err := Agents(context.Background(), &warn, client, "o1", "p1", "stack-1", desired)
	if err != nil {
		t.Fatalf("Agents() error = %v", err)
	}

	wantWarn := "Live: agent agent-a revision 13, last updated 2026-07-24 11:20 UTC — this update creates revision 14\n"
	if got := warn.String(); got != wantWarn {
		t.Errorf("banner = %q, want %q", got, wantWarn)
	}
	if len(result.Updated) != 1 || result.Updated[0] != "agent-a" {
		t.Errorf("Updated = %v, want [agent-a]", result.Updated)
	}
}
