package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHelpGroups(t *testing.T) {
	want := map[string]string{
		"api-keys":      groupAuth,
		"login":         groupAuth,
		"logout":        groupAuth,
		"organizations": groupAuth,
		"projects":      groupAuth,
		"agents":        groupAgents,
		"skills":        groupAgents,
		"collections":   groupContext,
		"glossaries":    groupContext,
		"macros":        groupContext,
		"policies":      groupContext,
		"prompts":       groupContext,
		"routines":      groupContext,
		"variables":     groupContext,
		"mcps":          groupIntegrations,
		"router":        groupIntegrations,
		"secrets":       groupIntegrations,
		"databases":     groupInfra,
		"files":         groupInfra,
		"images":        groupInfra,
		"jobs":          groupInfra,
		"replicas":      groupInfra,
		"services":      groupInfra,
		"stacks":        groupInfra,
		"comments":      groupObserve,
		"metrics":       groupObserve,
		"observations":  groupObserve,
		"reports":       groupObserve,
		"score-configs": groupObserve,
		"scores":        groupObserve,
		"sessions":      groupObserve,
		"traces":        groupObserve,
		"dataset-items": groupEvaluation,
		"dataset-runs":  groupEvaluation,
		"datasets":      groupEvaluation,
		"queue-items":   groupEvaluation,
		"queues":        groupEvaluation,
		"run-items":     groupEvaluation,
		"gen-docs":      "",
		"gen-index":     "",
		"update":        "",
	}
	got := map[string]string{}
	for _, c := range rootCmd.Commands() {
		got[c.Name()] = c.GroupID
	}
	// An ungrouped agents child would land under "Additional Commands".
	for _, c := range agentsCmd.Commands() {
		if c.GroupID == "" {
			got["agents "+c.Name()] = ""
		}
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("help groups (-want +got):\n%s", diff)
	}
}
