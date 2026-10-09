package files

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestRemoteAndDeferredStackPlans(t *testing.T) {
	liveMcp := McpConfig{
		Type:        "remote",
		EndpointURL: "https://old.example/mcp",
		Auth:        deployment.McpAuthBody{Type: "bearer"},
	}
	rotated := liveMcp
	rotated.Auth.Credential = "secret-must-not-be-printed"
	moved := rotated
	moved.EndpointURL = "https://new.example/mcp"
	agent := AgentConfig{AgentConfig: map[string]any{"mcps": []any{"tools"}}}
	const unchanged = `{"created":null,"updated":null,"deleted":null}`
	const createdMcp = `{"created":["tools"],"updated":null,"deleted":null}`
	const movedMcp = `{"created":null,"updated":[{"name":"tools","changes":{"endpointUrl":{"old":"https://old.example/mcp","new":"https://new.example/mcp"}}}],"deleted":null}`
	const deferredAgent = `{"created":null,"updated":null,"deleted":null,"deferred":{"helper":["tools"]}}`

	// These paths use local data only: remote comparison, creates, and deferred agents.
	tests := []struct {
		name                              string
		local, live                       StackConfig
		wantMcps, wantAgents, wantPrinted string
	}{
		{
			name:     "no resources",
			wantMcps: unchanged, wantAgents: unchanged,
			wantPrinted: "No differences found.\n",
		},
		{
			name:     "remote credential alone is not displayed or compared",
			local:    StackConfig{Mcps: map[string]McpConfig{"tools": rotated}},
			live:     StackConfig{Mcps: map[string]McpConfig{"tools": liveMcp}},
			wantMcps: unchanged, wantAgents: unchanged,
			wantPrinted: "No differences found.\n",
		},
		{
			name:        "remote endpoint change redacts credential in text and JSON",
			local:       StackConfig{Mcps: map[string]McpConfig{"tools": moved}},
			live:        StackConfig{Mcps: map[string]McpConfig{"tools": liveMcp}},
			wantMcps:    movedMcp,
			wantAgents:  unchanged,
			wantPrinted: "Stack: s\n\n  ~ mcp tools (update)\n    endpointUrl: https://old.example/mcp → https://new.example/mcp\n",
		},
		{
			name:     "created remote credential is not displayed",
			local:    StackConfig{Mcps: map[string]McpConfig{"tools": rotated}},
			wantMcps: createdMcp, wantAgents: unchanged,
			wantPrinted: "Stack: s\n\n  + mcp tools (create)\n",
		},
		{
			name: "existing agent waits for a new MCP",
			local: StackConfig{
				Mcps:   map[string]McpConfig{"tools": rotated},
				Agents: map[string]AgentConfig{"helper": agent},
			},
			live:        StackConfig{Agents: map[string]AgentConfig{"helper": agent}},
			wantMcps:    createdMcp,
			wantAgents:  deferredAgent,
			wantPrinted: "Stack: s\n\n  ? agent helper (plan deferred; changing mcps: tools)\n\n  + mcp tools (create)\n",
		},
		{
			name: "existing agent waits for a changed MCP",
			local: StackConfig{
				Mcps:   map[string]McpConfig{"tools": moved},
				Agents: map[string]AgentConfig{"helper": agent},
			},
			live: StackConfig{
				Mcps:   map[string]McpConfig{"tools": liveMcp},
				Agents: map[string]AgentConfig{"helper": agent},
			},
			wantMcps:    movedMcp,
			wantAgents:  deferredAgent,
			wantPrinted: "Stack: s\n\n  ? agent helper (plan deferred; changing mcps: tools)\n\n  ~ mcp tools (update)\n    endpointUrl: https://old.example/mcp → https://new.example/mcp\n",
		},
		{
			name:  "deleted dependencies do not make an agent look unchanged",
			local: StackConfig{Agents: map[string]AgentConfig{"helper": agent}},
			live: StackConfig{
				Mcps:   map[string]McpConfig{"tools": liveMcp},
				Agents: map[string]AgentConfig{"helper": agent},
			},
			wantMcps:    `{"created":null,"updated":null,"deleted":["tools"]}`,
			wantAgents:  deferredAgent,
			wantPrinted: "Stack: s\n\n  ? agent helper (plan deferred; changing mcps: tools)\n\n  - mcp tools (delete)\n",
		},
		{
			name: "new agents still report creation",
			local: StackConfig{
				Mcps:   map[string]McpConfig{"tools": rotated},
				Agents: map[string]AgentConfig{"helper": agent},
			},
			wantMcps:    createdMcp,
			wantAgents:  `{"created":["helper"],"updated":null,"deleted":null}`,
			wantPrinted: "Stack: s\n\n  + agent helper (create)\n\n  + mcp tools (create)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := PlanStack(
				context.Background(),
				nil,
				"org",
				"project",
				"s",
				&tt.local,
				&tt.live,
			)
			if err != nil {
				t.Fatal(err)
			}
			diff := DiffStackConfigs(plan, &tt.live)
			mcps, err := json.Marshal(diff.Mcps)
			if err != nil {
				t.Fatal(err)
			}
			if string(mcps) != tt.wantMcps {
				t.Fatalf("mcps = %s, want %s", mcps, tt.wantMcps)
			}
			agents, err := json.Marshal(diff.Agents)
			if err != nil {
				t.Fatal(err)
			}
			if string(agents) != tt.wantAgents {
				t.Fatalf("agents = %s, want %s", agents, tt.wantAgents)
			}
			var out bytes.Buffer
			if err := PrintStackDiffDetailed(&out, plan, &tt.live, diff); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.wantPrinted {
				t.Fatalf("printed = %q, want %q", out.String(), tt.wantPrinted)
			}
		})
	}
}
