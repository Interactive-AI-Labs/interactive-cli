package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintLogStream(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		showReplica bool
		opts        LogFormatOptions
		want        string
		wantErr     bool
	}{
		{
			name:  "empty input produces no output",
			input: "",
			want:  "",
		},
		{
			name:  "plain non-JSON lines are printed as-is",
			input: "hello world\nfoo bar\n",
			want:  "hello world\nfoo bar\n",
		},
		{
			name:  "empty lines are skipped",
			input: "first\n\nsecond\n",
			want:  "first\nsecond\n",
		},
		{
			name:  "JSON entry without replica prints line only",
			input: `{"line":"log message"}` + "\n",
			want:  "log message\n",
		},
		{
			name:  "JSON entry with replica but showReplica false prints line only",
			input: `{"replica":"pod-1","line":"hello"}` + "\n",
			want:  "hello\n",
		},
		{
			name:        "JSON entry with replica and showReplica true prints prefix without color",
			input:       `{"replica":"pod-1","line":"hello"}` + "\n",
			showReplica: true,
			want:        "[pod-1] hello\n",
		},
		{
			name:        "JSON entry with empty replica and showReplica true prints line only",
			input:       `{"replica":"","line":"no replica"}` + "\n",
			showReplica: true,
			want:        "no replica\n",
		},
		{
			name: "multiple replicas get prefixed",
			input: `{"replica":"pod-1","line":"first"}` + "\n" +
				`{"replica":"pod-2","line":"second"}` + "\n" +
				`{"replica":"pod-1","line":"third"}` + "\n",
			showReplica: true,
			want: "[pod-1] first\n" +
				"[pod-2] second\n" +
				"[pod-1] third\n",
		},
		{
			name: "mixed JSON and non-JSON lines",
			input: "raw line\n" +
				`{"line":"json line"}` + "\n" +
				"another raw\n",
			want: "raw line\n" +
				"json line\n" +
				"another raw\n",
		},
		{
			name:  "malformed JSON is printed as-is",
			input: `{"line":broken}` + "\n",
			want:  `{"line":broken}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := strings.NewReader(tt.input)
			err := PrintLogStream(&buf, r, tt.showReplica, LogsMeta{}, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PrintLogStream() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintLogStreamStructuredJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		showReplica bool
		opts        LogFormatOptions
		want        string
	}{
		{
			name:  "structured JSON log extracts level and msg",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"Starting up\",\"ts\":\"2026-01-01T00:00:00Z\"}"}` + "\n",
			want:  "INFO  Starting up\n",
		},
		{
			name:  "structured JSON with message field instead of msg",
			input: `{"line":"{\"level\":\"error\",\"message\":\"Something failed\"}"}` + "\n",
			want:  "ERROR Something failed\n",
		},
		{
			name:  "json flag outputs raw server envelope",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"hello\"}"}` + "\n",
			opts:  LogFormatOptions{JSON: true},
			want:  `{"line":"{\"level\":\"info\",\"msg\":\"hello\"}"}` + "\n",
		},
		{
			name:  "fields flag shows selected fields on indented line",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"hello\",\"logger\":\"postgres\",\"pid\":42}"}` + "\n",
			opts:  LogFormatOptions{Fields: []string{"logger", "pid"}},
			want:  "INFO  hello  logger=\"postgres\" pid=42\n",
		},
		{
			name:  "non-JSON line is passed through even in default mode",
			input: `{"line":"plain text log"}` + "\n",
			want:  "plain text log\n",
		},
		{
			name:        "structured JSON with replica prefix",
			input:       `{"replica":"db-1","line":"{\"level\":\"warning\",\"msg\":\"low memory\"}"}` + "\n",
			showReplica: true,
			want:        "[db-1] WARNING low memory\n",
		},
		{
			name:  "all-fields flag shows all extra fields sorted",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"hello\",\"pid\":42,\"logger\":\"postgres\"}"}` + "\n",
			opts:  LogFormatOptions{AllFields: true},
			want:  "INFO  hello  logger=\"postgres\" pid=42\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			r := strings.NewReader(tt.input)
			err := PrintLogStream(&buf, r, tt.showReplica, LogsMeta{}, tt.opts)
			if err != nil {
				t.Fatalf("PrintLogStream() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
