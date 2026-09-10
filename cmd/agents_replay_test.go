package cmd

import (
	"strings"
	"testing"
)

func resetReplayFlags(t *testing.T) {
	t.Helper()
	origHostname, origDeployHostname, origToken, origApiKey := hostname, deploymentHostname, token, apiKey
	origOrg, origProject := agentOrganization, agentProject
	t.Cleanup(func() {
		hostname, deploymentHostname, token, apiKey = origHostname, origDeployHostname, origToken, origApiKey
		agentOrganization, agentProject = origOrg, origProject
		replayDataset, replayFile, replayRunID = "", "", ""
		replayScenarios = nil
		replayRepeat, replayConcurrency, replayJSON = 1, 8, false
		for _, name := range []string{
			"dataset", "scenarios", "file", "run-id", "repeat", "concurrency",
			"timeout", "json",
		} {
			agentReplayCmd.Flags().Lookup(name).Changed = false
		}
		agentReplayCmd.SetOut(nil)
		agentReplayCmd.SetErr(nil)
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
