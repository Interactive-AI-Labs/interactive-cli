package cmd

import (
	"strings"
	"testing"
)

// resetReplayFlags restores what the cases below change: the flag values they
// set and the Changed state cobra reads when it validates the flag groups.
func resetReplayFlags(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		replayDataset, replayFile, replayRunID = "", "", ""
		replayScenarios = nil
		replayRepeat = 1
		for _, name := range []string{"dataset", "file", "run-id", "scenarios", "repeat"} {
			agentReplayCmd.Flags().Lookup(name).Changed = false
		}
	})
}

func TestAgentReplayFlagGroups(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "dataset only", args: []string{"--dataset", "d"}},
		{name: "run-id only", args: []string{"--run-id", "r"}},
		{
			name:    "dataset and file",
			args:    []string{"--dataset", "d", "--file", "f"},
			wantErr: "[dataset file] were all set",
		},
		{
			name:    "scenarios and run-id",
			args:    []string{"--run-id", "r", "--scenarios", "a"},
			wantErr: "[run-id scenarios] were all set",
		},
		{
			name:    "repeat and run-id",
			args:    []string{"--run-id", "r", "--repeat", "2"},
			wantErr: "[repeat run-id] were all set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetReplayFlags(t)
			if err := agentReplayCmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("ParseFlags: %v", err)
			}
			err := agentReplayCmd.ValidateFlagGroups()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
