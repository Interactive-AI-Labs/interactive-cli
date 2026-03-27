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

	fmt.Fprintf(
		out,
		"\nPage %d of %d (%d total items)\n",
		meta.Page,
		meta.TotalPages,
		meta.TotalItems,
	)
	return nil
}

func PrintSessionDetail(out io.Writer, session *clients.SessionDetail) error {
	fmt.Fprintf(out, "ID:          %s\n", session.ID)
	fmt.Fprintf(out, "Created At:  %s\n", LocalTime(session.CreatedAt))
	fmt.Fprintf(out, "Updated At:  %s\n", LocalTime(session.UpdatedAt))
	fmt.Fprintf(out, "Environment: %s\n", session.Environment)
	fmt.Fprintf(out, "User ID:     %s\n", session.UserID)

	fmt.Fprintf(out, "\n--- Metrics ---\n")
	fmt.Fprintf(out, "Trace Count:       %s\n", formatInt(session.TraceCount))
	fmt.Fprintf(out, "Duration Seconds:  %s\n", formatFloat(session.DurationSeconds, "s"))
	fmt.Fprintf(out, "Total Cost:        %s\n", formatCost(session.TotalCost))
	fmt.Fprintf(out, "Input Tokens:      %s\n", formatInt(session.InputTokens))
	fmt.Fprintf(out, "Output Tokens:     %s\n", formatInt(session.OutputTokens))
	fmt.Fprintf(out, "Total Tokens:      %s\n", formatInt(session.TotalTokens))

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
