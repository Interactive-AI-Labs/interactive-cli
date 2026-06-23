package summary

import (
	"encoding/json"
	"testing"
)

func TestAsString(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"plain json string", `"hello world"`, "hello world"},
		{"string-wrapped array", `"[\"a\",\"b\"]"`, "a\nb"},
		{"native array", `["x","y"]`, "x\ny"},
		{"object falls back to compact json", `{"k":1}`, `{"k":1}`},
		{"empty", ``, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := AsString(json.RawMessage(tc.raw)); got != tc.want {
				t.Fatalf("AsString(%s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestUnwrapJSON(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"string-wrapped object unwraps", `"{\"a\":1}"`, `{"a":1}`},
		{"native object passes through", `{"b":2}`, `{"b":2}`},
		{"plain string re-encodes as json string", `"hello"`, `"hello"`},
		{"non-string value passes through", `5`, `5`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(UnwrapJSON(json.RawMessage(tc.raw))); got != tc.want {
				t.Fatalf("UnwrapJSON(%s) = %s, want %s", tc.raw, got, tc.want)
			}
		})
	}
}

func TestCollapseWS(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"newlines and runs collapse", "a  b\n\tc", "a b c"},
		{"trimmed", "  x  ", "x"},
		{"empty", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CollapseWS(tc.in); got != tc.want {
				t.Fatalf("CollapseWS(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
