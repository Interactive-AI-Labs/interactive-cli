package cmd

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/google/go-cmp/cmp"
)

func TestAgentListFlag(t *testing.T) {
	const usage = "Filter by agent name; defaults to the selected agent"
	got, want := map[string]string{}, map[string]string{}
	for _, list := range []string{"policies", "routines", "glossaries", "macros", "prompts", "mcps", "collections"} {
		cmd, _, err := rootCmd.Find([]string{list, "list"})
		if err != nil {
			t.Fatal(err)
		}
		if f := cmd.Flags().Lookup("agent"); f != nil {
			got[list] = f.Usage
		}
		want[list] = usage
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("--agent usage per list (-want +got):\n%s", diff)
	}
}

func TestAgentCollection(t *testing.T) {
	search := func(s map[string]any) any { return map[string]any{"search": s} }
	kb := map[string]any{"type": "collections", "database": "kb-db", "collection": "docs"}
	tests := []struct {
		name     string
		config   any
		database string
		want     string
	}{
		{"collection search", search(kb), "", "kb-db/docs"},
		{"same -d", search(kb), "kb-db", "kb-db/docs"},
		{"other -d", search(kb), "other-db", "/"},
		{"external search", search(map[string]any{"type": "external", "url": "u"}), "", "/"},
		{"no search", map[string]any{"database": map[string]any{"name": "sessions"}}, "", "/"},
		{
			"no collection",
			search(map[string]any{"type": "collections", "database": "kb-db"}),
			"",
			"/",
		},
		{"no config", nil, "", "/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, coll := agentCollection(tt.config, tt.database)
			if got := db + "/" + coll; got != tt.want {
				t.Errorf("agentCollection = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWithDrafts(t *testing.T) {
	live := deployment.AgentOutput{Name: "live", Status: "Running", Revision: 2}
	tests := []struct {
		name   string
		agents []deployment.AgentOutput
		drafts []platform.AgentDraft
		want   []deployment.AgentOutput
	}{
		{"no drafts", []deployment.AgentOutput{live}, nil, []deployment.AgentOutput{live}},
		{
			"draft appended with its update time",
			[]deployment.AgentOutput{live},
			[]platform.AgentDraft{
				{Name: "wip", ProjectId: "proj-1", Status: "draft", CreatedAt: "c", UpdatedAt: "u"},
			},
			[]deployment.AgentOutput{
				live,
				{Name: "wip", ProjectId: "proj-1", Status: "draft", Updated: "u"},
			},
		},
		{
			"creation time when never updated",
			nil,
			[]platform.AgentDraft{{Name: "wip", Status: "draft", CreatedAt: "c"}},
			[]deployment.AgentOutput{{Name: "wip", Status: "draft", Updated: "c"}},
		},
		{
			"draft named like a deployed agent is skipped",
			[]deployment.AgentOutput{live},
			[]platform.AgentDraft{{Name: "live", Status: "draft"}},
			[]deployment.AgentOutput{live},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, withDrafts(tt.agents, tt.drafts)); diff != "" {
				t.Errorf("agents (-want +got):\n%s", diff)
			}
		})
	}
}
