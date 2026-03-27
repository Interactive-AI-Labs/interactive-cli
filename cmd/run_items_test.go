package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRunItemsCreateRequiredFlags(t *testing.T) {
	tests := []string{"run-name", "dataset-item-id"}

	for _, flagName := range tests {
		t.Run(flagName, func(t *testing.T) {
			flag := runItemsCreateCmd.Flag(flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", flagName)
			}

			values, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]
			if !ok || len(values) == 0 || values[0] != "true" {
				t.Fatalf("flag %q is not marked required", flagName)
			}
		})
	}
}
