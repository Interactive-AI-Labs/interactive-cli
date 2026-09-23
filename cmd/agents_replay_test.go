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
		replayExperimentName = ""
		replayExperimentNameReuse, replayKeepSessions = false, false
		for _, name := range []string{
			"dataset", "file", "run-id", "scenarios", "repeat",
			"experiment-name", "experiment-name-reuse", "keep-sessions",
		} {
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
		{
			name:    "experiment-name and run-id",
			args:    []string{"--run-id", "r", "--experiment-name", "sweep"},
			wantErr: "[experiment-name run-id] were all set",
		},
		{
			name:    "keep-sessions and run-id",
			args:    []string{"--run-id", "r", "--keep-sessions"},
			wantErr: "[keep-sessions run-id] were all set",
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

func TestNormalizeAgentURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "unset means the platform", raw: ""},
		{name: "blank means the platform", raw: "   "},
		{
			name: "a full url passes through",
			raw:  "http://127.0.0.1:8080",
			want: "http://127.0.0.1:8080",
		},
		{
			name: "surrounding space is trimmed",
			raw:  " http://127.0.0.1:8080\n",
			want: "http://127.0.0.1:8080",
		},
		{
			name: "https is fine",
			raw:  "https://agent.example.com",
			want: "https://agent.example.com",
		},
		{name: "a host and port without a scheme is refused", raw: "localhost:8080", wantErr: true},
		{name: "an address without a scheme is refused", raw: "127.0.0.1:8080", wantErr: true},
		{name: "a scheme without a host is refused", raw: "http://", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAgentURL(tt.raw)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeAgentURL(%q) = %q, want an error", tt.raw, got)
				}
				if !strings.Contains(err.Error(), "scheme and host") {
					t.Errorf("error = %v, want it to name what is missing", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeAgentURL(%q): %v", tt.raw, err)
			}
			if got != tt.want {
				t.Errorf("normalizeAgentURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
