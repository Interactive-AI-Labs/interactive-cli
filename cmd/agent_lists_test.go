package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

const wipDraft = `{"id":"d-1","name":"wip","projectId":"proj-1","status":"draft",
	"agentConfig":{"model":"m"},"agentVersion":"0.9","description":null,
	"createdAt":"2026-11-01T00:00:00Z","updatedAt":null}`

func agentListsServer(t *testing.T, filters *[]string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/session/organizations":
			fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"org"}]}`)
		case r.URL.Path == "/api/v1/session/organizations/org-1/projects":
			fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"proj"}]}`)
		case strings.HasPrefix(r.URL.Path, "/api/platform/v1/projects/proj-1/prompts"):
			*filters = append(*filters, r.URL.Query().Get("filter"))
			if agent := r.URL.Query().Get("agent"); agent != "" {
				*filters = append(*filters, "agent="+agent)
			}
			fmt.Fprint(w, `{"success":true,"data":{"prompts":[],"totalCount":0}}`)
		case strings.Contains(r.URL.Path, "/logs"):
			*filters = append(*filters, r.URL.RequestURI())
			fmt.Fprint(w, `{"total":3,"debug":0,"info":1,"warn":0,"error":2}`)
		case r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents":
			fmt.Fprint(w, `{"agents":[{"name":"live","status":"Running","revision":2},
				{"name":"both","status":"Running","revision":1}]}`)
		case r.URL.Path == "/v1/organizations/org-1/projects/proj-1/agents/live":
			fmt.Fprint(w, `{"name":"live","status":"Running","revision":2}`)
		case strings.HasPrefix(r.URL.Path, "/v1/organizations/org-1/projects/proj-1/agents/"):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"message":"agent not found"}`)
		case r.URL.Path == "/api/platform/v1/projects/proj-1/build-agent-configs":
			fmt.Fprint(w, `{"items":[`+wipDraft+`,{"name":"both","status":"draft"}]}`)
		case r.URL.Path == "/api/platform/v1/projects/proj-1/build-agent-configs/wip":
			fmt.Fprint(w, wipDraft)
		case strings.HasPrefix(r.URL.Path, "/api/platform/v1/projects/proj-1/build-agent-configs/"):
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"detail":{"error":{"message":"not found"}}}`)
		case r.URL.Path == "/api/platform/v1/organizations/org-1/projects/proj-1/mcps":
			fmt.Fprint(w, `{"success":true,"data":{"internal_available":true,"mcps":[
				{"name":"crm","attached_agents":["bot"],"extra":1},
				{"name":"mail","attached_agents":["other"]},
				{"name":"docs","attached_agents":["Bot","bot-2"]},
				{"name":"idle","attached_agents":[]}]}}`)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	origHostname, origDeploy, origToken, origAPIKey := hostname, deploymentHostname, token, apiKey
	t.Cleanup(func() {
		hostname, deploymentHostname, token, apiKey = origHostname, origDeploy, origToken, origAPIKey
	})
	hostname, deploymentHostname, token, apiKey = server.URL, server.URL, "test-token", ""
}

