package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/agent"
)

func TestPrintReplaySkipped(t *testing.T) {
	five := []agent.Skipped{
		{ID: "a", Reason: "r"},
		{ID: "b", Reason: "r"},
		{ID: "c", Reason: "r"},
		{ID: "d", Reason: "r"},
		{ID: "e", Reason: "r"},
	}
	tests := []struct {
		name      string
		skipped   []agent.Skipped
		requested []string
		want      string
	}{
		{name: "none", want: ""},
		{
			name:    "listed when few",
			skipped: []agent.Skipped{{ID: "probe-1", Reason: "messages: Field required"}},
			want:    "skipped probe-1: messages: Field required\n",
		},
		{
			name:    "collapsed past five",
			skipped: append(five, agent.Skipped{ID: "f", Reason: "r"}),
			want:    "skipped 6 dataset items (see --json)\n",
		},
		{
			name:      "filtered to requested",
			skipped:   five,
			requested: []string{"b", "zzz"},
			want:      "skipped b: r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintReplaySkipped(&buf, tt.skipped, tt.requested)
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

var (
	passedIt = agent.Iteration{
		Status: agent.StatusPassed, Turns: 4, SessionKey: "account-lock@r1-1-1",
		TraceIDs: []string{"7c31", "0aa1"}, EvalTraceID: "9b1e",
		Observed: agent.Observed{
			ToolsCalled: []string{"crm:lookup_customer"},
			ToolsDenied: []string{"billing:issue_refund"},
			Steps:       []string{"verify_identity", "confirm_identity"},
			Routines:    []string{"Account Access"},
			Policies:    []string{"authenticated-greeting"},
		},
		Judge: &agent.Judge{Score: "PASS", Reasoning: "Greets by name."},
	}
	failedIt = agent.Iteration{
		Status: agent.StatusFailed, Turns: 3, EvalTraceID: "fail-eval",
		Observed: agent.Observed{Steps: []string{"verify_identity"}},
		Diverged: []string{"tools:create_jira_ticket"},
		Failures: []string{"steps.reached: 'confirm_identity' not observed"},
	}
	errorIt = agent.Iteration{
		Status: agent.StatusError, Error: "JudgeError: the evaluator returned no verdict",
	}
	passedBlock = "--- run 1/1  PASSED ---\n" +
		"  turns          4\n" +
		"  tools called   crm:lookup_customer\n" +
		"  tools denied   billing:issue_refund\n" +
		"  steps          verify_identity, confirm_identity\n" +
		"  routines       Account Access\n" +
		"  policies       authenticated-greeting\n" +
		"  judge          PASS\n" +
		"    Greets by name.\n" +
		"  session        account-lock@r1-1-1\n" +
		"  eval trace     9b1e        turn-1 trace 7c31\n"
	failedBlock = "--- run 1/1  FAILED ---\n" +
		"  turns          3\n" +
		"  steps          verify_identity\n" +
		"  eval trace     fail-eval\n" +
		"  DIVERGED       recorded but not replayed: tools:create_jira_ticket\n" +
		"  FAIL           steps.reached: 'confirm_identity' not observed\n"
)

func TestPrintReplayRun(t *testing.T) {
	tests := []struct {
		name      string
		run       *agent.Run
		requested []string
		want      string
	}{
		{
			name: "single scenario pass",
			run: &agent.Run{
				RunID: "r1", Scenario: "account-lock", Status: agent.StatusPassed, Repeat: 1,
				Batches: []agent.Batch{{
					Scenario: "account-lock", Status: agent.StatusPassed, Repeat: 1, Passed: 1,
					Iterations: []agent.Iteration{passedIt},
				}},
			},
			want: "scenario account-lock   repeat 1\n" +
				"\n" +
				"account-lock\n" +
				passedBlock +
				"\n" +
				"PASS   1/1 passed     run r1\n",
		},
		{
			name: "single scenario fail with repeat",
			run: &agent.Run{
				RunID: "r1", Dataset: "replay-chat", Status: agent.StatusFailed, Repeat: 2,
				Concurrency: 8,
				Batches: []agent.Batch{{
					Scenario: "account-lock", Status: agent.StatusFailed, Repeat: 2, Passed: 1,
					Iterations: []agent.Iteration{passedIt, failedIt},
				}},
			},
			want: "dataset replay-chat   1 scenario   repeat 2   concurrency 8\n" +
				"\n" +
				"account-lock\n" +
				strings.Replace(passedBlock, "run 1/1", "run 1/2", 1) +
				strings.Replace(failedBlock, "run 1/1", "run 2/2", 1) +
				"\n" +
				"FAIL   1/2 passed     run r1\n",
		},
		{
			name: "multi scenario expands failed and errored, pads to longest name",
			run: &agent.Run{
				RunID: "r9", Dataset: "replay-chat", Status: agent.StatusError, Repeat: 1,
				Concurrency: 16,
				Skipped:     []agent.Skipped{{ID: "x"}, {ID: "y"}},
				Batches: []agent.Batch{
					{
						Scenario: "account-lock", Status: agent.StatusPassed, Repeat: 1, Passed: 1,
						Iterations: []agent.Iteration{passedIt},
					},
					{
						Scenario: "bonus-misrouted", Status: agent.StatusFailed, Repeat: 1,
						Iterations: []agent.Iteration{failedIt},
					},
					{
						Scenario: "bet-id", Status: agent.StatusError, Repeat: 1,
						Iterations: []agent.Iteration{errorIt},
					},
				},
			},
			want: "dataset replay-chat   3 scenarios   repeat 1   concurrency 16   skipped 2\n" +
				"\n" +
				"PASSED  account-lock    1/1\n" +
				"FAILED  bonus-misrouted 0/1\n" +
				failedBlock +
				"ERROR   bet-id          0/1\n" +
				"--- run 1/1  ERROR ---\n" +
				"  ERROR          JudgeError: the evaluator returned no verdict\n" +
				"\n" +
				"ERROR  33% of 3 scenarios passed     run r9\n",
		},
		{
			name: "skipped count filtered to requested scenarios",
			run: &agent.Run{
				RunID: "r1", Dataset: "replay-email", Status: agent.StatusPassed, Repeat: 1,
				Skipped: []agent.Skipped{{ID: "archived-1"}, {ID: "archived-2"}},
				Batches: []agent.Batch{{
					Scenario: "account-lock", Status: agent.StatusPassed, Repeat: 1, Passed: 1,
					Iterations: []agent.Iteration{passedIt},
				}},
			},
			requested: []string{"account-lock"},
			want: "dataset replay-email   1 scenario   repeat 1\n" +
				"\n" +
				"account-lock\n" +
				passedBlock +
				"\n" +
				"PASS   1/1 passed     run r1\n",
		},
		{
			name: "batch level error without iterations",
			run: &agent.Run{
				RunID: "r2", Scenario: "x", Status: agent.StatusError, Repeat: 1,
				Batches: []agent.Batch{{
					Scenario: "x", Status: agent.StatusError, Repeat: 1,
					Error: "unknown context variable",
				}},
			},
			want: "scenario x   repeat 1\n" +
				"\n" +
				"x\n" +
				"  ERROR          unknown context variable\n" +
				"\n" +
				"ERROR  0/1 passed     run r2\n",
		},
		{
			name: "run level error with no scenarios",
			run: &agent.Run{
				RunID:       "873c",
				Scenario:    "inline-bad-customer",
				Status:      agent.StatusError,
				Repeat:      1,
				Concurrency: 8,
				Error:       "scenario 'inline-bad-customer' is invalid — customer_id: must start with 'replay-'",
			},
			want: "scenario inline-bad-customer   repeat 1   concurrency 8\n" +
				"\n" +
				"  ERROR          scenario 'inline-bad-customer' is invalid — customer_id: must start with 'replay-'\n" +
				"\n" +
				"ERROR  run 873c\n",
		},
		{
			name: "run level error with no text",
			run: &agent.Run{
				RunID: "r4", Dataset: "replay-chat", Status: agent.StatusError, Repeat: 1,
			},
			want: "dataset replay-chat   repeat 1\n" +
				"\n" +
				"  ERROR          the run produced no scenarios\n" +
				"\n" +
				"ERROR  run r4\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := PrintReplayRun(&out, tt.run, tt.requested); err != nil {
				t.Fatalf("PrintReplayRun() error = %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("mismatch\ngot:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestPrintReplayPointer(t *testing.T) {
	const prefix = "scores: iai scores list --name replay.verdict --columns name,trace_id,comment" +
		" · eval trace "
	tests := []struct {
		name string
		run  *agent.Run
		want string
	}{
		{
			name: "single passed iteration",
			run: &agent.Run{Batches: []agent.Batch{{
				Scenario: "account-lock", Iterations: []agent.Iteration{passedIt},
			}}},
			want: prefix + "(account-lock): iai traces get 9b1e\n",
		},
		{
			name: "first failing iteration wins and names its scenario",
			run: &agent.Run{Batches: []agent.Batch{
				{Scenario: "account-lock", Iterations: []agent.Iteration{passedIt}},
				{Scenario: "bonus-misrouted", Iterations: []agent.Iteration{failedIt}},
			}},
			want: prefix + "(bonus-misrouted): iai traces get fail-eval\n",
		},
		{
			name: "no eval traces prints nothing",
			run: &agent.Run{
				Batches: []agent.Batch{{Scenario: "x", Iterations: []agent.Iteration{errorIt}}},
			},
			want: "",
		},
		{name: "no batches prints nothing", run: &agent.Run{}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintReplayPointer(&buf, tt.run)
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintReplayRetry(t *testing.T) {
	var buf bytes.Buffer
	PrintReplayRetry(&buf, &agent.Error{Status: 404, Detail: "No such replay run: r1"}, "5m0s")
	want := "agent returned 404: No such replay run: r1; retrying for up to 5m0s\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
