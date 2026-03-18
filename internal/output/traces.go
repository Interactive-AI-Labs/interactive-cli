package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var traceColumnMap = map[string]struct {
	Header string
	Value  func(t *clients.TraceInfo) string
}{
	"id":   {"ID", func(t *clients.TraceInfo) string { return t.ID }},
	"name": {"NAME", func(t *clients.TraceInfo) string { return t.Name }},
	"timestamp": {
		"TIMESTAMP",
		func(t *clients.TraceInfo) string { return LocalTime(t.Timestamp) },
	},
	"user_id":     {"USER ID", func(t *clients.TraceInfo) string { return t.UserID }},
	"session_id":  {"SESSION ID", func(t *clients.TraceInfo) string { return t.SessionID }},
	"release":     {"RELEASE", func(t *clients.TraceInfo) string { return t.Release }},
	"version":     {"VERSION", func(t *clients.TraceInfo) string { return t.Version }},
	"environment": {"ENVIRONMENT", func(t *clients.TraceInfo) string { return t.Environment }},
	"public": {
		"PUBLIC",
		func(t *clients.TraceInfo) string { return fmt.Sprintf("%t", t.Public) },
	},
	"latency": {
		"LATENCY",
		func(t *clients.TraceInfo) string { return formatFloat(t.Latency, "s") },
	},
	"cost": {"COST", func(t *clients.TraceInfo) string { return formatCost(t.TotalCost) }},
	"tags": {"TAGS", func(t *clients.TraceInfo) string { return TruncateList(t.Tags, 3) }},
}

func PrintTraceList(
	out io.Writer,
	traces []clients.TraceInfo,
	meta clients.TraceMeta,
	columns []string,
) error {
	if len(traces) == 0 {
		fmt.Fprintln(out, "No traces found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = traceColumnMap[col].Header
	}

	rows := make([][]string, len(traces))
	for i, t := range traces {
		row := make([]string, len(columns))
		for j, col := range columns {
			row[j] = traceColumnMap[col].Value(&t)
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	fmt.Fprintf(
		out,
		"\nPage %d of %d (%d total items)\n",
		meta.Page,
		meta.TotalPages,
		meta.TotalItems,
	)
	return nil
}

const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[91m"
	colorOrange = "\033[33m"
)

func PrintTraceDetail(out io.Writer, trace *clients.TraceDetail) error {
	fmt.Fprintf(out, "ID:          %s\n", trace.ID)
	fmt.Fprintf(out, "Name:        %s\n", trace.Name)
	fmt.Fprintf(out, "Timestamp:   %s\n", LocalTime(trace.Timestamp))
	fmt.Fprintf(out, "Session ID:  %s\n", trace.SessionID)
	fmt.Fprintf(out, "User ID:     %s\n", trace.UserID)
	fmt.Fprintf(out, "Environment: %s\n", trace.Environment)
	fmt.Fprintf(out, "Release:     %s\n", trace.Release)
	fmt.Fprintf(out, "Version:     %s\n", trace.Version)
	fmt.Fprintf(out, "Public:      %t\n", trace.Public)
	fmt.Fprintf(out, "Latency:     %s\n", formatFloat(trace.Latency, "s"))
	fmt.Fprintf(out, "Total Cost:  %s\n", formatCost(trace.TotalCost))

	if len(trace.Tags) > 0 {
		fmt.Fprintf(out, "Tags:        %s\n", strings.Join(trace.Tags, ", "))
	}

	if trace.HtmlPath != "" {
		fmt.Fprintf(out, "URL Path:    %s\n", trace.HtmlPath)
	}

	isTTY := isTerminal(out)

	const jsonPrefix = "  "
	if len(trace.Input) > 0 && string(trace.Input) != "null" {
		fmt.Fprintf(
			out,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Input:", colorGreen),
			indentLines(prettyJSON(trace.Input, isTTY), jsonPrefix),
		)
	}

	if len(trace.Output) > 0 && string(trace.Output) != "null" {
		fmt.Fprintf(
			out,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Output:", colorRed),
			indentLines(prettyJSON(trace.Output, isTTY), jsonPrefix),
		)
	}

	if len(trace.Metadata) > 0 && string(trace.Metadata) != "null" {
		fmt.Fprintf(
			out,
			"\n%s\n%s\n",
			colorHeader(isTTY, "Metadata:", colorOrange),
			indentLines(prettyJSON(trace.Metadata, isTTY), jsonPrefix),
		)
	}

	return nil
}

func indentLines(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func colorHeader(useColor bool, label string, color string) string {
	if !useColor {
		return label
	}
	return color + label + colorReset
}

func formatFloat(v *float64, suffix string) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.2f%s", *v, suffix)
}

func formatCost(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("$%.6f", *v)
}

// jsonUnescaper performs a single-pass unescape of JSON escape sequences for
// human-readable terminal output. NewReplacer matches left-to-right without
// re-processing replaced text, so \\n correctly becomes literal \n (not a newline).
var jsonUnescaper = strings.NewReplacer(`\\`, `\`, `\n`, "\n", `\t`, "\t", `\"`, `"`)

func prettyJSON(raw json.RawMessage, unescape bool) string {
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return string(raw)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// show &, <, > as-is instead of \u0026, \u003c, \u003e
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(decoded); err != nil {
		return string(raw)
	}
	s := strings.TrimRight(buf.String(), "\n")
	if unescape {
		s = jsonUnescaper.Replace(s)
	}
	return s
}
