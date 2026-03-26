package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestValidateScoreColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		wantErr bool
	}{
		{"default columns", DefaultScoreColumns, false},
		{"all columns", AllScoreColumns, false},
		{"unknown", []string{"id", "unknown"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColumns(tt.columns, AllScoreColumns)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScoreColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrepareScoreListOptions(t *testing.T) {
	minValue := 2.0
	maxValue := 1.0

	tests := []struct {
		name    string
		opts    clients.ScoreListOptions
		wantErr bool
		wantOp  string
	}{
		{
			name:    "defaults operator for value",
			opts:    clients.ScoreListOptions{FromTimestamp: "2025-01-01T00:00:00Z", Value: "1"},
			wantErr: false,
			wantOp:  "=",
		},
		{
			name:    "rejects mixed exact and range filters",
			opts:    clients.ScoreListOptions{Value: "1", MinValue: &minValue},
			wantErr: true,
			wantOp:  "=",
		},
		{
			name:    "rejects invalid range",
			opts:    clients.ScoreListOptions{MinValue: &minValue, MaxValue: &maxValue},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrepareScoreListOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PrepareScoreListOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got.Operator != tt.wantOp {
				t.Errorf(
					"PrepareScoreListOptions() operator = %q, want %q",
					got.Operator,
					tt.wantOp,
				)
			}
		})
	}
}

func TestBuildScoreCreateBody(t *testing.T) {
	tests := []struct {
		name      string
		input     ScoreCreateInput
		wantErr   bool
		wantValue any
	}{
		{
			name: "numeric score",
			input: ScoreCreateInput{
				Name:    "quality",
				TraceID: "trace-1",
				Value:   "0.95",
			},
			wantErr:   false,
			wantValue: 0.95,
		},
		{
			name: "boolean true coercion",
			input: ScoreCreateInput{
				Name:     "passed",
				TraceID:  "trace-1",
				Value:    "yes",
				DataType: "BOOLEAN",
			},
			wantErr:   false,
			wantValue: 1,
		},
		{
			name: "categorical keeps string",
			input: ScoreCreateInput{
				Name:      "label",
				SessionID: "sess-1",
				Value:     "good",
				DataType:  "CATEGORICAL",
			},
			wantErr:   false,
			wantValue: "good",
		},
		{
			name: "rejects multiple targets",
			input: ScoreCreateInput{
				Name:          "quality",
				TraceID:       "trace-1",
				ObservationID: "obs-1",
				Value:         "1",
			},
			wantErr: true,
		},
		{
			name: "rejects invalid metadata",
			input: ScoreCreateInput{
				Name:         "quality",
				TraceID:      "trace-1",
				Value:        "1",
				MetadataJSON: `["not","object"]`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := BuildScoreCreateBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildScoreCreateBody() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if body.Value != tt.wantValue {
				t.Errorf("BuildScoreCreateBody() value = %#v, want %#v", body.Value, tt.wantValue)
			}
		})
	}
}
