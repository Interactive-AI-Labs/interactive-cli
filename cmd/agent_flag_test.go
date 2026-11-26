package cmd

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
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
		{queuesListCmd, []string{"--agent", "a", "--sort-by", "name"}, ""},
		{queueItemsListCmd, []string{"--agent", "a", "--status", "PENDING"}, ""},
		{datasetsListCmd, []string{"--agent", "a"}, ""},
		{scoresListCmd, []string{"--agent", "a", "--user-id", "u"}, ""},
		{metricsListCmd, []string{"--agent", "a", "--environment", "e"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.cmd.CommandPath()+" "+strings.Join(tt.args, " "), func(t *testing.T) {
			t.Cleanup(func() {
				tracesAgent, tracesUserID, sessionsAgent, sessionsEnvironment = "", "", "", ""
				obsListAgent, obsListModel, obsTraceID = "", "", ""
				queuesListAgent, queuesListSortBy, datasetsListAgent = "", "", ""
				queueItemsListAgent, queueItemsListStatus = "", ""
				scoresAgent, scoresUserID, metricsListAgent, metricsListEnvironment = "", "", "", ""
				for _, name := range []string{
					"agent", "user-id", "environment", "model", "trace-id", "sort-by", "status",
				} {
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

func TestAgentName(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr string
	}{
		{"my-agent", "my-agent", ""},
		{" my-agent ", "my-agent", ""},
		{" ", "", "agent name is required"},
		{"", "", "agent name is required"},
	}
	for _, tt := range tests {
		got, err := agentName(tt.input)
		gotErr := ""
		if err != nil {
			gotErr = err.Error()
		}
		if got != tt.want || gotErr != tt.wantErr {
			t.Errorf(
				"agentName(%q) = %q, %q; want %q, %q",
				tt.input,
				got,
				gotErr,
				tt.want,
				tt.wantErr,
			)
		}
	}
}

func saveSelection(t *testing.T, agent string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("INTERACTIVE_ORGANIZATION", "")
	t.Setenv("INTERACTIVE_PROJECT", "")
	cfg := &files.Config{SelectedOrg: "org", SelectedProject: "proj", SelectedAgent: agent}
	if err := files.SaveConfig(cfgDirName, cfg); err != nil {
		t.Fatal(err)
	}
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestSelectedAgentFallback(t *testing.T) {
	const noAgent = "accepts 1 arg(s), received 0: pass the agent name or run 'iai agents select'"
	tests := []struct {
		name     string
		selected string
		args     []string
		flag     string
		project  string
		want     string
	}{
		{"typed agent wins", "sel", []string{"typed"}, "typed", "", "typed typed "},
		{"typed name sent as typed", "sel", []string{" "}, "", "", " sel "},
		{"selected agent used", "sel", nil, "", "", "sel sel "},
		{"no selected agent", "", nil, "", "", "  " + noAgent},
		{"selected on another project", "sel", nil, "", "other", "  " + noAgent},
		{"selected project in another case", "sel", nil, "", "PROJ", "sel sel "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveSelection(t, tt.selected)
			arg, _, argErr := agentArg(tt.args, 1, "", tt.project)
			flag, flagErr := agentFlag(tt.flag, "", tt.project)
			got := strings.Join([]string{arg, flag, errText(argErr) + errText(flagErr)}, " ")
			if got != tt.want {
				t.Errorf("agent, flag, error = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSplitAgentArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		n        int
		selected string
		want     string
	}{
		{"events create typed", []string{"a", "sess"}, 2, "sel", "a [sess] "},
		{"events create selected", []string{"sess"}, 2, "sel", "sel [sess] "},
		{"routines run typed", []string{"a", "billing/refund"}, 2, "sel", "a [billing/refund] "},
		{"routines run selected", []string{"billing/refund"}, 2, "sel", "sel [billing/refund] "},
		{"events get typed", []string{"a", "sess", "evt"}, 3, "sel", "a [sess evt] "},
		{"events get selected", []string{"sess", "evt"}, 3, "sel", "sel [sess evt] "},
		{"diff typed", []string{"a", "3", "5"}, 3, "", "a [3 5] "},
		{"diff selected", []string{"3", "5"}, 3, "sel", "sel [3 5] "},
		{
			"routines run without routine or selection",
			[]string{"my-agent"},
			2, "",
			" [] accepts 2 arg(s), received 1: pass the agent name or run 'iai agents select'",
		},
		{
			"diff without agent or selection",
			[]string{"my-agent", "3"},
			3, "",
			" [] accepts 3 arg(s), received 2: pass the agent name or run 'iai agents select'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, rest, err := splitAgentArgs(tt.args, tt.n, tt.selected)
			got := fmt.Sprintf("%s %v %s", agent, rest, errText(err))
			if got != tt.want {
				t.Errorf("agent, rest, error = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAgentSelectArgs(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		clear bool
		want  string
	}{
		{"name plus clear", []string{"bot"}, true, "pass an agent name or --clear, not both"},
		{"no name", nil, false, "agent name is required: pass the agent name, or --clear"},
		{
			"blank name",
			[]string{" "},
			false,
			"agent name is required: pass the agent name, or --clear",
		},
		{"clear", nil, true, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saveSelection(t, "sel")
			agentSelectClear = tt.clear
			t.Cleanup(func() { agentSelectClear = false })
			cmd := &cobra.Command{}
			cmd.SetOut(io.Discard)
			if got := errText(agentSelectCmd.RunE(cmd, tt.args)); got != tt.want {
				t.Errorf("error = %q, want %q", got, tt.want)
			}
		})
	}
}
