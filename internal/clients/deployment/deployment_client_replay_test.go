package deployment

import "testing"

func TestDecodeReplayError(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantDetail  string
		wantSkipped int
	}{
		{
			name:        "flat detail with skipped items",
			status:      400,
			body:        `{"detail":"dataset 'x' holds no replayable scenario","skipped":[{"id":"p","reason":"invalid"}]}`,
			wantDetail:  "dataset 'x' holds no replayable scenario",
			wantSkipped: 1,
		},
		{
			name:       "422 list keeps the last loc element",
			status:     422,
			body:       `{"detail":[{"loc":["body","repeat"],"msg":"Input should be less than or equal to 20","type":"x"}]}`,
			wantDetail: "repeat: Input should be less than or equal to 20",
		},
		{
			name:       "the platform's own refusal carries message",
			status:     403,
			body:       `{"code":403,"message":"Not allowed to replay this agent"}`,
			wantDetail: "Not allowed to replay this agent",
		},
		{name: "plain text", status: 401, body: "Unauthorized", wantDetail: "Unauthorized"},
		{
			name:       "non-JSON html",
			status:     403,
			body:       "<html>blocked</html>",
			wantDetail: "<html>blocked</html>",
		},
		{name: "empty body", status: 502, body: "", wantDetail: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := decodeReplayError(tt.status, []byte(tt.body))
			if e.Status != tt.status || e.Detail != tt.wantDetail ||
				len(e.Skipped) != tt.wantSkipped {
				t.Errorf("got %+v", e)
			}
		})
	}
}

func TestDecodeReplayRun(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantBatches  int
		wantScenario string
		wantErr      bool
	}{
		{
			name:         "dataset run keeps its batches",
			body:         `{"run_id":"r","status":"running","batches":[{"scenario":"a"},{"scenario":"b"}]}`,
			wantBatches:  2,
			wantScenario: "a",
		},
		{
			name:         "inline run becomes a suite of one batch",
			body:         `{"run_id":"r","scenario":"solo","status":"passed","iterations":[{"status":"passed"}]}`,
			wantBatches:  1,
			wantScenario: "solo",
		},
		{name: "not json", body: `<html>`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run, err := decodeReplayRun([]byte(tt.body))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(run.Batches) != tt.wantBatches || run.Batches[0].Scenario != tt.wantScenario {
				t.Errorf("batches = %+v", run.Batches)
			}
			if string(run.Raw) != tt.body {
				t.Errorf("Raw not kept verbatim")
			}
		})
	}
}
