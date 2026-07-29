package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// TestAgentUpdateRefusesUnpinnedRoutine runs the full update command with a
// --file config that keeps a routine entry but drops its version. Unpinning
// has the same effect as removing the pin, so the update must be refused
// before the PATCH unless --force is passed.
func TestAgentUpdateRefusesUnpinnedRoutine(t *testing.T) {
	patchCalls := 0
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
						`"agentConfig":{"context":{"routines":[{"id":"welcome","version":13}]}}}`,
				)
			case r.Method == http.MethodPatch &&
				r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents/chat-agent":
				patchCalls++
				fmt.Fprint(w, `{"message":"Update submitted"}`)
			default:
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(server.Close)

	configPath := filepath.Join(t.TempDir(), "agent.yaml")
	config := "context:\n  routines:\n    - id: welcome\n"
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	t.Setenv("HOME", t.TempDir())
	origHostname, origDeployHostname, origToken, origApiKey := hostname, deploymentHostname, token, apiKey
	origOrg, origProject := agentOrganization, agentProject
	t.Cleanup(func() {
		hostname, deploymentHostname, token, apiKey = origHostname, origDeployHostname, origToken, origApiKey
		agentOrganization, agentProject = origOrg, origProject
		agentFile = ""
		agentUpdateCmd.Flags().Lookup("file").Changed = false
		agentUpdateCmd.SetOut(nil)
		agentUpdateCmd.SetErr(nil)
	})
	hostname, deploymentHostname, token, apiKey = server.URL, server.URL, "test-token", ""
	agentOrganization, agentProject = "acme", "alunafi"

	if err := agentUpdateCmd.Flags().Set("file", configPath); err != nil {
		t.Fatalf("set --file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	agentUpdateCmd.SetOut(&stdout)
	agentUpdateCmd.SetErr(&stderr)
	agentUpdateCmd.SetContext(context.Background())

	err := agentUpdateCmd.RunE(agentUpdateCmd, []string{"chat-agent"})

	wantErr := "refusing to apply: this update downgrades or removes live content pins" +
		" (details above) — pass --force if this is intended"
	if err == nil || err.Error() != wantErr {
		t.Fatalf("error = %v, want %q", err, wantErr)
	}
	if patchCalls != 0 {
		t.Errorf("PATCH called %d times before the gate, want 0", patchCalls)
	}
	wantWarn := "routine welcome: v13 → (unpinned)  (REMOVED — this update drops the version pin)"
	if !strings.Contains(stderr.String(), wantWarn) {
		t.Errorf("stderr missing unpin detail %q:\n%s", wantWarn, stderr.String())
	}
}

// TestAgentUpdateGatesDroppedEnv runs the full update command with an --env
// replacement that keeps only one of the two live env vars. --env replaces
// the whole list, so "add one var" passed alone silently wipes the rest:
// the update is refused before the PATCH with the dropped names detailed on
// stderr, and --force is the single override that lets it through. The
// untouched secret list must stay silent either way.
func TestAgentUpdateGatesDroppedEnv(t *testing.T) {
	tests := []struct {
		name  string
		force bool
	}{
		{name: "refused without force"},
		{name: "applied with force", force: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patchCalls := 0
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
						patchCalls++
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
				agentForce = false
				agentUpdateCmd.Flags().Lookup("env").Changed = false
				agentUpdateCmd.SetOut(nil)
				agentUpdateCmd.SetErr(nil)
			})
			hostname, deploymentHostname, token, apiKey = server.URL, server.URL, "test-token", ""
			agentOrganization, agentProject = "acme", "alunafi"
			agentForce = tt.force

			if err := agentUpdateCmd.Flags().Set("env", "LOG_LEVEL=debug"); err != nil {
				t.Fatalf("set --env: %v", err)
			}

			var stdout, stderr bytes.Buffer
			agentUpdateCmd.SetOut(&stdout)
			agentUpdateCmd.SetErr(&stderr)
			agentUpdateCmd.SetContext(context.Background())

			err := agentUpdateCmd.RunE(agentUpdateCmd, []string{"chat-agent"})

			errOut := stderr.String()
			want := "⚠ this update drops live env vars: DB_HOST" +
				" (--env replaces the entire list; pass every value you want to keep)"
			if !strings.Contains(errOut, want) {
				t.Errorf("stderr missing dropped-env warning %q:\n%s", want, errOut)
			}
			if strings.Contains(errOut, "secret refs") {
				t.Errorf("untouched --secret list must not warn:\n%s", errOut)
			}

			if tt.force {
				if err != nil {
					t.Fatalf("agents update with --force: %v", err)
				}
				if patchCalls != 1 {
					t.Errorf("PATCH called %d times, want 1", patchCalls)
				}
				if !strings.Contains(stdout.String(), "Update submitted") {
					t.Errorf("update did not go through:\n%s", stdout.String())
				}
				return
			}

			wantErr := "refusing to apply: this update drops live env vars or secret refs" +
				" (details above) — pass --force if this is intended"
			if err == nil || err.Error() != wantErr {
				t.Fatalf("error = %v, want %q", err, wantErr)
			}
			if patchCalls != 0 {
				t.Errorf("PATCH called %d times before the gate, want 0", patchCalls)
			}
		})
	}
}
