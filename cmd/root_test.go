package cmd

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

func TestHelpGroups(t *testing.T) {
	want := map[string]string{
		"api-keys":                    groupAuth,
		"login":                       groupAuth,
		"logout":                      groupAuth,
		"organizations":               groupAuth,
		"projects":                    groupAuth,
		"agents":                      groupAgents,
		"skills":                      groupAgents,
		"collections":                 groupContext,
		"glossaries":                  groupContext,
		"macros":                      groupContext,
		"policies":                    groupContext,
		"prompts":                     groupContext,
		"routines":                    groupContext,
		"variables":                   groupContext,
		"mcps":                        groupConnectors,
		"router":                      groupConnectors,
		"secrets":                     groupConnectors,
		"databases":                   groupInfra,
		"files":                       groupInfra,
		"images":                      groupInfra,
		"jobs":                        groupInfra,
		"replicas":                    groupInfra,
		"services":                    groupInfra,
		"stacks":                      groupInfra,
		"comments":                    groupGovernance,
		"metrics":                     groupGovernance,
		"observations":                groupGovernance,
		"reports":                     groupGovernance,
		"score-configs":               groupImprovement,
		"scores":                      groupImprovement,
		"sessions":                    groupGovernance,
		"traces":                      groupGovernance,
		"dataset-items":               groupImprovement,
		"dataset-runs":                groupImprovement,
		"datasets":                    groupImprovement,
		"queue-items":                 groupImprovement,
		"queues":                      groupImprovement,
		"run-items":                   groupImprovement,
		"gen-docs":                    "",
		"gen-index":                   "",
		"update":                      "",
		"agents create":               groupAgentCreate,
		"agents activate":             groupAgentManage,
		"agents deactivate":           groupAgentManage,
		"agents delete":               groupAgentManage,
		"agents describe":             groupAgentManage,
		"agents diff":                 groupAgentManage,
		"agents list":                 groupAgentManage,
		"agents log-fields":           groupAgentManage,
		"agents logs":                 groupAgentManage,
		"agents port-forward":         groupAgentManage,
		"agents restart":              groupAgentManage,
		"agents revisions":            groupAgentManage,
		"agents update":               groupAgentManage,
		"agents chat":                 groupAgentConversations,
		"agents events":               groupAgentConversations,
		"agents replay":               groupAgentConversations,
		"agents routines":             groupAgentConversations,
		"agents sessions":             groupAgentConversations,
		"agents observations":         groupAgentObservability,
		"agents trace-sessions":       groupAgentObservability,
		"agents traces":               groupAgentObservability,
		"agents overview":             groupAgentObservability,
		"agents activity":             groupAgentObservability,
		"agents analytics":            groupAgentObservability,
		"agents catalog":              groupAgentReference,
		"agents compatibility-matrix": groupAgentReference,
		"agents schema":               groupAgentReference,
		"agents select":               groupAgentManage,
		"sections":                    "Auth: Agents: Context: Connectors: Infrastructure: Governance: Improvement:",
		"agents sections":             "Create: Manage: Conversations: Observability: Reference:",
	}
	got := map[string]string{
		"sections":        titles(rootCmd.Groups()),
		"agents sections": titles(agentsCmd.Groups()),
	}
	for _, c := range rootCmd.Commands() {
		// Cobra adds these at Execute; they stay under "Additional Commands".
		if c.Name() != "help" && c.Name() != "completion" {
			got[c.Name()] = c.GroupID
		}
	}
	for _, c := range agentsCmd.Commands() {
		got["agents "+c.Name()] = c.GroupID
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("help groups (-want +got):\n%s", diff)
	}
}

func titles(groups []*cobra.Group) string {
	var out []string
	for _, g := range groups {
		out = append(out, g.Title)
	}
	return strings.Join(out, " ")
}

func TestRefuseBlankArgs(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"a"}, ""},
		{[]string{" "}, "arguments cannot be blank"},
		{[]string{""}, "arguments cannot be blank"},
		{[]string{"a", "b"}, "accepts 1 arg(s), received 2"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, ","), func(t *testing.T) {
			root := &cobra.Command{Use: "root"}
			root.AddCommand(
				&cobra.Command{
					Use:  "delete",
					Args: cobra.ExactArgs(1),
					RunE: func(*cobra.Command, []string) error { return nil },
				},
			)
			refuseBlankArgs(root)
			root.SetArgs(append([]string{"delete"}, tt.args...))
			root.SilenceErrors, root.SilenceUsage = true, true
			got := ""
			if err := root.Execute(); err != nil {
				got = err.Error()
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
