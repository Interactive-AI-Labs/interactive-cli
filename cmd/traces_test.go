package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
	"github.com/google/go-cmp/cmp"
)

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
					 "output":"{\"matches\":[{\"type\":\"routine\",\"routine_id\":\"bonus-chat\",\"condition\":\"c\",\"score\":10}]}"},
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
	want := []summary.TraceSummaryItem{
		{
			TraceID: "t1",
			Summary: &summary.TraceSummaryModel{
				Name:  "turn-t1",
				Level: "DEFAULT",
				Input: `"hi"`,
				Iterations: []summary.Iteration{{
					Number:   1,
					Routines: []string{"bonus-chat"},
					Tools: []summary.ToolCall{{
						Name:   "get_bonus_eligibility",
						Args:   json.RawMessage(`{"party_id":"1"}`),
						Result: json.RawMessage(`{"eligible":true}`),
					}},
				}},
				Reply: "yo",
			},
		},
		{TraceID: "t2", Error: "trace exploded"},
		{
			TraceID: "t3",
			Summary: &summary.TraceSummaryModel{
				Name: "turn-t3", Level: "DEFAULT", Input: `"hi"`, Reply: "yo",
			},
		},
	}
	canonicalJSON := cmp.Transformer("CanonicalJSON", func(raw json.RawMessage) string {
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			return string(raw)
		}
		canonical, _ := json.Marshal(value)
		return string(canonical)
	})
	if diff := cmp.Diff(want, items, canonicalJSON); diff != "" {
		t.Errorf("summary items mismatch (-want +got):\n%s", diff)
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

	want := `Trace t1
Turn — turn-t1 · 1 iteration

Customer:
  "hi"

Iteration 1
  Routines: bonus-chat
  Tools called:
    → get_bonus_eligibility(party_id="1")
      Result:
        {"eligible":true}

Agent:
  yo

────────────────────────────────────────────────────────────────────────

Trace t2
  (summary unavailable: trace exploded)

────────────────────────────────────────────────────────────────────────

Trace t3
Turn — turn-t3 · 0 iterations

Customer:
  "hi"

Agent:
  yo

Page 1 of 1 (3 total items)
`
	if diff := cmp.Diff(want, stdout.String()); diff != "" {
		t.Errorf("human summary mismatch (-want +got):\n%s", diff)
	}
}

func TestTraceSummariesForReturnsCancellation(t *testing.T) {
	apiClient, err := api.NewAPIClient("http://invalid", time.Second, "token", "", nil)
	if err != nil {
		t.Fatalf("new API client: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	traces := make([]api.TraceInfo, 100)
	for i := range traces {
		traces[i].ID = fmt.Sprintf("t%d", i)
	}
	items, err := traceSummariesFor(ctx, apiClient, "org-1", "proj-1", traces)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("traceSummariesFor error = %v, want context.Canceled", err)
	}
	if items != nil {
		t.Fatalf("traceSummariesFor items = %v, want nil on cancellation", items)
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
