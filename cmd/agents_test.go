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
