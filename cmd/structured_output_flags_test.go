package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
		{
			name: "files list",
			cmd: func() []string {
				return validateFilesStructuredOutputFlagGroup(
					filesListCmd,
					&filesListJSON,
					&filesListYAML,
				)
			},
		},
		{
			name: "files get",
			cmd: func() []string {
				return validateFilesStructuredOutputFlagGroup(
					filesGetCmd,
					&filesGetJSON,
					&filesGetYAML,
				)
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

func validateFilesStructuredOutputFlagGroup(
	cmd *cobra.Command,
	asJSON, asYAML *bool,
) []string {
	jsonFlag := cmd.Flags().Lookup("json")
	yamlFlag := cmd.Flags().Lookup("yaml")
	originalJSON, originalYAML := *asJSON, *asYAML
	originalJSONChanged, originalYAMLChanged := jsonFlag.Changed, yamlFlag.Changed
	defer func() {
		*asJSON, *asYAML = originalJSON, originalYAML
		jsonFlag.Changed, yamlFlag.Changed = originalJSONChanged, originalYAMLChanged
	}()

	*asJSON, *asYAML = false, false
	jsonFlag.Changed, yamlFlag.Changed = false, false
	if err := cmd.Flags().Set("json", "true"); err != nil {
		return []string{err.Error()}
	}
	if err := cmd.Flags().Set("yaml", "true"); err != nil {
		return []string{err.Error()}
	}
	if err := cmd.ValidateFlagGroups(); err != nil {
		return []string{err.Error()}
	}
	return nil
}

func executeCommand(cmd interface{ Execute() error }) []string {
	if err := cmd.Execute(); err != nil {
		return []string{err.Error()}
	}
	return nil
}
