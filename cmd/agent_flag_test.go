package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestAgentFlag(t *testing.T) {
	tests := []struct {
		cmd     *cobra.Command
		args    []string
		wantErr string
	}{
		{tracesListCmd, []string{"--agent", "a", "--user-id", "u"}, ""},
		{sessionsListCmd, []string{"--agent", "a", "--environment", "e"}, ""},
		{obsListCmd, []string{"--agent", "a", "--model", "m"}, ""},
		{obsListCmd, []string{"--trace-id", "t"}, ""},
		{obsListCmd, []string{"--agent", "a", "--trace-id", "t"}, "[agent trace-id] were all set"},
	}
	for _, tt := range tests {
		t.Run(tt.cmd.CommandPath()+" "+strings.Join(tt.args, " "), func(t *testing.T) {
			t.Cleanup(func() {
				tracesAgent, tracesUserID, sessionsAgent, sessionsEnvironment = "", "", "", ""
				obsListAgent, obsListModel, obsTraceID = "", "", ""
				for _, name := range []string{"agent", "user-id", "environment", "model", "trace-id"} {
					if f := tt.cmd.Flags().Lookup(name); f != nil {
						f.Changed = false
					}
				}
			})
			err := tt.cmd.ParseFlags(tt.args)
			if err == nil {
				err = tt.cmd.ValidateFlagGroups()
			}
			if (err == nil) != (tt.wantErr == "") ||
				err != nil && !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
