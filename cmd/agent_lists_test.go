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

func agentListsServer(t *testing.T, filters *[]string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/session/organizations":
			fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"org"}]}`)
		case r.URL.Path == "/api/v1/session/organizations/org-1/projects":
			fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"proj"}]}`)
		case strings.HasPrefix(r.URL.Path, "/api/platform/v1/projects/proj-1/prompts/"):
			*filters = append(*filters, r.URL.Query().Get("filter"))
			fmt.Fprint(w, `{"success":true,"data":{"prompts":[],"totalCount":0}}`)
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
	origHostname, origToken, origAPIKey := hostname, token, apiKey
	t.Cleanup(func() { hostname, token, apiKey = origHostname, origToken, origAPIKey })
	hostname, token, apiKey = server.URL, "test-token", ""
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
