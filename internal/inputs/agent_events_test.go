package inputs

import (
	"strings"
	"testing"
)

func TestBuildEventBody(t *testing.T) {
	tests := []struct {
		name    string
		message string
		file    string
		stdin   string
		want    string
		wantErr string
	}{
		{
			name:    "message",
			message: "Hi",
			want:    `{"kind":"message","message":"Hi","source":"customer"}`,
		},
		{
			name:    "message is JSON-escaped",
			message: `say "hi" <now>`,
			want:    `{"kind":"message","message":"say \"hi\" \u003cnow\u003e","source":"customer"}`,
		},
		{
			name:  "stdin JSON",
			file:  "-",
			stdin: `{"kind":"custom","source":"customer","data":{"x":1}}`,
			want:  `{"kind":"custom","source":"customer","data":{"x":1}}`,
		},
		{
			name:    "stdin not JSON",
			file:    "-",
			stdin:   `kind: custom`,
			wantErr: "event file must be a JSON object",
		},
		{
			name:    "stdin JSON that is not an object",
			file:    "-",
			stdin:   `[1]`,
			wantErr: "event file must be a JSON object",
		},
		{
			name:    "stdin null",
			file:    "-",
			stdin:   `null`,
			wantErr: "event file must be a JSON object",
		},
		{
			name:    "empty message",
			wantErr: "--message must not be empty",
		},
		{
			name:    "missing file",
			file:    "/nonexistent/event.json",
			wantErr: "failed to read event file: open /nonexistent/event.json: no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildEventBody(tt.message, tt.file, strings.NewReader(tt.stdin))
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Fatalf("error = %q, want %q", gotErr, tt.wantErr)
			}
			if string(got) != tt.want {
				t.Errorf("body = %s, want %s", got, tt.want)
			}
		})
	}
}
