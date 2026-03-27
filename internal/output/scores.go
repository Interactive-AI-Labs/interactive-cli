package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var scoreColumnMap = map[string]struct {
	Header string
	Value  func(s *clients.ScoreInfo) string
}{
	"id":        {"ID", func(s *clients.ScoreInfo) string { return s.ID }},
	"name":      {"NAME", func(s *clients.ScoreInfo) string { return s.Name }},
	"data_type": {"DATA TYPE", func(s *clients.ScoreInfo) string { return s.DataType }},
	"value":     {"VALUE", func(s *clients.ScoreInfo) string { return formatScoreValue(s.Value) }},
	"source":    {"SOURCE", func(s *clients.ScoreInfo) string { return s.Source }},
	"timestamp": {"TIMESTAMP", func(s *clients.ScoreInfo) string { return LocalTime(s.Timestamp) }},
	"trace_id":  {"TRACE ID", func(s *clients.ScoreInfo) string { return s.TraceID }},
	"observation_id": {
		"OBSERVATION ID",
		func(s *clients.ScoreInfo) string { return s.ObservationID },
	},
	"session_id": {"SESSION ID", func(s *clients.ScoreInfo) string { return s.SessionID }},
	"environment": {
		"ENVIRONMENT",
		func(s *clients.ScoreInfo) string { return s.Environment },
	},
	"config_id": {"CONFIG ID", func(s *clients.ScoreInfo) string { return s.ConfigID }},
	"user_id":   {"USER ID", func(s *clients.ScoreInfo) string { return s.UserID }},
	"comment":   {"COMMENT", func(s *clients.ScoreInfo) string { return s.Comment }},
}

func PrintScoreList(
	out io.Writer,
	scores []clients.ScoreInfo,
	meta clients.CursorMeta,
	columns []string,
) error {
	if len(scores) == 0 {
		fmt.Fprintln(out, "No scores found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := scoreColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(scores))
	for i, score := range scores {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := scoreColumnMap[col]; ok {
				row[j] = def.Value(&score)
			}
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	if meta.NextCursor != "" {
		fmt.Fprintf(out, "\nNext cursor: %s\n", meta.NextCursor)
	}

	return nil
}

func PrintScoreCreateResult(out io.Writer, score *clients.ScoreInfo) error {
	fmt.Fprintf(out, "Created score %q.\n", score.ID)
	fmt.Fprintf(out, "Name:       %s\n", score.Name)
	fmt.Fprintf(out, "Data Type:  %s\n", score.DataType)
	fmt.Fprintf(out, "Value:      %s\n", formatScoreValue(score.Value))
	fmt.Fprintf(out, "Timestamp:  %s\n", LocalTime(score.Timestamp))
	if score.TraceID != "" {
		fmt.Fprintf(out, "Trace ID:   %s\n", score.TraceID)
	}
	if score.ObservationID != "" {
		fmt.Fprintf(out, "Observation ID: %s\n", score.ObservationID)
	}
	if score.SessionID != "" {
		fmt.Fprintf(out, "Session ID: %s\n", score.SessionID)
	}
	return nil
}

func PrintDeleteSuccess(out io.Writer, resourceID, resourceType, message string) error {
	if strings.TrimSpace(message) != "" {
		_, err := fmt.Fprintln(out, message)
		return err
	}

	_, err := fmt.Fprintf(out, "Deleted %s %q.\n", resourceType, resourceID)
	return err
}

func formatScoreValue(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return "-"
	}

	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		return stringValue
	}

	var floatValue float64
	if err := json.Unmarshal(raw, &floatValue); err == nil {
		return strconv.FormatFloat(floatValue, 'f', -1, 64)
	}

	var boolValue bool
	if err := json.Unmarshal(raw, &boolValue); err == nil {
		return strconv.FormatBool(boolValue)
	}

	return string(raw)
}
