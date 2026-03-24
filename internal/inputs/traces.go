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
	"observation_count",
	"input_tokens",
	"output_tokens",
	"total_tokens",
	"level",
}

var (
	validOrderByFields   = []string{"timestamp", "latency", "cost", "name"}
	validOrderDirections = []string{"asc", "desc"}
	validLevels          = []string{"DEBUG", "DEFAULT", "WARNING", "ERROR"}
	validFieldGroups     = []string{"core", "io", "metrics"}
)

const maxSearchLength = 200

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
	if err := validateOrder(opts.Order); err != nil {
		return err
	}
	if opts.MinCost != nil && *opts.MinCost < 0 {
		return fmt.Errorf("--min-cost must be >= 0")
	}
	if opts.MaxCost != nil && *opts.MaxCost < 0 {
		return fmt.Errorf("--max-cost must be >= 0")
	}
	if opts.MinCost != nil && opts.MaxCost != nil && *opts.MinCost > *opts.MaxCost {
		return fmt.Errorf("--min-cost cannot be greater than --max-cost")
	}
	if opts.MinLatency != nil && *opts.MinLatency < 0 {
		return fmt.Errorf("--min-latency must be >= 0")
	}
	if opts.MaxLatency != nil && *opts.MaxLatency < 0 {
		return fmt.Errorf("--max-latency must be >= 0")
	}
	if opts.MinLatency != nil && opts.MaxLatency != nil && *opts.MinLatency > *opts.MaxLatency {
		return fmt.Errorf("--min-latency cannot be greater than --max-latency")
	}
	if opts.MinTokens != nil && *opts.MinTokens < 0 {
		return fmt.Errorf("--min-tokens must be >= 0")
	}
	if opts.MaxTokens != nil && *opts.MaxTokens < 0 {
		return fmt.Errorf("--max-tokens must be >= 0")
	}
	if opts.MinTokens != nil && opts.MaxTokens != nil && *opts.MinTokens > *opts.MaxTokens {
		return fmt.Errorf("--min-tokens cannot be greater than --max-tokens")
	}
	if err := validateLevel(opts.Level); err != nil {
		return err
	}
	if len(opts.Search) > maxSearchLength {
		return fmt.Errorf(
			"--search must be at most %d characters, got %d",
			maxSearchLength,
			len(opts.Search),
		)
	}
	if err := ValidateFieldGroups(opts.Fields); err != nil {
		return err
	}
	return nil
}

func ValidateFieldGroups(fields string) error {
	if fields == "" {
		return nil
	}
	for _, f := range strings.Split(fields, ",") {
		f = strings.TrimSpace(f)
		if !slices.Contains(validFieldGroups, f) {
			return fmt.Errorf(
				"invalid field group %q (valid: %s)",
				f,
				strings.Join(validFieldGroups, ", "),
			)
		}
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
	if !slices.Contains(validOrderByFields, value) {
		return fmt.Errorf(
			"invalid --order-by %q (valid: %s)",
			value,
			strings.Join(validOrderByFields, ", "),
		)
	}
	return nil
}

func validateOrder(value string) error {
	if value == "" {
		return nil
	}
	if !slices.Contains(validOrderDirections, value) {
		return fmt.Errorf("invalid --order %q: must be asc or desc", value)
	}
	return nil
}

func validateLevel(value string) error {
	if value == "" {
		return nil
	}
	if !slices.Contains(validLevels, value) {
		return fmt.Errorf(
			"invalid --level %q (valid: %s)",
			value,
			strings.Join(validLevels, ", "),
		)
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
