package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func TestPrintReplaySkipped(t *testing.T) {
	five := []deployment.ReplaySkipped{
		{ID: "a", Reason: "r"},
		{ID: "b", Reason: "r"},
		{ID: "c", Reason: "r"},
		{ID: "d", Reason: "r"},
		{ID: "e", Reason: "r"},
	}
	tests := []struct {
		name      string
		skipped   []deployment.ReplaySkipped
		requested []string
		want      string
	}{
		{name: "none", want: ""},
		{
			name: "listed when few",
			skipped: []deployment.ReplaySkipped{
				{ID: "probe-1", Reason: "messages: Field required"},
			},
			want: "skipped probe-1: messages: Field required\n",
		},
		{
			name:    "collapsed past five",
			skipped: append(five, deployment.ReplaySkipped{ID: "f", Reason: "r"}),
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
	passedIt = deployment.ReplayIteration{
		Status: deployment.ReplayStatusPassed, Turns: 4, SessionKey: "account-lock@r1-1-1",
		TraceIDs: []string{"7c31", "0aa1"}, EvalTraceID: "9b1e",
		Observed: deployment.ReplayObserved{
			ToolsCalled: []string{"crm:lookup_customer"},
			ToolsDenied: []string{"billing:issue_refund"},
			Steps:       []string{"verify_identity", "confirm_identity"},
			Routines:    []string{"Account Access"},
			Policies:    []string{"authenticated-greeting"},
		},
		Judge: &deployment.ReplayJudge{Score: "PASS", Reasoning: "Greets by name."},
	}
	failedIt = deployment.ReplayIteration{
		Status: deployment.ReplayStatusFailed, Turns: 3, EvalTraceID: "fail-eval",
		Observed: deployment.ReplayObserved{Steps: []string{"verify_identity"}},
		Diverged: []string{"tools:create_jira_ticket"},
		Failures: []string{"steps.reached: 'confirm_identity' not observed"},
	}
	errorIt = deployment.ReplayIteration{
		Status: deployment.ReplayStatusError,
		Error:  "JudgeError: the evaluator returned no verdict",
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
		run       *deployment.ReplayRun
		requested []string
		want      string
	}{
		{
			name: "single scenario pass",
			run: &deployment.ReplayRun{
				RunID:    "r1",
				Scenario: "account-lock",
				Status:   deployment.ReplayStatusPassed,
				Repeat:   1,
				Batches: []deployment.ReplayBatch{
					{
						Scenario:   "account-lock",
						Status:     deployment.ReplayStatusPassed,
						Repeat:     1,
						Passed:     1,
						Iterations: []deployment.ReplayIteration{passedIt},
					},
				},
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
			run: &deployment.ReplayRun{
				RunID:       "r1",
				Dataset:     "replay-chat",
				Status:      deployment.ReplayStatusFailed,
				Repeat:      2,
				Concurrency: 8,
				Batches: []deployment.ReplayBatch{
					{
						Scenario:   "account-lock",
						Status:     deployment.ReplayStatusFailed,
						Repeat:     2,
						Passed:     1,
						Iterations: []deployment.ReplayIteration{passedIt, failedIt},
					},
				},
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
			run: &deployment.ReplayRun{
				RunID:       "r9",
				Dataset:     "replay-chat",
				Status:      deployment.ReplayStatusError,
				Repeat:      1,
				Concurrency: 16,
				Skipped:     []deployment.ReplaySkipped{{ID: "x"}, {ID: "y"}},
				Batches: []deployment.ReplayBatch{
					{
						Scenario:   "account-lock",
						Status:     deployment.ReplayStatusPassed,
						Repeat:     1,
						Passed:     1,
						Iterations: []deployment.ReplayIteration{passedIt},
					},
					{
						Scenario:   "bonus-misrouted",
						Status:     deployment.ReplayStatusFailed,
						Repeat:     1,
						Iterations: []deployment.ReplayIteration{failedIt},
					},
					{
						Scenario: "bet-id", Status: deployment.ReplayStatusError, Repeat: 1,
						Iterations: []deployment.ReplayIteration{errorIt},
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
			run: &deployment.ReplayRun{
				RunID:   "r1",
				Dataset: "replay-email",
				Status:  deployment.ReplayStatusPassed,
				Repeat:  1,
				Skipped: []deployment.ReplaySkipped{{ID: "archived-1"}, {ID: "archived-2"}},
				Batches: []deployment.ReplayBatch{
					{
						Scenario:   "account-lock",
						Status:     deployment.ReplayStatusPassed,
						Repeat:     1,
						Passed:     1,
						Iterations: []deployment.ReplayIteration{passedIt},
					},
				},
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
			run: &deployment.ReplayRun{
				RunID: "r2", Scenario: "x", Status: deployment.ReplayStatusError, Repeat: 1,
				Batches: []deployment.ReplayBatch{{
					Scenario: "x", Status: deployment.ReplayStatusError, Repeat: 1,
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
			run: &deployment.ReplayRun{
				RunID:       "873c",
				Scenario:    "inline-bad-customer",
				Status:      deployment.ReplayStatusError,
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
			run: &deployment.ReplayRun{
				RunID:   "r4",
				Dataset: "replay-chat",
				Status:  deployment.ReplayStatusError,
				Repeat:  1,
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
		run  *deployment.ReplayRun
		want string
	}{
		{
			name: "single passed iteration",
			run: &deployment.ReplayRun{Batches: []deployment.ReplayBatch{{
				Scenario: "account-lock", Iterations: []deployment.ReplayIteration{passedIt},
			}}},
			want: prefix + "(account-lock): iai traces get 9b1e\n",
		},
		{
			name: "first failing iteration wins and names its scenario",
			run: &deployment.ReplayRun{Batches: []deployment.ReplayBatch{
				{Scenario: "account-lock", Iterations: []deployment.ReplayIteration{passedIt}},
				{Scenario: "bonus-misrouted", Iterations: []deployment.ReplayIteration{failedIt}},
			}},
			want: prefix + "(bonus-misrouted): iai traces get fail-eval\n",
		},
		{
			name: "no eval traces prints nothing",
			run: &deployment.ReplayRun{
				Batches: []deployment.ReplayBatch{
					{Scenario: "x", Iterations: []deployment.ReplayIteration{errorIt}},
				},
			},
			want: "",
		},
		{name: "no batches prints nothing", run: &deployment.ReplayRun{}, want: ""},
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
