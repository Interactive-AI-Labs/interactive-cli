package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestAgentUpdateMcpOverlayShowDiff runs the full update command for an
// --mcp overlay (no --file) and checks the pre-flight diff on stderr.
// Regression test: the overlay helpers mutate the live config map in place,
// which used to make --show-diff compare the config against itself and
// print "No differences found." for the very change being applied.
func TestAgentUpdateMcpOverlayShowDiff(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/api/v1/session/organizations":
				fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"acme"}]}`)
			case r.Method == http.MethodGet &&
				r.URL.Path == "/api/v1/session/organizations/org-1/projects":
				fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"alunafi"}]}`)
			case r.Method == http.MethodGet &&
				r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents/chat-agent":
				fmt.Fprint(
					w,
					`{"name":"chat-agent","projectId":"proj-1","revision":13,"status":"ready",`+
						`"updated":"2026-07-24T11:20:00Z","id":"interactive-agent","version":"0.0.2",`+
						`"agentConfig":{"mcps":["github"],"context":{"routines":[{"id":"welcome","version":13}]}}}`,
				)
			case r.Method == http.MethodPatch &&
				r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents/chat-agent":
				fmt.Fprint(w, `{"message":"Update submitted"}`)
			default:
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(server.Close)

	// Isolate every global the command reads, and restore afterwards.
	t.Setenv("HOME", t.TempDir())
	origHostname, origDeployHostname, origToken, origApiKey := hostname, deploymentHostname, token, apiKey
	origOrg, origProject := agentOrganization, agentProject
	t.Cleanup(func() {
		hostname, deploymentHostname, token, apiKey = origHostname, origDeployHostname, origToken, origApiKey
		agentOrganization, agentProject = origOrg, origProject
		agentMcpNames = nil
		agentShowDiff = false
		for _, name := range []string{"mcp", "show-diff"} {
			agentUpdateCmd.Flags().Lookup(name).Changed = false
		}
		agentUpdateCmd.SetOut(nil)
		agentUpdateCmd.SetErr(nil)
	})
	hostname, deploymentHostname, token, apiKey = server.URL, server.URL, "test-token", ""
	agentOrganization, agentProject = "acme", "alunafi"

	if err := agentUpdateCmd.Flags().Set("mcp", "stripe"); err != nil {
		t.Fatalf("set --mcp: %v", err)
	}
	if err := agentUpdateCmd.Flags().Set("show-diff", "true"); err != nil {
		t.Fatalf("set --show-diff: %v", err)
	}

	var stdout, stderr bytes.Buffer
	agentUpdateCmd.SetOut(&stdout)
	agentUpdateCmd.SetErr(&stderr)
	agentUpdateCmd.SetContext(context.Background())

	if err := agentUpdateCmd.RunE(agentUpdateCmd, []string{"chat-agent"}); err != nil {
		t.Fatalf("agents update: %v", err)
	}

	wantStdout := "\nSubmitting agent update request...\nUpdate submitted\n"
	if got := stdout.String(); got != wantStdout {
		t.Errorf("stdout = %q, want %q", got, wantStdout)
	}

	errOut := stderr.String()
	if !strings.Contains(
		errOut,
		"Live: revision 13, last updated 2026-07-24 11:20 UTC — this update creates revision 14",
	) {
		t.Errorf("stderr missing revision banner:\n%s", errOut)
	}
	if strings.Contains(errOut, "No differences found.") {
		t.Errorf("diff compared the live config against itself:\n%s", errOut)
	}
	if !strings.Contains(errOut, "--- live") || !strings.Contains(errOut, "- stripe") {
		t.Errorf("stderr missing live-vs-incoming diff with the attached mcp:\n%s", errOut)
	}
	if strings.Contains(errOut, "content pins changed") {
		t.Errorf("mcp-only update must not report pin changes:\n%s", errOut)
	}
}

