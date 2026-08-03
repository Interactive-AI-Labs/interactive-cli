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

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

// tracesListSummaryServer serves a three-trace page where t2 always fails,
// exercising order preservation and per-item errors in the batch summary.
func tracesListSummaryServer(t *testing.T) *httptest.Server {
	t.Helper()
	base := "/api/platform/v1/organizations/org-1/projects/proj-1"
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/api/v1/session/organizations":
				fmt.Fprint(w, `{"organizations":[{"id":"org-1","name":"acme"}]}`)
			case r.URL.Path == "/api/v1/session/organizations/org-1/projects":
				fmt.Fprint(w, `{"projects":[{"id":"proj-1","name":"alunafi"}]}`)
			case r.URL.Path == base+"/traces":
				fmt.Fprint(w, `{"success":true,"data":{
					"traces":[{"id":"t1","name":"turn-1"},{"id":"t2","name":"turn-2"},{"id":"t3","name":"turn-3"}],
					"meta":{"page":1,"limit":100,"total_items":3,"total_pages":1}}}`)
			case r.URL.Path == base+"/traces/t2":
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, `{"success":false,"message":"trace exploded"}`)
			case r.URL.Path == base+"/traces/t1" || r.URL.Path == base+"/traces/t3":
				id := strings.TrimPrefix(r.URL.Path, base+"/traces/")
				fmt.Fprintf(w, `{"success":true,"data":{"trace":{
					"id":%[1]q,"name":"turn-%[1]s","level":"DEFAULT",
					"input":"\"hi\"","output":"[\"yo\"]"}}}`, id)
			case r.URL.Path == base+"/traces/t1/observations":
				fmt.Fprint(w, `{"success":true,"data":{"observations":[
					{"id":"it1","parent_observation_id":"","type":"CHAIN","name":"Iteration: 1"},
					{"id":"ec1","parent_observation_id":"it1","type":"CHAIN","name":"Evaluate: Context",
					 "output":{"matches":[{"type":"routine","routine_id":"bonus-chat","condition":"c","score":10}]}},
					{"id":"ex1","parent_observation_id":"it1","type":"CHAIN","name":"Execute: Tools"},
					{"id":"tc1","parent_observation_id":"ex1","type":"TOOL","name":"get_bonus_eligibility",
					 "input":{"party_id":"1"},"output":{"eligible":true}}
				]}}`)
			case r.URL.Path == base+"/traces/t3/observations":
				fmt.Fprint(w, `{"success":true,"data":{"observations":[]}}`)
			default:
				t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)
	t.Cleanup(server.Close)
	return server
}

func setupTracesListSummaryTest(t *testing.T, server *httptest.Server) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	origHostname, origToken, origApiKey := hostname, token, apiKey
	t.Cleanup(func() {
		hostname, token, apiKey = origHostname, origToken, origApiKey
		tracesListOrg, tracesListProject = "", ""
		tracesListSummary, tracesJSON = false, false
		for _, name := range []string{"summary", "json"} {
			tracesListCmd.Flags().Lookup(name).Changed = false
		}
		tracesListCmd.SetOut(nil)
		tracesListCmd.SetErr(nil)
	})
	hostname, token, apiKey = server.URL, "test-token", ""
	tracesListOrg, tracesListProject = "acme", "alunafi"
}

func TestTracesListSummaryJSON(t *testing.T) {
	server := tracesListSummaryServer(t)
	setupTracesListSummaryTest(t, server)

	if err := tracesListCmd.Flags().Set("summary", "true"); err != nil {
		t.Fatalf("set --summary: %v", err)
	}
	if err := tracesListCmd.Flags().Set("json", "true"); err != nil {
		t.Fatalf("set --json: %v", err)
	}

	var stdout bytes.Buffer
	tracesListCmd.SetOut(&stdout)
	tracesListCmd.SetContext(context.Background())
	if err := tracesListCmd.RunE(tracesListCmd, nil); err != nil {
		t.Fatalf("traces list --summary --json: %v", err)
	}

	var items []summary.TraceSummaryItem
	if err := json.Unmarshal(stdout.Bytes(), &items); err != nil {
		t.Fatalf("output is not a summary item array: %v\n%s", err, stdout.String())
	}
	if len(items) != 3 {
		t.Fatalf("items = %d, want 3", len(items))
	}
	for i, wantID := range []string{"t1", "t2", "t3"} {
		if items[i].TraceID != wantID {
			t.Errorf("items[%d].TraceID = %q, want %q (order must follow the list)",
				i, items[i].TraceID, wantID)
		}
	}
	if items[0].Summary == nil || len(items[0].Summary.Iterations) != 1 {
		t.Fatalf("t1 summary missing display-name iteration: %+v", items[0].Summary)
	}
	tools := items[0].Summary.Iterations[0].Tools
	if len(tools) != 1 || tools[0].Name != "get_bonus_eligibility" {
		t.Errorf("t1 tools = %+v, want get_bonus_eligibility", tools)
	}
	if items[1].Summary != nil || !strings.Contains(items[1].Error, "trace exploded") {
		t.Errorf("t2 = %+v, want error item carrying the server message", items[1])
	}
	if items[2].Summary == nil || items[2].Error != "" {
		t.Errorf("t3 = %+v, want summary despite t2 failing", items[2])
	}
}

func TestTracesListSummaryHuman(t *testing.T) {
	server := tracesListSummaryServer(t)
	setupTracesListSummaryTest(t, server)

	if err := tracesListCmd.Flags().Set("summary", "true"); err != nil {
		t.Fatalf("set --summary: %v", err)
	}

	var stdout bytes.Buffer
	tracesListCmd.SetOut(&stdout)
	tracesListCmd.SetContext(context.Background())
	if err := tracesListCmd.RunE(tracesListCmd, nil); err != nil {
		t.Fatalf("traces list --summary: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"Trace t1",
		"→ get_bonus_eligibility",
		"Trace t2",
		"summary unavailable: trace exploded",
		"Trace t3",
		"Page 1 of 1 (3 total items)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("human output missing %q:\n%s", want, got)
		}
	}
}

func TestTracesListSummaryFlagExclusions(t *testing.T) {
	for _, other := range []string{"columns", "fields"} {
		t.Run("summary with "+other, func(t *testing.T) {
			mark := func(name string) { tracesListCmd.Flags().Lookup(name).Changed = true }
			t.Cleanup(func() {
				for _, name := range []string{"summary", other} {
					tracesListCmd.Flags().Lookup(name).Changed = false
				}
			})
			mark("summary")
			mark(other)
			if err := tracesListCmd.ValidateFlagGroups(); err == nil {
				t.Fatalf("--summary with --%s should be rejected", other)
			}
		})
	}
}
