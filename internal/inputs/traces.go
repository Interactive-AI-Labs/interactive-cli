package inputs

import (
	"fmt"
	"regexp"
	"strings"
)

// traceIDRegex matches hex trace IDs (32-char) and UUIDs (with or without hyphens).
var traceIDRegex = regexp.MustCompile(`^[a-fA-F0-9-]{1,64}$`)

var DefaultTraceColumns = []string{"id", "name", "timestamp", "latency", "cost", "tags"}

var AllTraceColumns = []string{"id", "name", "timestamp", "user_id", "session_id", "release", "version", "environment", "public", "latency", "cost", "tags"}

func ValidateTraceID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("trace ID cannot be empty")
	}
	if !traceIDRegex.MatchString(id) {
		return fmt.Errorf("trace ID %q is not valid", id)
	}
	return nil
}

func ValidateTraceColumns(columns []string) error {
	valid := make(map[string]bool, len(AllTraceColumns))
	for _, col := range AllTraceColumns {
		valid[col] = true
	}
	for _, col := range columns {
		if !valid[col] {
			return fmt.Errorf("unknown column %q (available: %s)", col, strings.Join(AllTraceColumns, ", "))
		}
	}
	return nil
}
