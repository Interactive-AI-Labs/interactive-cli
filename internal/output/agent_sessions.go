package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

const eventSummaryWidth = 80

func PrintAgentSessionList(out io.Writer, list *deployment.AgentSessionList) error {
	if len(list.Items) == 0 {
		fmt.Fprintln(out, "No sessions found.")
		return nil
	}
	headers := []string{"ID", "CUSTOMER", "TITLE", "LABELS", "CREATED"}
	rows := make([][]string, len(list.Items))
	for i, s := range list.Items {
		rows[i] = []string{
			s.ID,
			s.CustomerID,
			derefString(s.Title),
			TruncateList(s.Labels, 3),
			LocalTime(s.CreatedAt),
		}
	}
	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}
	if list.HasMore && list.NextCursor != nil {
		fmt.Fprintf(out, "\nMore results — next page: --cursor %s\n", *list.NextCursor)
	}
	return nil
}

func PrintAgentSession(out io.Writer, s *deployment.AgentSession) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", s.ID)
	fmt.Fprintf(w, "Customer:\t%s\n", s.CustomerID)
	if s.Title != nil {
		fmt.Fprintf(w, "Title:\t%s\n", *s.Title)
	}
	fmt.Fprintf(w, "Mode:\t%s\n", s.Mode)
	if len(s.Labels) > 0 {
		fmt.Fprintf(w, "Labels:\t%s\n", strings.Join(s.Labels, ", "))
	}
	fmt.Fprintf(w, "Created:\t%s\n", LocalTime(s.CreatedAt))
	return w.Flush()
}

func PrintAgentEventList(out io.Writer, events []deployment.AgentEvent) error {
	if len(events) == 0 {
		fmt.Fprintln(out, "No events found.")
		return nil
	}
	headers := []string{"OFFSET", "ID", "KIND", "SOURCE", "CREATED", "SUMMARY"}
	rows := make([][]string, len(events))
	for i, e := range events {
		rows[i] = []string{
			fmt.Sprintf("%d", e.Offset),
			e.ID,
			e.Kind,
			e.Source,
			LocalTime(e.CreatedAt),
			truncateString(firstLine(EventSummary(e)), eventSummaryWidth),
		}
	}
	return PrintTable(out, headers, rows)
}

func PrintAgentEvent(out io.Writer, e *deployment.AgentEvent) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "ID:\t%s\n", e.ID)
	fmt.Fprintf(w, "Offset:\t%d\n", e.Offset)
	fmt.Fprintf(w, "Kind:\t%s\n", e.Kind)
	fmt.Fprintf(w, "Source:\t%s\n", e.Source)
	fmt.Fprintf(w, "Created:\t%s\n", LocalTime(e.CreatedAt))
	fmt.Fprintf(w, "Trace:\t%s\n", e.TraceID)
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintln(out)
	fmt.Fprintln(out, EventSummary(*e))
	return nil
}

// EventSummary is what an event says: a message's text, a status's state, and
// any other event's data as JSON.
func EventSummary(e deployment.AgentEvent) string {
	if data, ok := e.Message(); ok {
		return data.Message
	}
	if status := e.Status(); status != "" {
		return status
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, e.Data); err != nil {
		return string(e.Data)
	}
	return compact.String()
}

// EventLine is an event as one transcript line: who said what, or what happened.
func EventLine(e deployment.AgentEvent, labels map[string]string) string {
	switch e.Kind {
	case deployment.EventKindMessage:
		return sourceLabel(e.Source, labels) + ": " + EventSummary(e)
	case deployment.EventKindStatus:
		return "(" + EventSummary(e) + ")"
	}
	return e.Kind + " " + e.Source + ": " + EventSummary(e)
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
