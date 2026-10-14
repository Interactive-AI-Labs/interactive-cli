package sync

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/google/go-cmp/cmp"
)

func TestPlanRemoteMcp(t *testing.T) {
	endpoint, authType := "https://example.com/mcp", "bearer"
	live := platform.McpSchema{EndpointURL: &endpoint, AuthType: &authType, HasCredential: true}
	tests := []struct {
		name        string
		config      files.McpConfig
		want        files.McpConfig
		wantChanged bool
		wantErr     string
	}{
		{
			name: "omitted bearer credential is preserved",
			config: files.McpConfig{
				Type:        "remote",
				EndpointURL: endpoint,
				Auth:        deployment.McpAuthBody{Type: "bearer"},
			},
			want: files.McpConfig{
				Type:        "remote",
				EndpointURL: endpoint,
				Auth:        deployment.McpAuthBody{Type: "bearer"},
			},
		},
		{
			name: "credential rotation is reported but redacted",
			config: files.McpConfig{
				Type:        "remote",
				EndpointURL: endpoint,
				Auth:        deployment.McpAuthBody{Type: "bearer", Credential: "do-not-print"},
			},
			want: files.McpConfig{
				Type:        "remote",
				EndpointURL: endpoint,
				Auth:        deployment.McpAuthBody{Type: "bearer"},
			},
			wantChanged: true,
		},
		{
			name: "endpoint change preserves omitted credential",
			config: files.McpConfig{
				Type:        "remote",
				EndpointURL: "https://new.example.com/mcp",
				Auth:        deployment.McpAuthBody{Type: "bearer"},
			},
			want: files.McpConfig{
				Type:        "remote",
				EndpointURL: "https://new.example.com/mcp",
				Auth:        deployment.McpAuthBody{Type: "bearer"},
			},
			wantChanged: true,
		},
		{
			name: "invalid auth change propagates validation error",
			config: files.McpConfig{
				Type:        "remote",
				EndpointURL: endpoint,
				Auth:        deployment.McpAuthBody{Type: "api_key"},
			},
			want: files.McpConfig{
				Type:        "remote",
				EndpointURL: endpoint,
				Auth:        deployment.McpAuthBody{Type: "api_key"},
			},
			wantErr: "auth.credential is required to update mcp \"tools\"; stack get never exports credentials",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed, err := planRemoteMcp("tools", live, tt.config, "stack")
			if (err != nil && err.Error() != tt.wantErr) || (err == nil && tt.wantErr != "") {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("config mismatch (-want +got):\n%s", diff)
			}
			if changed != tt.wantChanged {
				t.Fatalf("changed = %v, want %v", changed, tt.wantChanged)
			}
		})
	}
}

