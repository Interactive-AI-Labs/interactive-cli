package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func resetReplayFlags(t *testing.T) {
	t.Helper()
	origHostname, origDeployHostname, origToken, origApiKey := hostname, deploymentHostname, token, apiKey
	origOrg, origProject := agentOrganization, agentProject
	t.Cleanup(func() {
		hostname, deploymentHostname, token, apiKey = origHostname, origDeployHostname, origToken, origApiKey
		agentOrganization, agentProject = origOrg, origProject
		replayDataset, replayFile, replayRunID, replayAgentURL, replayAPIKey = "", "", "", "", ""
		replayScenarios = nil
		replayRepeat, replayConcurrency, replayJSON = 1, 8, false
		for _, name := range []string{
			"dataset", "scenarios", "file", "run-id", "repeat", "concurrency",
			"timeout", "agent-url", "agent-api-key", "json",
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

func TestAgentReplayEndToEnd(t *testing.T) {
	var polls atomic.Int32
	agentServer := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer the-agent-key" {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, "Unauthorized")
				return
			}
			switch {
			case r.Method == http.MethodPost && r.URL.Path == "/replays":
				w.WriteHeader(http.StatusAccepted)
				fmt.Fprint(w, `{"run_id":"run-42","skipped":[],"status":"accepted"}`)
			case r.Method == http.MethodGet && r.URL.Path == "/replays/run-42":
				polls.Add(1)
				fmt.Fprint(
					w,
					`{"run_id":"run-42","dataset":"replay-chat","status":"failed","repeat":1,"concurrency":8,`+
						`"batches":[{"scenario":"account-lock","status":"failed","repeat":1,"passed":0,`+
						`"iterations":[{"status":"failed","turns":2,"eval_trace_id":"ev1",`+
						`"observed":{"steps":["a"]},"failures":["steps.reached: 'b' not observed"]}]}]}`,
				)
			default:
				t.Errorf("unexpected agent request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(agentServer.Close)

	encodedKey := base64.StdEncoding.EncodeToString([]byte("the-agent-key"))
	platform := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/api/v1/session/organizations":
				fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"acme"}]}`)
			case r.Method == http.MethodGet &&
				r.URL.Path == "/api/v1/session/organizations/org-1/projects":
				fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"support"}]}`)
			case r.Method == http.MethodGet &&
				r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents/agent-chat-dev":
				fmt.Fprint(w, `{"name":"agent-chat-dev","revision":592,"version":"0.15.1",`+
					`"endpoint":"agent-chat-dev.example.com",`+
					`"agentConfig":{"runtime":{"api_key":"${AGENT_API_KEY}"}},`+
					`"secretRefs":[{"secretName":"platform-dev"}]}`)
			case r.Method == http.MethodGet &&
				r.URL.Path == "/v1/organizations/org-1/projects/proj-1/secrets/platform-dev":
				fmt.Fprintf(
					w,
					`{"secret":{"name":"platform-dev","data":{"AGENT_API_KEY":%q}}}`,
					encodedKey,
				)
			default:
				t.Errorf("unexpected platform request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(platform.Close)

	t.Setenv("HOME", t.TempDir())
	t.Setenv("INTERACTIVE_AGENT_API_KEY", "")
	resetReplayFlags(t)
	hostname, deploymentHostname, token, apiKey = platform.URL, platform.URL, "test-token", ""
	agentOrganization, agentProject = "acme", "support"
	replayDataset, replayAgentURL = "replay-chat", agentServer.URL
	replayRepeat, replayConcurrency = 1, 8

	var stdout, stderr bytes.Buffer
	agentReplayCmd.SetOut(&stdout)
	agentReplayCmd.SetErr(&stderr)
	agentReplayCmd.SetContext(context.Background())

	// A failed verdict is a successful command.
	if err := agentReplayCmd.RunE(agentReplayCmd, []string{"agent-chat-dev"}); err != nil {
		t.Fatalf("agents replay: %v", err)
	}
	if polls.Load() != 1 {
		t.Errorf("polls = %d, want 1", polls.Load())
	}

	wantStdout := "dataset replay-chat   1 scenario   repeat 1   concurrency 8\n" +
		"\n" +
		"account-lock\n" +
		"--- run 1/1  FAILED ---\n" +
		"  turns          2\n" +
		"  steps          a\n" +
		"  eval trace     ev1\n" +
		"  FAIL           steps.reached: 'b' not observed\n" +
		"\n" +
		"FAIL   0/1 passed     run run-42\n"
	if got := stdout.String(); got != wantStdout {
		t.Errorf("stdout mismatch\ngot:\n%s\nwant:\n%s", got, wantStdout)
	}
	wantStderr := "agent-chat-dev  0.15.1  rev 592    " + agentServer.URL + "\n" +
		"using AGENT_API_KEY from secret platform-dev\n" +
		"1/1 scenarios finished\n" +
		"scores: iai scores list --name replay.verdict --columns name,value,trace_id,comment" +
		" · eval trace: iai traces get ev1\n"
	if got := stderr.String(); got != wantStderr {
		t.Errorf("stderr mismatch\ngot:\n%s\nwant:\n%s", got, wantStderr)
	}
	if strings.Contains(stdout.String()+stderr.String(), "the-agent-key") {
		t.Error("bearer leaked into output")
	}
}