func TestAgentListFilters(t *testing.T) {
	const filter = `[{"column":"Agents","operator":"any of","type":"arrayOptions","value":["bot"]}]`
	tests := []struct {
		list     []string
		selected string
		args     []string
		want     string
	}{
		{[]string{"policies", "list"}, "", []string{"--agent", "bot"}, filter},
		{[]string{"policies", "list"}, "bot", nil, filter},
		{[]string{"policies", "list"}, "", nil, ""},
		{[]string{"routines", "list"}, "other", []string{"--agent", "bot"}, filter},
		{[]string{"routines", "list"}, "bot", nil, filter},
		{[]string{"routines", "list"}, "", nil, ""},
		{[]string{"glossaries", "list"}, "", []string{"--agent", "bot"}, filter},
		{[]string{"glossaries", "list"}, "bot", nil, filter},
		{[]string{"glossaries", "list"}, "", nil, ""},
		{[]string{"macros", "list"}, "other", []string{"--agent", "bot"}, "agent=bot"},
		{[]string{"macros", "list"}, "bot", nil, "agent=bot"},
		{[]string{"macros", "list"}, "", nil, ""},
		{[]string{"prompts", "list"}, "other", []string{"--agent", "bot"}, "agent=bot"},
		{[]string{"prompts", "list"}, "bot", nil, "agent=bot"},
		{[]string{"prompts", "list"}, "", nil, ""},
		{
			[]string{"mcps", "list"},
			"",
			[]string{"--agent", "bot"},
			`[{"attached_agents":["bot"],"extra":1,"name":"crm"}]`,
		},
		{
			[]string{"mcps", "list"},
			"bot",
			nil,
			`[{"attached_agents":["bot"],"extra":1,"name":"crm"}]`,
		},
		{[]string{"mcps", "list"}, "nobody", nil, `[]`},
		{[]string{"mcps", "list"}, "", nil, `["crm","mail","docs","idle"]`},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("%s selected=%q %s", strings.Join(tt.list, " "), tt.selected, tt.args)
		t.Run(name, func(t *testing.T) {
			saveSelection(t, tt.selected)
			var filters []string
			agentListsServer(t, &filters)
			cmd, _, err := rootCmd.Find(tt.list)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				cmd.Flags().VisitAll(func(f *pflag.Flag) {
					_ = f.Value.Set(f.DefValue)
					f.Changed = false
				})
				cmd.SetOut(nil)
			})
			if err := cmd.ParseFlags(
				append(tt.args, "--json", "-o", "org", "-p", "proj"),
			); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetContext(context.Background())
			if err := cmd.RunE(cmd, nil); err != nil {
				t.Fatal(err)
			}
			got := strings.Join(filters, "")
			if tt.list[0] == "mcps" {
				got = mcpsOutput(t, out.Bytes(), tt.selected != "" || len(tt.args) > 0)
			}
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// mcpsOutput returns the filtered MCPs as printed, or only their names for the unfiltered list.
func mcpsOutput(t *testing.T, out []byte, filtered bool) string {
	t.Helper()
	var body struct {
		Data struct {
			Mcps []map[string]any `json:"mcps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(out, &body); err != nil {
		t.Fatalf("decode %s: %v", out, err)
	}
	var v any = body.Data.Mcps
	if !filtered {
		var names []any
		for _, m := range body.Data.Mcps {
			names = append(names, m["name"])
		}
		v = names
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func TestAgentLogsFilters(t *testing.T) {
	const path = "/v1/organizations/org-1/projects/proj-1/agents/bot/logs"
	const counts = `{"total":3,"debug":0,"info":1,"warn":0,"error":2}`
	tests := []struct {
		args []string
		want string
	}{
		{
			[]string{"--level", "error", "--message", "time|fail"},
			path + "?level=error&message=time%7Cfail",
		},
		{[]string{"--summary"}, path + "/summary? " + counts},
		{
			[]string{"--summary", "--since", "24h", "--message", "fail"},
			path + "/summary?message=fail&since=24h " + counts,
		},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			saveSelection(t, "")
			var requests []string
			agentListsServer(t, &requests)
			t.Cleanup(func() {
				agentLogsCmd.Flags().VisitAll(func(f *pflag.Flag) {
					_ = f.Value.Set(f.DefValue)
					f.Changed = false
				})
				agentLogsCmd.SetOut(nil)
			})
			if err := agentLogsCmd.ParseFlags(
				append(tt.args, "-o", "org", "-p", "proj"),
			); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			agentLogsCmd.SetOut(&out)
			agentLogsCmd.SetContext(context.Background())
			if err := agentLogsCmd.RunE(agentLogsCmd, []string{"bot"}); err != nil {
				t.Fatal(err)
			}
			got := strings.Join(requests, "")
			if strings.Contains(got, "/summary") {
				var compact bytes.Buffer
				_ = json.Compact(&compact, out.Bytes())
				got += " " + compact.String()
			}
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestAgentDrafts(t *testing.T) {
	const wipJSON = `{"name":"wip","projectId":"proj-1","revision":0,"status":"draft",` +
		`"updated":"2026-11-01T00:00:00Z","id":"d-1","version":"0.9","agentConfig":{"model":"m"}}`
	tests := []struct {
		cmd  string
		args []string
		want string
	}{
		{
			"list",
			[]string{"--json"},
			`[{"name":"live","projectId":"","revision":2,"status":"Running"},` +
				`{"name":"both","projectId":"","revision":1,"status":"Running"},` +
				`{"name":"wip","projectId":"proj-1","revision":0,"status":"draft","updated":"2026-11-01T00:00:00Z"}]`,
		},
		{"describe", []string{"--json", "wip"}, wipJSON},
		{
			"describe",
			[]string{"--json", "live"},
			`{"name":"live","projectId":"","revision":2,"status":"Running","id":"","version":"","agentConfig":null}`,
		},
		{"describe", []string{"ghost"}, "error: agent not found"},
		{"select", []string{"wip"}, "Selected agent wip"},
		{"select", []string{"ghost"}, `error: failed to resolve agent "ghost": agent not found`},
	}
	for _, tt := range tests {
		t.Run(tt.cmd+" "+strings.Join(tt.args, " "), func(t *testing.T) {
			saveSelection(t, "")
			agentListsServer(t, new([]string))
			cmd, _, err := rootCmd.Find([]string{"agents", tt.cmd})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				cmd.Flags().VisitAll(func(f *pflag.Flag) {
					_ = f.Value.Set(f.DefValue)
					f.Changed = false
				})
				cmd.SetOut(nil)
			})
			if err := cmd.ParseFlags(append(tt.args, "-o", "org", "-p", "proj")); err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetContext(context.Background())
			err = cmd.RunE(cmd, cmd.Flags().Args())
			var compact bytes.Buffer
			got := strings.TrimSpace(out.String())
			if err != nil {
				got = "error: " + err.Error()
			} else if json.Compact(&compact, out.Bytes()) == nil {
				got = compact.String()
			}
			if got != tt.want {
				t.Errorf("got %s\nwant %s", got, tt.want)
			}
		})
	}
}
