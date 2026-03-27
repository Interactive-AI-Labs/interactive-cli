package output

import (
	"bytes"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

func TestPrintScoreList(t *testing.T) {
	tests := []struct {
		name    string
		scores  []clients.ScoreInfo
		meta    clients.CursorMeta
		columns []string
		want    string
	}{
		{
			name:    "empty list",
			columns: inputs.DefaultScoreColumns,
			want:    "No scores found.\n",
		},
		{
			name: "default columns with cursor",
			scores: []clients.ScoreInfo{{
				ID:        "score-1",
				Name:      "quality",
				DataType:  "NUMERIC",
				Value:     []byte(`0.95`),
				Source:    "HUMAN",
				Timestamp: "2025-01-01",
				TraceID:   "trace-1",
			}},
			meta:    clients.CursorMeta{NextCursor: "cursor-2"},
			columns: inputs.DefaultScoreColumns,
			want: "ID        NAME      DATA TYPE   VALUE   SOURCE   TIMESTAMP    TRACE ID\n" +
				"score-1   quality   NUMERIC     0.95    HUMAN    2025-01-01   trace-1\n" +
				"\nNext cursor: cursor-2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintScoreList(&buf, tt.scores, tt.meta, tt.columns)
			if err != nil {
				t.Fatalf("PrintScoreList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintScoreCreateResultAndDeleteSuccess(t *testing.T) {
	score := &clients.ScoreInfo{
		ID:        "score-1",
		Name:      "quality",
		DataType:  "NUMERIC",
		Value:     []byte(`0.95`),
		Timestamp: "2025-01-01",
		TraceID:   "trace-1",
	}

	var buf bytes.Buffer
	if err := PrintScoreCreateResult(&buf, score); err != nil {
		t.Fatalf("PrintScoreCreateResult() error = %v", err)
	}

	wantCreate := "Created score \"score-1\".\n" +
		"Name:       quality\n" +
		"Data Type:  NUMERIC\n" +
		"Value:      0.95\n" +
		"Timestamp:  2025-01-01\n" +
		"Trace ID:   trace-1\n"
	if got := buf.String(); got != wantCreate {
		t.Errorf("create output mismatch\ngot:\n%q\nwant:\n%q", got, wantCreate)
	}

	buf.Reset()
	if err := PrintDeleteSuccess(&buf, "score-1", "score", ""); err != nil {
		t.Fatalf("PrintDeleteSuccess() error = %v", err)
	}
	if got := buf.String(); got != "Deleted score \"score-1\".\n" {
		t.Errorf("delete output mismatch got %q", got)
	}
}
