package deployment

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeReplayError(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   *ReplayError
	}{
		{
			name:   "flat detail with skipped items",
			status: 400,
			body: `{"detail":"dataset 'x' holds no replayable scenario",` +
				`"skipped":[{"id":"probe-1","reason":"messages: Field required"},` +
				`{"id":"probe-2","reason":"archived"}]}`,
			want: &ReplayError{
				Status: 400,
				Detail: "dataset 'x' holds no replayable scenario",
				Skipped: []ReplaySkipped{
					{ID: "probe-1", Reason: "messages: Field required"},
					{ID: "probe-2", Reason: "archived"},
				},
			},
		},
		{
			name:   "422 list keeps the last loc element",
			status: 422,
			body: `{"detail":[{"loc":["body","repeat"],` +
				`"msg":"Input should be less than or equal to 20","type":"x"}]}`,
			want: &ReplayError{
				Status: 422,
				Detail: "repeat: Input should be less than or equal to 20",
			},
		},
		{
			name:   "422 list joins every message",
			status: 422,
			body: `{"detail":[{"loc":["body","repeat"],"msg":"too high"},` +
				`{"loc":["body","concurrency"],"msg":"too low"}]}`,
			want: &ReplayError{Status: 422, Detail: "repeat: too high; concurrency: too low"},
		},
		{
			name:   "the platform's own refusal carries message",
			status: 403,
			body:   `{"code":403,"message":"Not allowed to replay this agent"}`,
			want:   &ReplayError{Status: 403, Detail: "Not allowed to replay this agent"},
		},
		{
			name:   "plain text",
			status: 401,
			body:   "Unauthorized",
			want:   &ReplayError{Status: 401, Detail: "Unauthorized"},
		},
		{
			name:   "non-JSON html",
			status: 403,
			body:   "<html>blocked</html>",
			want:   &ReplayError{Status: 403, Detail: "<html>blocked</html>"},
		},
		{name: "empty body", status: 502, body: "", want: &ReplayError{Status: 502}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeReplayError(tt.status, []byte(tt.body))
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("decodeReplayError() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDecodeReplayRun(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    *ReplayRun
		wantErr bool
	}{
		{
			name: "dataset run keeps its batches",
			body: `{"run_id":"r","dataset":"replay-chat","status":"running","repeat":2,` +
				`"concurrency":8,"skipped":[{"id":"archived-1","reason":"archived"}],` +
				`"batches":[{"run_id":"r","scenario":"a","status":"passed","repeat":2,"passed":2,` +
				`"iterations":[{"status":"passed","turns":3,"eval_trace_id":"e1"}]},` +
				`{"run_id":"r","scenario":"b","status":"running","repeat":2}]}`,
			want: &ReplayRun{
				RunID: "r", Dataset: "replay-chat", Status: ReplayStatusRunning,
				Repeat: 2, Concurrency: 8,
				Skipped: []ReplaySkipped{{ID: "archived-1", Reason: "archived"}},
				Batches: []ReplayBatch{
					{
						RunID: "r", Scenario: "a", Status: ReplayStatusPassed, Repeat: 2, Passed: 2,
						Iterations: []ReplayIteration{
							{Status: ReplayStatusPassed, Turns: 3, EvalTraceID: "e1"},
						},
					},
					{RunID: "r", Scenario: "b", Status: ReplayStatusRunning, Repeat: 2},
				},
			},
		},
		{
			name: "inline run becomes a suite of one batch",
			body: `{"run_id":"r","scenario":"solo","status":"failed","repeat":1,"passed":0,` +
				`"error":"judge failed","iterations":[{"status":"failed","turns":2,` +
				`"session_key":"solo@r-1","failures":["no greeting"]}]}`,
			want: &ReplayRun{
				RunID: "r", Scenario: "solo", Status: ReplayStatusFailed, Repeat: 1,
				Error: "judge failed",
				Batches: []ReplayBatch{{
					RunID: "r", Scenario: "solo", Status: ReplayStatusFailed, Repeat: 1,
					Error: "judge failed",
					Iterations: []ReplayIteration{{
						Status: ReplayStatusFailed, Turns: 2, SessionKey: "solo@r-1",
						Failures: []string{"no greeting"},
					}},
				}},
			},
		},
		{
			name: "a run with neither batches nor iterations is left alone",
			body: `{"run_id":"r","status":"error","error":"scenario is invalid"}`,
			want: &ReplayRun{RunID: "r", Status: ReplayStatusError, Error: "scenario is invalid"},
		},
		{name: "not json", body: `<html>`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeReplayRun([]byte(tt.body))
			if tt.wantErr {
				if err == nil {
					t.Fatal("decodeReplayRun() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("decodeReplayRun() unexpected error = %v", err)
			}
			// Raw is the body verbatim, for --json.
			tt.want.Raw = []byte(tt.body)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("decodeReplayRun() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
