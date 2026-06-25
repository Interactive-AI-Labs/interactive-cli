package summary

import (
	"encoding/json"
	"strings"
)

// UnwrapJSON removes one layer of JSON-string wrapping if the raw value is a
// JSON string whose contents are themselves valid JSON. Langfuse stores some
// IO this way (e.g. "{\"a\":1}"). A native JSON value passes through unchanged.
func UnwrapJSON(raw json.RawMessage) json.RawMessage {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		var inner json.RawMessage
		if json.Unmarshal([]byte(s), &inner) == nil {
			return inner
		}
		b, _ := json.Marshal(s) // plain string: re-encode as valid JSON
		return b
	}
	return raw
}

// AsString renders a raw IO value as human text:
//   - JSON string            -> the string (recursing if it wraps an array)
//   - JSON array of strings  -> joined by newlines
//   - anything else          -> compact JSON
func AsString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if joined, ok := tryJoinStringArray([]byte(s)); ok {
			return joined
		}
		return s
	}
	if joined, ok := tryJoinStringArray(raw); ok {
		return joined
	}
	return CompactJSON(raw)
}

func tryJoinStringArray(b []byte) (string, bool) {
	var arr []string
	if err := json.Unmarshal(b, &arr); err == nil {
		return strings.Join(arr, "\n"), true
	}
	return "", false
}

// CompactJSON re-encodes raw JSON without indentation; on failure returns the
// raw bytes as a string.
func CompactJSON(raw json.RawMessage) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}
	return string(b)
}

// CollapseWS collapses all runs of whitespace (incl. newlines) to single spaces.
func CollapseWS(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
