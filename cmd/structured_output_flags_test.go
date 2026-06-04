package cmd

import (
	"strings"
	"testing"
)

func TestStructuredOutputFlagsAreMutuallyExclusive(t *testing.T) {
	tests := []struct {
		name string
		cmd  func() []string
	}{
		{
			name: "generic prompt list",
			cmd: func() []string {
				cmd := makeGenericListCmd()
				cmd.SetArgs([]string{"--json", "--yaml"})
				return executeCommand(cmd)
			},
		},
		{
			name: "generic prompt get",
			cmd: func() []string {
				cmd := makeGenericGetCmd()
				cmd.SetArgs([]string{"example", "--json", "--yaml"})
				return executeCommand(cmd)
			},
		},
		{
			name: "typed prompt list",
			cmd: func() []string {
				cmd := makeListCmd(PromptTypeConfig{TypeName: "routine", Plural: "routines"})
				cmd.SetArgs([]string{"--json", "--yaml"})
				return executeCommand(cmd)
			},
		},
		{
			name: "typed prompt get",
			cmd: func() []string {
				cmd := makeGetCmd(PromptTypeConfig{TypeName: "routine", Plural: "routines"})
				cmd.SetArgs([]string{"example", "--json", "--yaml"})
				return executeCommand(cmd)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmd()
			if len(result) != 1 {
				t.Fatalf("expected one result, got %d", len(result))
			}
			if !strings.Contains(result[0], "if any flags in the group [json yaml]") {
				t.Fatalf("expected mutual exclusion error, got %q", result[0])
			}
		})
	}
}

func executeCommand(cmd interface{ Execute() error }) []string {
	if err := cmd.Execute(); err != nil {
		return []string{err.Error()}
	}
	return nil
}
