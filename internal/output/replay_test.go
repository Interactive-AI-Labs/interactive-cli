package output

import (
	"bytes"
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

func TestPrintReplayRun(t *testing.T) {
	judge := &agent.Judge{Score: "PASS", Reasoning: "Greets by name."}
	passedIt := agent.Iteration{
		Status: agent.StatusPassed, Turns: 4, SessionKey: "account-lock@r1-1-1",
		TraceIDs: []string{"7c31", "0aa1"}, EvalTraceID: "9b1e",
		Observed: agent.Observed{
			ToolsCalled: []string{
				"crm:lookup_customer",
			},
			ToolsDenied: []string{"billing:issue_refund"},
			Steps: []string{
				"verify_identity",
				"confirm_identity",
			},
			Routines: []string{"Account Access"},
			Policies: []string{"authenticated-greeting"},
		},
		Judge: judge,
	}
	failedIt := agent.Iteration{
		Status: agent.StatusFailed, Turns: 3, EvalTraceID: "fail-eval",
		Observed: agent.Observed{Steps: []string{"verify_identity"}},
		Diverged: []string{"tools:create_jira_ticket"},
		Failures: []string{"steps.reached: 'confirm_identity' not observed"},
	}
	errorIt := agent.Iteration{
		Status: agent.StatusError,
		Error:  "JudgeError: the evaluator returned no verdict",
	}

	tests := []struct {
		name       string
		run        *agent.Run
		wantOut    string
		wantErrOut string
	}{
		{
			name: "single scenario pass",
			run: &agent.Run{
				RunID:    "r1",
				Scenario: "account-lock",
				Status:   agent.StatusPassed,
				Repeat:   1,
				Batches: []agent.Batch{
					{
						Scenario:   "account-lock",
						Status:     agent.StatusPassed,
						Repeat:     1,
						Passed:     1,
						Iterations: []agent.Iteration{passedIt},
					},
				},
			},
			wantOut: "scenario account-lock   repeat 1\n" +
				"\n" +
				"account-lock\n" +
				"--- run 1/1  PASSED ---\n" +
				"  turns          4\n" +
				"  tools called   crm:lookup_customer\n" +
				"  tools denied   billing:issue_refund\n" +
				"  steps          verify_identity, confirm_identity\n" +
				"  routines       Account Access\n" +
				"  policies       authenticated-greeting\n" +
				"  judge          PASS\n" +
				"    Greets by name.\n" +
				"  session        account-lock@r1-1-1\n" +
				"  eval trace     9b1e        turn-1 trace 7c31\n" +
				"\n" +
				"PASS   1/1 passed     run r1\n",
			wantErrOut: "scores: iai scores list --name replay.verdict --columns name,value,trace_id,comment · eval trace: iai traces get 9b1e\n",
		},
		{
			name: "single scenario fail with repeat",
			run: &agent.Run{
				RunID:       "r1",
				Dataset:     "replay-chat",
				Status:      agent.StatusFailed,
				Repeat:      2,
				Concurrency: 8,
				Batches: []agent.Batch{
					{
						Scenario:   "account-lock",
						Status:     agent.StatusFailed,
						Repeat:     2,
						Passed:     1,
						Iterations: []agent.Iteration{passedIt, failedIt},
					},
				},
			},
			wantOut: "dataset replay-chat   1 scenario   repeat 2   concurrency 8\n" +
				"\n" +
				"account-lock\n" +
				"--- run 1/2  PASSED ---\n" +
				"  turns          4\n" +
				"  tools called   crm:lookup_customer\n" +
				"  tools denied   billing:issue_refund\n" +
				"  steps          verify_identity, confirm_identity\n" +
				"  routines       Account Access\n" +
				"  policies       authenticated-greeting\n" +
				"  judge          PASS\n" +
				"    Greets by name.\n" +
				"  session        account-lock@r1-1-1\n" +
				"  eval trace     9b1e        turn-1 trace 7c31\n" +
				"--- run 2/2  FAILED ---\n" +
				"  turns          3\n" +
				"  steps          verify_identity\n" +
				"  eval trace     fail-eval\n" +
				"  DIVERGED       recorded but not replayed: tools:create_jira_ticket\n" +
				"  FAIL           steps.reached: 'confirm_identity' not observed\n" +
				"\n" +
				"FAIL   1/2 passed     run r1\n",
			wantErrOut: "scores: iai scores list --name replay.verdict --columns name,value,trace_id,comment · eval trace: iai traces get fail-eval\n",
		},
		{
			name: "multi scenario expands failed and errored only",
			run: &agent.Run{
				RunID:       "r9",
				Dataset:     "replay-chat",
				Status:      agent.StatusError,
				Repeat:      1,
				Concurrency: 16,
				Skipped:     []agent.Skipped{{ID: "x"}, {ID: "y"}},
				Batches: []agent.Batch{
					{
						Scenario:   "account-lock",
						Status:     agent.StatusPassed,
						Repeat:     1,
						Passed:     1,
						Iterations: []agent.Iteration{passedIt},
					},
					{
						Scenario:   "bonus-misrouted",
						Status:     agent.StatusFailed,
						Repeat:     1,
						Passed:     0,
						Iterations: []agent.Iteration{failedIt},
					},
					{
						Scenario:   "bet-id",
						Status:     agent.StatusError,
						Repeat:     1,
						Passed:     0,
						Iterations: []agent.Iteration{errorIt},
					},
				},
			},
			wantOut: "dataset replay-chat   3 scenarios   repeat 1   concurrency 16   skipped 2\n" +
				"\n" +
				"PASSED  account-lock                     1/1\n" +
				"FAILED  bonus-misrouted                  0/1\n" +
				"--- run 1/1  FAILED ---\n" +
				"  turns          3\n" +
				"  steps          verify_identity\n" +
				"  eval trace     fail-eval\n" +
				"  DIVERGED       recorded but not replayed: tools:create_jira_ticket\n" +
				"  FAIL           steps.reached: 'confirm_identity' not observed\n" +
				"ERROR   bet-id                           0/1\n" +
				"--- run 1/1  ERROR ---\n" +
				"  ERROR          JudgeError: the evaluator returned no verdict\n" +
				"\n" +
				"ERROR  33% of 3 scenarios passed     run r9\n",
			wantErrOut: "scores: iai scores list --name replay.verdict --columns name,value,trace_id,comment · eval trace: iai traces get fail-eval\n",
		},
		{
			name: "batch level error without iterations",
			run: &agent.Run{
				RunID:    "r2",
				Scenario: "x",
				Status:   agent.StatusError,
				Repeat:   1,
				Batches: []agent.Batch{
					{
						Scenario: "x",
						Status:   agent.StatusError,
						Repeat:   1,
						Error:    "unknown context variable",
					},
				},
			},
			wantOut: "scenario x   repeat 1\n" +
				"\n" +
				"x\n" +
				"  ERROR          unknown context variable\n" +
				"\n" +
				"ERROR  0/1 passed     run r2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			if err := PrintReplayRun(&out, &errOut, tt.run); err != nil {
				t.Fatalf("PrintReplayRun() error = %v", err)
			}
			if got := out.String(); got != tt.wantOut {
				t.Errorf("stdout mismatch\ngot:\n%s\nwant:\n%s", got, tt.wantOut)
			}
			if got := errOut.String(); got != tt.wantErrOut {
				t.Errorf("stderr mismatch\ngot:\n%s\nwant:\n%s", got, tt.wantErrOut)
			}
		})
	}
}
