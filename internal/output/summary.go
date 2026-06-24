package output

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/summary"
)

// compactArgs renders tool arguments as `k=v, k=v` (keys sorted) for a flat
// object, falling back to compact JSON for anything else.
func compactArgs(raw json.RawMessage) string {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return summary.CompactJSON(raw)
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, formatValue(m[k])))
	}
	return strings.Join(parts, ", ")
}

// formatValue renders decoded JSON values compactly.
func formatValue(v any) string {
	switch t := v.(type) {
	case string:
		return fmt.Sprintf("%q", t)
	case float64:
		// JSON numbers decode to float64; print integers without trailing zeros.
		// Guard the int64 conversion: out-of-range floats overflow undefined.
		if t == math.Trunc(t) && t >= math.MinInt64 && t <= math.MaxInt64 {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case nil:
		return "null"
	default:
		b, _ := json.Marshal(t)
		return string(b)
	}
}
