package output

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// formatCost formats a pointer cost as dollars for terminal output.
func formatCost(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("$%.6f", *v)
}

// formatFloat formats a pointer float with a suffix for terminal output.
func formatFloat(v *float64, suffix string) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f%s", *v, suffix)
}

// formatInt formats a pointer int for terminal output.
func formatInt(v *int) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *v)
}

// formatLatencyMs formats milliseconds as ms or seconds for terminal output.
func formatLatencyMs(v *float64) string {
	if v == nil {
		return "-"
	}
	if *v <= 1000 {
		return fmt.Sprintf("%.0fms", *v)
	}
	return fmt.Sprintf("%.2fs", *v/1000)
}

// formatOptionalFloat formats an optional float without forced trailing zeros.
func formatOptionalFloat(v *float64) string {
	if v == nil {
		return "-"
	}
	return strconv.FormatFloat(*v, 'f', -1, 64)
}

// formatSecretKeys formats a truncated comma-separated key list.
func formatSecretKeys(keys []string, maxVisible int) string {
	if len(keys) == 0 {
		return ""
	}
	if len(keys) <= maxVisible {
		return strings.Join(keys, ", ")
	}
	visible := strings.Join(keys[:maxVisible], ", ")
	return fmt.Sprintf("%s (+%d more)", visible, len(keys)-maxVisible)
}

// formatSummaryValue renders decoded JSON values compactly.
func formatSummaryValue(v any) string {
	switch t := v.(type) {
	case string:
		return fmt.Sprintf("%q", t)
	case float64:
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

// formatUSD formats numeric values as dollars for table display.
func formatUSD(value any) string {
	if value == nil {
		return ""
	}
	var n float64
	switch v := value.(type) {
	case float64:
		n = v
	case float32:
		n = float64(v)
	case int:
		n = float64(v)
	case int64:
		n = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return v
		}
		n = parsed
	default:
		return fmt.Sprint(value)
	}
	return fmt.Sprintf("$%.2f", n)
}
