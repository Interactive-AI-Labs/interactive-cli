package inputs

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

const maxTraceIDLength = 256

var DefaultTraceColumns = []string{"id", "name", "timestamp", "latency", "cost", "tags"}

var AllTraceColumns = []string{
	"id",
	"name",
	"timestamp",
	"user_id",
	"session_id",
	"release",
	"version",
	"environment",
	"public",
	"latency",
	"cost",
	"tags",
}

func ValidateTraceID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("trace ID cannot be empty")
	}
	if len(id) > maxTraceIDLength {
		return fmt.Errorf("trace ID is too long (max %d characters)", maxTraceIDLength)
	}
	return nil
}

var validOrderByFields = []string{
	"id",
	"name",
	"timestamp",
	"userId",
	"sessionId",
	"release",
	"version",
	"public",
}

var validOrderByDirections = []string{"asc", "desc"}

func ValidateTraceListOptions(opts clients.TraceListOptions) error {
	if opts.Page < 1 {
		return fmt.Errorf("page must be >= 1, got %d", opts.Page)
	}
	if opts.Limit < 0 {
		return fmt.Errorf("limit must be non-negative, got %d", opts.Limit)
	}
	if err := validateTimestamp(opts.FromTimestamp, "from-timestamp"); err != nil {
		return err
	}
	if err := validateTimestamp(opts.ToTimestamp, "to-timestamp"); err != nil {
		return err
	}
	if err := validateOrderBy(opts.OrderBy); err != nil {
		return err
	}
	return nil
}

func validateTimestamp(value, name string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse(time.RFC3339, value); err != nil {
		return fmt.Errorf(
			"invalid %s %q: must be ISO 8601 format (e.g. 2025-01-01T00:00:00Z)",
			name,
			value,
		)
	}
	return nil
}

func validateOrderBy(value string) error {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return fmt.Errorf(
			"invalid order-by %q: must be field.direction (e.g. timestamp.desc)",
			value,
		)
	}
	if !slices.Contains(validOrderByFields, parts[0]) {
		return fmt.Errorf(
			"invalid order-by field %q (available: %s)",
			parts[0],
			strings.Join(validOrderByFields, ", "),
		)
	}
	if !slices.Contains(validOrderByDirections, parts[1]) {
		return fmt.Errorf("invalid order-by direction %q: must be asc or desc", parts[1])
	}
	return nil
}

func ValidateTraceColumns(columns []string) error {
	for _, col := range columns {
		if !slices.Contains(AllTraceColumns, col) {
			return fmt.Errorf(
				"unknown column %q (available: %s)",
				col,
				strings.Join(AllTraceColumns, ", "),
			)
		}
	}
	return nil
}
