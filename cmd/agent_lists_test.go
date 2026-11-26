package cmd

import (
	"errors"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/google/go-cmp/cmp"
)

func TestAgentListFlag(t *testing.T) {
	const usage = "Filter by agent name; defaults to the selected agent"
	got, want := map[string]string{}, map[string]string{}
	for _, list := range []string{"policies", "routines", "glossaries", "macros", "prompts", "mcps"} {
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

func TestDraftOrError(t *testing.T) {
	type result struct {
		Agent *deployment.DescribeAgentResponse
		Err   string
	}
	draft := &platform.AgentDraft{
		Id: "d-1", Name: "wip", ProjectId: "proj-1", Status: "draft", AgentVersion: "0.9",
		AgentConfig: map[string]any{"model": "m"}, CreatedAt: "c",
	}
	tests := []struct {
		name     string
		draft    *platform.AgentDraft
		draftErr error
		want     result
	}{
		{"draft described", draft, nil, result{Agent: &deployment.DescribeAgentResponse{
			Id: "d-1", Name: "wip", ProjectId: "proj-1", Status: "draft", Version: "0.9",
			AgentConfig: map[string]any{"model": "m"}, Updated: "c",
		}}},
		{
			"no draft keeps the operator's error",
			nil,
			errors.New("not found"),
			result{Err: "agent not found"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := draftOrError(tt.draft, tt.draftErr, errors.New("agent not found"))
			if diff := cmp.Diff(tt.want, result{agent, errText(err)}); diff != "" {
				t.Errorf("result (-want +got):\n%s", diff)
			}
		})
	}
}
