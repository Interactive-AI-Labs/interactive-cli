package cmd

import "testing"

func TestMacrosSchemaCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "plural", args: []string{"macros", "schema"}},
		{name: "singular alias", args: []string{"macro", "schema"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, rest, err := rootCmd.Find(tt.args)
			if err != nil {
				t.Fatalf("Find(%v) error = %v", tt.args, err)
			}
			if cmd.Name() != "schema" || cmd.Parent().Name() != "macros" || len(rest) != 0 {
				t.Fatalf(
					"Find(%v) = %q with args %v, want macros schema",
					tt.args,
					cmd.CommandPath(),
					rest,
				)
			}
			for _, flag := range []string{"schema-version", "json", "yaml"} {
				if cmd.Flags().Lookup(flag) == nil {
					t.Errorf("%s has no --%s flag", cmd.CommandPath(), flag)
				}
			}
		})
	}
}
