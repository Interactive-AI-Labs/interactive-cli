package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

// addChoiceFlag puts the choices in help and shell completion; it does not validate the value.
func addChoiceFlag(
	cmd *cobra.Command,
	target *string,
	name, value, usage string,
	choices []string,
) {
	cmd.Flags().StringVar(target, name, value, usage+" ("+strings.Join(choices, "|")+")")
	complete := cobra.FixedCompletions(choices, cobra.ShellCompDirectiveNoFileComp)
	if err := cmd.RegisterFlagCompletionFunc(name, complete); err != nil {
		flagSetupErrs = append(flagSetupErrs, err)
	}
}

var sortOrders = []string{"asc", "desc"}
