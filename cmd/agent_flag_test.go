package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func TestAgentObservabilityFlags(t *testing.T) {
	tests := []struct {
		name    string
		list    *cobra.Command
		dropped []string
	}{
		{"traces", tracesListCmd, []string{"agent"}},
		{"trace-sessions", sessionsListCmd, []string{"agent"}},
		{"observations", obsListCmd, []string{"agent", "include-io", "trace-id"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _, _ := agentsCmd.Find([]string{tt.name})
			got := []string{c.Name()}
			tt.list.LocalFlags().VisitAll(func(f *pflag.Flag) {
				if c.Flags().Lookup(f.Name) == nil {
					got = append(got, f.Name)
				}
			})
			if diff := cmp.Diff(append([]string{tt.name}, tt.dropped...), got); diff != "" {
				t.Errorf("command and dropped flags (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAgentObservabilityAgentArg(t *testing.T) {
	tests := []struct {
		name  string
		list  *cobra.Command
		agent *string
	}{
		{"traces", tracesListCmd, &tracesAgent},
		{"trace-sessions", sessionsListCmd, &sessionsAgent},
		{"observations", obsListCmd, &obsListAgent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := tt.list.RunE
			t.Cleanup(func() { tt.list.RunE = run; *tt.agent = "" })
			tt.list.RunE = func(*cobra.Command, []string) error { return nil }
			c, _, _ := agentsCmd.Find([]string{tt.name})
			var got []string
			for _, arg := range []string{" my-agent ", " "} {
				err := c.RunE(c, []string{arg})
				got = append(got, fmt.Sprint(*tt.agent, " ", err))
			}
			want := []string{"my-agent <nil>", " agent name is required"}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("agent and error per argument (-want +got):\n%s", diff)
			}
		})
	}
}