func TestPendingMcpNames(t *testing.T) {
	tests := []struct {
		name        string
		plan        files.ResourcePlan[files.McpConfig]
		live        map[string]files.McpConfig
		allowDelete bool
		want        []string
	}{
		{name: "empty"},
		{
			name: "unchanged dependency",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{"tools": {}},
			},
			live: map[string]files.McpConfig{"tools": {}},
		},
		{
			name: "created changed and deleted dependencies are sorted",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{"new": {}, "changed": {}},
				Changed: map[string]bool{"changed": true},
			},
			live:        map[string]files.McpConfig{"changed": {}, "deleted": {}},
			allowDelete: true,
			want:        []string{"changed", "deleted", "new"},
		},
		{
			name: "remote credential-only update does not defer an agent",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{
					"tools": {
						Type:        "remote",
						EndpointURL: "https://example.com/mcp",
						Auth:        deployment.McpAuthBody{Type: "bearer"},
					},
				},
				Changed: map[string]bool{"tools": true},
			},
			live: map[string]files.McpConfig{
				"tools": {
					Type:        "remote",
					EndpointURL: "https://example.com/mcp",
					Auth:        deployment.McpAuthBody{Type: "bearer"},
				},
			},
		},
		{
			name: "catalog remote credential-only update does not defer an agent",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{
					"tools": {
						Type:      "remote",
						CatalogID: "catalog-tools",
						Auth:      deployment.McpAuthBody{Type: "api_key"},
					},
				},
				Changed: map[string]bool{"tools": true},
			},
			live: map[string]files.McpConfig{
				"tools": {
					Type:      "remote",
					CatalogID: "catalog-tools",
					Auth:      deployment.McpAuthBody{Type: "api_key"},
				},
			},
		},
		{
			name: "remote endpoint update still defers an agent",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{
					"tools": {
						Type:        "remote",
						EndpointURL: "https://new.example.com/mcp",
						Auth:        deployment.McpAuthBody{Type: "bearer"},
					},
				},
				Changed: map[string]bool{"tools": true},
			},
			live: map[string]files.McpConfig{
				"tools": {
					Type:        "remote",
					EndpointURL: "https://example.com/mcp",
					Auth:        deployment.McpAuthBody{Type: "bearer"},
				},
			},
			want: []string{"tools"},
		},
		{
			name: "remote auth update still defers an agent",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{
					"tools": {
						Type:        "remote",
						EndpointURL: "https://example.com/mcp",
						Auth:        deployment.McpAuthBody{Type: "api_key"},
					},
				},
				Changed: map[string]bool{"tools": true},
			},
			live: map[string]files.McpConfig{
				"tools": {
					Type:        "remote",
					EndpointURL: "https://example.com/mcp",
					Auth:        deployment.McpAuthBody{Type: "bearer"},
				},
			},
			want: []string{"tools"},
		},
		{
			name: "remote custom header update still defers an agent",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{
					"tools": {
						Type: "remote",
						Auth: deployment.McpAuthBody{Type: "custom", Header: "X-New-Key"},
					},
				},
				Changed: map[string]bool{"tools": true},
			},
			live: map[string]files.McpConfig{
				"tools": {
					Type: "remote",
					Auth: deployment.McpAuthBody{Type: "custom", Header: "X-Key"},
				},
			},
			want: []string{"tools"},
		},
		{
			name: "hosted update still defers an agent when its visible config is unchanged",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{"tools": {}},
				Changed: map[string]bool{"tools": true},
			},
			live: map[string]files.McpConfig{"tools": {}},
			want: []string{"tools"},
		},
		{
			name: "new remote dependency still defers an agent",
			plan: files.ResourcePlan[files.McpConfig]{
				Desired: map[string]files.McpConfig{"tools": {Type: "remote"}},
			},
			want: []string{"tools"},
		},
		{
			name:        "permitted remote deletion still defers an agent",
			live:        map[string]files.McpConfig{"tools": {Type: "remote"}},
			allowDelete: true,
			want:        []string{"tools"},
		},
		{
			name: "protected deletion does not defer an agent",
			live: map[string]files.McpConfig{"protected": {}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pendingMcpNames(tt.plan, tt.live, tt.allowDelete)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("pending mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPlanResources(t *testing.T) {
	normalize := func(_ string, cfg string) (string, bool, error) {
		n, err := strconv.Atoi(cfg)
		if err != nil {
			return "", false, err
		}
		desired := strconv.Itoa(n)
		return desired, desired != cfg, nil
	}
	tests := []struct {
		name        string
		local, live map[string]string
		want        files.ResourcePlan[string]
		wantErr     string
	}{
		{
			name: "omitted live resources are not planned",
			live: map[string]string{"removed": "bad"},
			want: files.ResourcePlan[string]{
				Desired: map[string]string{},
				Changed: map[string]bool{},
			},
		},
		{
			name:  "creates bypass planning even when the planner would reject their input",
			local: map[string]string{"new": "bad"},
			want: files.ResourcePlan[string]{
				Desired: map[string]string{"new": "bad"},
				Changed: map[string]bool{},
			},
		},
		{
			name:  "only existing resources receive predictions",
			local: map[string]string{"existing": "01", "new": "02"},
			live:  map[string]string{"existing": "1"},
			want: files.ResourcePlan[string]{
				Desired: map[string]string{"existing": "1", "new": "02"},
				Changed: map[string]bool{"existing": true},
			},
		},
		{
			name:  "uses the planner flag rather than comparing with the live snapshot",
			local: map[string]string{"existing": "1"},
			live:  map[string]string{"existing": "2"},
			want: files.ResourcePlan[string]{
				Desired: map[string]string{"existing": "1"},
				Changed: map[string]bool{},
			},
		},
		{
			name:  "sorted planning stops at the first error and retains earlier results",
			local: map[string]string{"charlie": "03", "bravo": "bad", "alpha": "01"},
			live:  map[string]string{"alpha": "0", "bravo": "0"},
			want: files.ResourcePlan[string]{
				Desired: map[string]string{"alpha": "1"},
				Changed: map[string]bool{"alpha": true},
			},
			wantErr: "strconv.Atoi: parsing \"bad\": invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := planResources(tt.local, tt.live, normalize)
			if (err != nil && err.Error() != tt.wantErr) || (err == nil && tt.wantErr != "") {
				t.Fatalf("error = %v, want %q", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("plan mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPrintStackPlan(t *testing.T) {
	tests := []struct {
		name        string
		plan        files.StackDiff
		allowDelete []string
		want        string
	}{
		{
			name: "one summary for an empty or unchanged stack",
			want: "No changes required.\n",
		},
		{
			name: "services-only plan omits unrelated sections",
			plan: files.StackDiff{Services: files.ResourceTypeDiff{Created: []string{"api"}}},
			want: "Would create services: api\n",
		},
		{
			name: "deferred-only plan is not unchanged",
			plan: files.StackDiff{
				Agents: files.ResourceTypeDiff{Deferred: map[string][]string{"helper": {"tools"}}},
			},
			want: "Plan deferred for agents helper: changing mcps: tools; check again after those changes.\n",
		},
		{
			name: "protected deletion",
			plan: files.StackDiff{Mcps: files.ResourceTypeDiff{Deleted: []string{"tools"}}},
			want: "Would refuse to delete mcps: tools (a config that omits a resource looks identical to a stale one — pass --allow-delete=mcps to delete)\n",
		},
		{
			name: "per-kind deletion and sorted deferred agents alongside known operations",
			plan: files.StackDiff{
				Mcps: files.ResourceTypeDiff{
					Created: []string{"new"},
					Updated: []files.ResourceChange{{Name: "changed"}},
					Deleted: []string{"old"},
				},
				Agents: files.ResourceTypeDiff{
					Deferred: map[string][]string{"zulu": {"new"}, "alpha": {"changed", "old"}},
				},
			},
			allowDelete: []string{"mcps"},
			want: "Would create mcps: new\nWould update mcps: changed\nWould delete mcps: old\n" +
				"Plan deferred for agents alpha: changing mcps: changed, old; check again after those changes.\n" +
				"Plan deferred for agents zulu: changing mcps: new; check again after those changes.\n",
		},
		{
			name:        "allow all deletions",
			plan:        files.StackDiff{Mcps: files.ResourceTypeDiff{Deleted: []string{"tools"}}},
			allowDelete: []string{"all"},
			want:        "Would delete mcps: tools\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			PrintStackPlan(&out, &tt.plan, tt.allowDelete)
			if got := out.String(); got != tt.want {
				t.Fatalf("output = %q, want %q", got, tt.want)
			}
		})
	}
}
