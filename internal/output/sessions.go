package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var sessionColumnMap = map[string]struct {
	Header string
	Value  func(s *clients.SessionInfo) string
}{
	"id": {"ID", func(s *clients.SessionInfo) string { return s.ID }},
	"created_at": {
		"CREATED AT",
		func(s *clients.SessionInfo) string { return LocalTime(s.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(s *clients.SessionInfo) string { return LocalTime(s.UpdatedAt) },
	},
	"environment": {
		"ENVIRONMENT",
		func(s *clients.SessionInfo) string { return s.Environment },
	},
	"user_id": {"USER ID", func(s *clients.SessionInfo) string { return s.UserID }},
	"trace_count": {
		"TRACE COUNT",
		func(s *clients.SessionInfo) string { return formatInt(s.TraceCount) },
	},
	"duration_seconds": {
		"DURATION (s)",
		func(s *clients.SessionInfo) string { return formatFloat(s.DurationSeconds, "s") },
	},
	"total_cost": {
		"TOTAL COST",
		func(s *clients.SessionInfo) string { return formatCost(s.TotalCost) },
	},
	"input_tokens": {
		"INPUT TOKENS",
		func(s *clients.SessionInfo) string { return formatInt(s.InputTokens) },
	},
	"output_tokens": {
		"OUTPUT TOKENS",
		func(s *clients.SessionInfo) string { return formatInt(s.OutputTokens) },
	},
	"total_tokens": {
		"TOTAL TOKENS",
		func(s *clients.SessionInfo) string { return formatInt(s.TotalTokens) },
	},
}

func PrintSessionList(
	out io.Writer,
	sessions []clients.SessionInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(sessions) == 0 {
		fmt.Fprintln(out, "No sessions found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := sessionColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(sessions))
	for i, session := range sessions {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := sessionColumnMap[col]; ok {
				row[j] = def.Value(&session)
			}
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	PrintPageMeta(out, meta.Page, meta.TotalPages, meta.TotalItems)
	return nil
}

func PrintSessionDetail(out io.Writer, session *clients.SessionDetail) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", session.ID)
	fmt.Fprintf(w, "Created At:\t%s\n", LocalTime(session.CreatedAt))
	fmt.Fprintf(w, "Updated At:\t%s\n", LocalTime(session.UpdatedAt))
	fmt.Fprintf(w, "Environment:\t%s\n", session.Environment)
	fmt.Fprintf(w, "User ID:\t%s\n", session.UserID)

	fmt.Fprintln(w)
	fmt.Fprintln(w, "--- Metrics ---")
	fmt.Fprintf(w, "Trace Count:\t%s\n", formatInt(session.TraceCount))
	fmt.Fprintf(w, "Duration Seconds:\t%s\n", formatFloat(session.DurationSeconds, "s"))
	fmt.Fprintf(w, "Total Cost:\t%s\n", formatCost(session.TotalCost))
	fmt.Fprintf(w, "Input Tokens:\t%s\n", formatInt(session.InputTokens))
	fmt.Fprintf(w, "Output Tokens:\t%s\n", formatInt(session.OutputTokens))
	fmt.Fprintf(w, "Total Tokens:\t%s\n", formatInt(session.TotalTokens))
	if err := w.Flush(); err != nil {
		return err
	}

	if len(session.Traces) == 0 {
		return nil
	}

	fmt.Fprintf(out, "\n--- Trace Summaries ---\n")
	headers := []string{
		"ID",
		"NAME",
		"TIMESTAMP",
		"LATENCY",
		"COST",
		"OBSERVATIONS",
		"TOTAL TOKENS",
		"LEVEL",
	}
	rows := make([][]string, len(session.Traces))
	for i, trace := range session.Traces {
		rows[i] = []string{
			trace.ID,
			trace.Name,
			LocalTime(trace.Timestamp),
			formatLatencyMs(trace.LatencyMs),
			formatCost(trace.TotalCost),
			formatInt(trace.ObservationCount),
			formatInt(trace.TotalTokens),
			trace.Level,
		}
	}

	return PrintTable(out, headers, rows)
}
