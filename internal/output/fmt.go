package output

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

const missingMetadata = "—" // table cell the API supplied no value for

// formatCost formats a pointer cost as dollars for terminal output.
func formatCost(v *float64) string {
	if v == nil {
		return "-"
	}
	// Fifteen significant digits absorbs float noise such as 2.9999999999999996 at any magnitude.
	rounded := roundSignificant(*v, 15)
	if rounded == 0 {
		rounded = 0 // drops the sign of -0
	}
	whole, frac, _ := strings.Cut(strconv.FormatFloat(rounded, 'f', -1, 64), ".")
	for len(frac) < 2 {
		frac += "0"
	}
	return "$" + whole + "." + frac
}

func roundSignificant(v float64, digits int) float64 {
	r, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'g', digits, 64), 64)
	return r
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

// Integer arithmetic so 1,050,000 and 1,150,000 both round half up.
func formatTokenCount(n *int) string {
	switch {
	case n == nil:
		return "-"
	case *n < 1000:
		return strconv.Itoa(*n)
	case *n < 999_500:
		return strconv.Itoa((*n+500)/1000) + "k"
	}
	tenths := (*n + 50_000) / 100_000
	return fmt.Sprintf("%d.%dM", tenths/10, tenths%10)
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

// Printable neutralises terminal control characters in server-supplied text.
func Printable(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || unicode.IsPrint(r) {
			return r
		}
		return '\ufffd'
	}, s)
}

// HumanBytes renders a byte count in the largest unit that keeps it >= 1.
func HumanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	val := float64(n) / float64(div)
	if val >= 1023.95 { // %.1f would render 1024.0; step up to the next unit
		val /= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", val, "KMGTPE"[exp])
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
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return v
		}
		n = parsed
	default:
		return fmt.Sprint(value)
	}
	return formatCost(&n)
}
