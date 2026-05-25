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
			name:  "raw flag prints exact server JSON",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"hello\"}"}` + "\n",
			opts:  LogFormatOptions{Raw: true},
			want:  `{"line":"{\"level\":\"info\",\"msg\":\"hello\"}"}` + "\n",
		},
		{
			name:  "raw with decode decodes structured line field",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"hello\"}"}` + "\n",
			opts:  LogFormatOptions{Raw: true, Decode: true},
			want:  `{"line":{"level":"info","msg":"hello"}}` + "\n",
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
			name:  "raw with decode preserves plain-text line field",
			input: `{"line":"plain text","replica":"pod-1"}` + "\n",
			opts:  LogFormatOptions{Raw: true, Decode: true},
			want:  `{"line":"plain text","replica":"pod-1"}` + "\n",
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
		{
			name:  "record field is not unwrapped without CNPGFormat",
			input: `{"line":"{\"msg\":\"record\",\"record\":{\"message\":\"checkpoint complete\",\"error_severity\":\"LOG\"}}"}` + "\n",
			want:  "record\n",
		},
		{
			name:  "CNPGFormat unwraps record envelope",
			input: `{"line":"{\"msg\":\"record\",\"record\":{\"message\":\"checkpoint complete\",\"error_severity\":\"LOG\"}}"}` + "\n",
			opts:  LogFormatOptions{CNPGFormat: true},
			want:  "LOG   checkpoint complete\n",
		},
		{
			name:  "CNPGFormat with all-fields shows extra fields without flattening record",
			input: `{"line":"{\"msg\":\"record\",\"logger\":\"postgres\",\"record\":{\"message\":\"checkpoint complete\",\"error_severity\":\"LOG\",\"pid\":123,\"backend_type\":\"checkpointer\"}}"}` + "\n",
			opts:  LogFormatOptions{CNPGFormat: true, AllFields: true},
			want:  "LOG   checkpoint complete  logger=\"postgres\"\n",
		},
		{
			name:  "CNPGFormat falls back to standard fields when no record found",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"Setting 0750 permissions\"}"}` + "\n",
			opts:  LogFormatOptions{CNPGFormat: true},
			want:  "INFO  Setting 0750 permissions\n",
		},
		{
			name:  "CNPGFormat falls back to raw when no record and no standard fields",
			input: `{"line":"{\"some\":\"data\"}"}` + "\n",
			opts:  LogFormatOptions{CNPGFormat: true},
			want:  `{"some":"data"}` + "\n",
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

func TestPrintLogStreamFieldsMissing(t *testing.T) {
	input := `{"line":"{\"level\":\"info\",\"msg\":\"hello\",\"pid\":42}"}` + "\n"

	var buf bytes.Buffer
	r := strings.NewReader(input)
	err := PrintLogStream(&buf, r, false, LogsMeta{}, LogFormatOptions{
		Fields: []string{"nonexistent", "also_missing"},
	})
	if err != nil {
		t.Fatalf("PrintLogStream() error = %v", err)
	}
	want := "INFO  hello\n"
	if got := buf.String(); got != want {
		t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestPrintNoLogsFound(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "prints plain no-logs message for non-terminal writer",
			want: "No logs found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintNoLogsFound(&buf)
			if got := buf.String(); got != tt.want {
				t.Errorf("message mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintLogFieldDiscoveryTruncationWarning(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "prints plain warning text for non-terminal writer",
			want: "Warning: field discovery may be incomplete because the server truncated the log response (max 5000 lines). Use --since to scan a narrower time range.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			PrintLogFieldDiscoveryTruncationWarning(&buf)
			if got := buf.String(); got != tt.want {
				t.Errorf("warning mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestDiscoverLogFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []LogField
	}{
		{
			name:  "empty input returns no fields",
			input: "",
			want:  nil,
		},
		{
			name:  "non-JSON lines are skipped",
			input: "plain text\nnot json\n",
			want:  nil,
		},
		{
			name:  "standard fields are excluded",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"hello\",\"ts\":\"2026-01-01T00:00:00Z\"}"}` + "\n",
			want:  nil,
		},
		{
			name: "extra fields are discovered with counts",
			input: `{"line":"{\"level\":\"info\",\"msg\":\"a\",\"logger\":\"pg\",\"pid\":1}"}` + "\n" +
				`{"line":"{\"level\":\"info\",\"msg\":\"b\",\"logger\":\"pg\"}"}` + "\n",
			want: []LogField{
				{Name: "logger", Count: 2},
				{Name: "pid", Count: 1},
			},
		},
		{
			name:  "plain-text line field is skipped",
			input: `{"line":"not json at all"}` + "\n",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := DiscoverLogFields(r)
			if err != nil {
				t.Fatalf("DiscoverLogFields() error = %v", err)
			}
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("field count mismatch: got %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("field[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
