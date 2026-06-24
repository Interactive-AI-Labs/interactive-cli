package output

import (
	"encoding/json"
	"testing"
)

func TestCompactArgs(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{
			"flat object sorts keys as k=v",
			`{"qty":2,"dates":"next weekend"}`,
			`dates="next weekend", qty=2`,
		},
		{"integer float prints without trailing zeros", `{"n":3}`, `n=3`},
		{"non-object falls back to compact json", `[1,2]`, `[1,2]`},
		{"empty input", ``, ``},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := compactArgs(json.RawMessage(tc.raw)); got != tc.want {
				t.Fatalf("compactArgs(%s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"string quoted", "hi", `"hi"`},
		{"integer float", float64(3), "3"},
		{"fractional float", 3.5, "3.5"},
		{"nil", nil, "null"},
		{"slice as compact json", []any{1.0, 2.0}, "[1,2]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatValue(tc.in); got != tc.want {
				t.Fatalf("formatValue(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
