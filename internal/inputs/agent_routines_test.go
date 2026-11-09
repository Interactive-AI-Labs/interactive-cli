package inputs

import (
	"strings"
	"testing"
)

func TestValidateRoutineRun(t *testing.T) {
	const invalidID = "routine ID must contain nonempty path segments without '.' or '..' segments or backslashes"
	tests := []struct {
		name, agent, routine, callback, wantErr string
	}{
		{"folder routine", "my-agent", "billing/refund", "https://trusted.example/results", ""},
		{"local callback", "my-agent", "refund", "http://localhost:8080/result", ""},
		{"empty agent", " ", "refund", "https://trusted.example", "agent name must not be empty"},
		{"empty routine", "my-agent", "", "https://trusted.example", invalidID},
		{"empty folder", "my-agent", "billing//refund", "https://trusted.example", invalidID},
		{"traversal", "my-agent", "billing/../refund", "https://trusted.example", invalidID},
		{"backslash", "my-agent", `billing\refund`, "https://trusted.example", invalidID},
		{
			"missing callback",
			"my-agent",
			"refund",
			"",
			"--callback-url must be an HTTP(S) URL with a host",
		},
		{
			"relative callback",
			"my-agent",
			"refund",
			"/results",
			"--callback-url must be an HTTP(S) URL with a host",
		},
		{
			"wrong scheme",
			"my-agent",
			"refund",
			"file:///tmp/result",
			"--callback-url must be an HTTP(S) URL with a host",
		},
		{
			"missing host",
			"my-agent",
			"refund",
			"https:///result",
			"--callback-url must be an HTTP(S) URL with a host",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRoutineRun(tt.agent, tt.routine, tt.callback)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("error = %q, want %q", gotErr, tt.wantErr)
			}
		})
	}
}

func TestReadRoutineInput(t *testing.T) {
	tests := []struct {
		name, path, input, want, wantErr string
	}{
		{"object", "-", `{"order_id":"order_123"}`, `{"order_id":"order_123"}`, ""},
		{
			"exact numbers and nested keys",
			"-",
			`{"n":9007199254740993,"sessionId":{"n":1.234567890123456789}}`,
			`{"n":9007199254740993,"sessionId":{"n":1.234567890123456789}}`,
			"",
		},
		{"empty object", "-", `{}`, `{}`, ""},
		{"missing path", "", `{}`, "", "--input-file must not be empty"},
		{"empty input", "-", "", "", "routine input must be a JSON object"},
		{"null", "-", `null`, "", "routine input must be a JSON object"},
		{"array", "-", `[]`, "", "routine input must be a JSON object"},
		{"scalar", "-", `1`, "", "routine input must be a JSON object"},
		{"YAML", "-", "order_id: order_123", "", "routine input must be a JSON object"},
		{"trailing value", "-", `{} {}`, "", "routine input must be a JSON object"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadRoutineInput(tt.path, strings.NewReader(tt.input))
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Fatalf("error = %q, want %q", gotErr, tt.wantErr)
			}
			if string(got) != tt.want {
				t.Errorf("input = %s, want %s", got, tt.want)
			}
		})
	}
}
