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
			err := PrintLogStream(&buf, r, tt.showReplica, LogsMeta{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("PrintLogStream() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
