package cmd

import (
	"errors"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// agentObservabilityCmd runs list for one agent, with list's flags except drop.
func agentObservabilityCmd(
	name string, list *cobra.Command, agent *string, drop ...string,
) *cobra.Command {
	c := &cobra.Command{
		Use:     name + " <agent_name>",
		Short:   "List an agent's " + strings.ReplaceAll(name, "-", " "),
		GroupID: groupAgentObservability,
		Example: "  iai agents " + name + " my-agent\n" +
			"  iai agents " + name + " my-agent --limit 20 --json",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			*agent = strings.TrimSpace(args[0])
			if *agent == "" {
				return errors.New("agent name is required")
			}
			return list.RunE(cmd, nil)
		},
	}
	list.Flags().VisitAll(func(f *pflag.Flag) {
		if slices.Contains(drop, f.Name) {
			return
		}
		// Copied so help and Changed are per command; the value stays shared.
		nf := *f
		c.Flags().AddFlag(&nf)
		if complete, ok := list.GetFlagCompletionFunc(f.Name); ok {
			if err := c.RegisterFlagCompletionFunc(f.Name, complete); err != nil {
				flagSetupErrs = append(flagSetupErrs, err)
			}
		}
	})
	return c
}
