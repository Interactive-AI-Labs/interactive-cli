package clients

import (
	"encoding/json"
	"testing"
)

func TestWithHybridMode(t *testing.T) {
	// Injects mode when absent and preserves the rest of the body.
	out, err := withHybridMode([]byte(`{"queries":[{"query":"x"}],"limit":5}`))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if m["mode"] != "hybrid" {
		t.Errorf("mode not set to hybrid: %v", m["mode"])
	}
	if _, ok := m["queries"]; !ok {
		t.Error("queries lane dropped")
	}

	// Overrides an existing mode too (the subcommand always means hybrid).
	out, _ = withHybridMode([]byte(`{"mode":"single","queries":[]}`))
	_ = json.Unmarshal(out, &m)
	if m["mode"] != "hybrid" {
		t.Errorf("mode not forced to hybrid, got %v", m["mode"])
	}

	// A non-object body is a clear error, not a silent pass-through.
	if _, err := withHybridMode([]byte(`[1,2,3]`)); err == nil {
		t.Error("expected error for non-object body")
	}
}
