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
			got := AsString(json.RawMessage(tc.raw))
			if got != tc.want {
				t.Fatalf("AsString(%s) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestUnwrapJSON(t *testing.T) {
	// A JSON string whose content is a JSON object must unwrap to the object.
	got := UnwrapJSON(json.RawMessage(`"{\"a\":1}"`))
	var m map[string]int
	if err := json.Unmarshal(got, &m); err != nil || m["a"] != 1 {
		t.Fatalf("UnwrapJSON did not unwrap string-wrapped object: %s (err %v)", got, err)
	}
	// A native object passes through.
	got = UnwrapJSON(json.RawMessage(`{"b":2}`))
	if err := json.Unmarshal(got, &m); err != nil || m["b"] != 2 {
		t.Fatalf("UnwrapJSON mangled native object: %s", got)
	}
}

func TestTruncateRuneSafe(t *testing.T) {
	if got := Truncate("  abcdef  ", 3); got != "abc… (truncated)" {
		t.Fatalf("Truncate = %q", got)
	}
	if got := Truncate("héllo", 10); got != "héllo" {
		t.Fatalf("Truncate short = %q", got)
	}
}

func TestUnwrapToolResult(t *testing.T) {
	// The engine envelope collapses to just its data payload.
	env := `{"data":{"ok":true,"value":3},"metadata":{},"control":{},"canned_responses":[],"canned_response_fields":{},"guidelines":[]}`
	got := CompactJSON(UnwrapToolResult(json.RawMessage(env)))
	if got != `{"ok":true,"value":3}` {
		t.Fatalf("envelope unwrap = %q", got)
	}
	// String-wrapped envelope also unwraps.
	got = CompactJSON(UnwrapToolResult(json.RawMessage(`"{\"data\":{\"x\":1},\"metadata\":{},\"control\":{}}"`)))
	if got != `{"x":1}` {
		t.Fatalf("string-wrapped envelope unwrap = %q", got)
	}
	// A plain object with an unexpected sibling is left untouched.
	plain := `{"data":{"x":1},"other":true}`
	if got := CompactJSON(UnwrapToolResult(json.RawMessage(plain))); got != `{"data":{"x":1},"other":true}` {
		t.Fatalf("non-envelope should pass through, got %q", got)
	}
	// A value with no data key passes through.
	if got := CompactJSON(UnwrapToolResult(json.RawMessage(`{"count":3}`))); got != `{"count":3}` {
		t.Fatalf("no-data passthrough = %q", got)
	}
	// A non-object passes through.
	if got := CompactJSON(UnwrapToolResult(json.RawMessage(`[1,2]`))); got != `[1,2]` {
		t.Fatalf("array passthrough = %q", got)
	}
}

func TestCompactArgs(t *testing.T) {
	got := CompactArgs(json.RawMessage(`"{\"dates\":\"next weekend\",\"qty\":2}"`))
	// keys sorted, k=v form
	if got != `dates="next weekend", qty=2` {
		t.Fatalf("CompactArgs = %q", got)
	}
}
