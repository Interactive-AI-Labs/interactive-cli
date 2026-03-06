package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var traceColumnMap = map[string]struct {
	Header string
	Value  func(t *clients.TraceInfo) string
}{
	"id":          {"ID", func(t *clients.TraceInfo) string { return t.ID }},
	"name":        {"NAME", func(t *clients.TraceInfo) string { return t.Name }},
	"timestamp":   {"TIMESTAMP", func(t *clients.TraceInfo) string { return LocalTime(t.Timestamp) }},
	"user_id":     {"USER ID", func(t *clients.TraceInfo) string { return t.UserID }},
	"session_id":  {"SESSION ID", func(t *clients.TraceInfo) string { return t.SessionID }},
	"release":     {"RELEASE", func(t *clients.TraceInfo) string { return t.Release }},
	"version":     {"VERSION", func(t *clients.TraceInfo) string { return t.Version }},
	"environment": {"ENVIRONMENT", func(t *clients.TraceInfo) string { return t.Environment }},
	"public":      {"PUBLIC", func(t *clients.TraceInfo) string { return fmt.Sprintf("%t", t.Public) }},
	"latency":     {"LATENCY", func(t *clients.TraceInfo) string { return formatFloat(t.Latency, "s") }},
	"cost":        {"COST", func(t *clients.TraceInfo) string { return formatCost(t.TotalCost) }},
	"tags":        {"TAGS", func(t *clients.TraceInfo) string { return TruncateList(t.Tags, 3) }},
}

func PrintTraceList(out io.Writer, traces []clients.TraceInfo, meta clients.TraceMeta, columns []string) error {
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

	fmt.Fprintf(out, "\nPage %d of %d (%d total items)\n", meta.Page, meta.TotalPages, meta.TotalItems)
	return nil
}

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

	if len(trace.Input) > 0 && string(trace.Input) != "null" {
		fmt.Fprintf(out, "\nInput:\n%s\n", prettyJSON(trace.Input))
	}

	if len(trace.Output) > 0 && string(trace.Output) != "null" {
		fmt.Fprintf(out, "\nOutput:\n%s\n", prettyJSON(trace.Output))
	}

	if len(trace.Metadata) > 0 && string(trace.Metadata) != "null" {
		fmt.Fprintf(out, "\nMetadata:\n%s\n", prettyJSON(trace.Metadata))
	}

	return nil
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

func prettyJSON(raw json.RawMessage) string {
	var buf []byte
	buf, err := json.MarshalIndent(json.RawMessage(raw), "  ", "  ")
	if err != nil {
		return string(raw)
	}
	return "  " + string(buf)
}
