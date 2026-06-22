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

func TestCompactArgs(t *testing.T) {
	got := CompactArgs(json.RawMessage(`"{\"dates\":\"next weekend\",\"qty\":2}"`))
	// keys sorted, k=v form
	if got != `dates="next weekend", qty=2` {
		t.Fatalf("CompactArgs = %q", got)
	}
}