func TestAgentUpdateValidatesBeforeDescribe(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]string
		wantErr string
	}{
		{
			name:    "no fields",
			wantErr: "no fields to update; pass at least one flag",
		},
		{
			name: "show diff without config replacement",
			flags: map[string]string{
				"show-diff": "true",
				"version":   "0.0.3",
			},
			wantErr: "--show-diff requires --file, --mcp, or --detach-mcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			describeCalls := 0
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.Method == http.MethodGet &&
						r.URL.Path == "/api/v1/session/organizations":
						fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"acme"}]}`)
					case r.Method == http.MethodGet &&
						r.URL.Path == "/api/v1/session/organizations/org-1/projects":
						fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"alunafi"}]}`)
					case r.Method == http.MethodGet &&
						r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents/chat-agent":
						describeCalls++
						fmt.Fprint(w, `{"name":"chat-agent","revision":1,"agentConfig":{}}`)
					default:
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
						w.WriteHeader(http.StatusNotFound)
					}
				}),
			)
			t.Cleanup(server.Close)

			t.Setenv("HOME", t.TempDir())
			origHostname, origDeployHostname := hostname, deploymentHostname
			origToken, origApiKey := token, apiKey
			origOrg, origProject := agentOrganization, agentProject
			origVersion, origShowDiff := agentVersion, agentShowDiff
			originalChanged := map[string]bool{}
			for name := range tt.flags {
				originalChanged[name] = agentUpdateCmd.Flags().Lookup(name).Changed
			}
			t.Cleanup(func() {
				hostname, deploymentHostname = origHostname, origDeployHostname
				token, apiKey = origToken, origApiKey
				agentOrganization, agentProject = origOrg, origProject
				agentVersion, agentShowDiff = origVersion, origShowDiff
				for name, changed := range originalChanged {
					agentUpdateCmd.Flags().Lookup(name).Changed = changed
				}
				agentUpdateCmd.SetOut(nil)
				agentUpdateCmd.SetErr(nil)
			})

			hostname, deploymentHostname = server.URL, server.URL
			token, apiKey = "test-token", ""
			agentOrganization, agentProject = "acme", "alunafi"
			for name, value := range tt.flags {
				if err := agentUpdateCmd.Flags().Set(name, value); err != nil {
					t.Fatalf("set --%s: %v", name, err)
				}
			}

			agentUpdateCmd.SetOut(&bytes.Buffer{})
			agentUpdateCmd.SetErr(&bytes.Buffer{})
			agentUpdateCmd.SetContext(context.Background())

			err := agentUpdateCmd.RunE(agentUpdateCmd, []string{"chat-agent"})
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
			}
			if describeCalls != 0 {
				t.Errorf("DescribeAgent called %d times before validation", describeCalls)
			}
		})
	}
}

// TestAgentUpdateWarnsOnDroppedEnv runs the full update command with an
// --env replacement that keeps only one of the two live env vars and checks
// the dropped-name warning on stderr. --env replaces the whole list, so
// "add one var" passed alone silently wipes the rest — the warning is the
// only signal. The untouched secret list must stay silent.
func TestAgentUpdateWarnsOnDroppedEnv(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/api/v1/session/organizations":
				fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"acme"}]}`)
			case r.Method == http.MethodGet &&
				r.URL.Path == "/api/v1/session/organizations/org-1/projects":
				fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"alunafi"}]}`)
			case r.Method == http.MethodGet &&
				r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents/chat-agent":
				fmt.Fprint(
					w,
					`{"name":"chat-agent","projectId":"proj-1","revision":13,"status":"ready",`+
						`"updated":"2026-07-24T11:20:00Z","id":"interactive-agent","version":"0.0.2",`+
						`"agentConfig":{},`+
						`"env":[{"name":"LOG_LEVEL","value":"info"},{"name":"DB_HOST","value":"db"}],`+
						`"secretRefs":[{"secretName":"api-keys"}]}`,
				)
			case r.Method == http.MethodPatch &&
				r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents/chat-agent":
				fmt.Fprint(w, `{"message":"Update submitted"}`)
			default:
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(server.Close)

	t.Setenv("HOME", t.TempDir())
	origHostname, origDeployHostname, origToken, origApiKey := hostname, deploymentHostname, token, apiKey
	origOrg, origProject := agentOrganization, agentProject
	t.Cleanup(func() {
		hostname, deploymentHostname, token, apiKey = origHostname, origDeployHostname, origToken, origApiKey
		agentOrganization, agentProject = origOrg, origProject
		agentEnvVars = nil
		agentUpdateCmd.Flags().Lookup("env").Changed = false
		agentUpdateCmd.SetOut(nil)
		agentUpdateCmd.SetErr(nil)
	})
	hostname, deploymentHostname, token, apiKey = server.URL, server.URL, "test-token", ""
	agentOrganization, agentProject = "acme", "alunafi"

	if err := agentUpdateCmd.Flags().Set("env", "LOG_LEVEL=debug"); err != nil {
		t.Fatalf("set --env: %v", err)
	}

	var stdout, stderr bytes.Buffer
	agentUpdateCmd.SetOut(&stdout)
	agentUpdateCmd.SetErr(&stderr)
	agentUpdateCmd.SetContext(context.Background())

	if err := agentUpdateCmd.RunE(agentUpdateCmd, []string{"chat-agent"}); err != nil {
		t.Fatalf("agents update: %v", err)
	}

	errOut := stderr.String()
	want := "⚠ this update drops live env vars: DB_HOST" +
		" (--env replaces the entire list; pass every value you want to keep)"
	if !strings.Contains(errOut, want) {
		t.Errorf("stderr missing dropped-env warning %q:\n%s", want, errOut)
	}
	if strings.Contains(errOut, "secret refs") {
		t.Errorf("untouched --secret list must not warn:\n%s", errOut)
	}
	if !strings.Contains(stdout.String(), "Update submitted") {
		t.Errorf("update did not go through:\n%s", stdout.String())
	}
}
