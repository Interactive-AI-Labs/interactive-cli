package inputs

import (
	"fmt"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

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

const maxSearchLength = 200

// ValidateTraceListOptions validates structural constraints on trace list
// options. Enum-style validations (--level, --order-by, --order, --fields) are
// delegated to the server to avoid client/server divergence.
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
	if len(opts.Search) > maxSearchLength {
		return fmt.Errorf(
			"--search must be at most %d characters, got %d",
			maxSearchLength,
			len(opts.Search),
		)
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

func ValidateTraceDeleteInput(traceID string, ids []string) error {
	traceID = strings.TrimSpace(traceID)

	if traceID != "" && len(ids) > 0 {
		return fmt.Errorf("positional trace ID and --ids are mutually exclusive")
	}

	if traceID == "" && len(ids) == 0 {
		return fmt.Errorf("trace ID is required; provide a positional trace ID or --ids")
	}

	if len(ids) > 500 {
		return fmt.Errorf("bulk delete supports at most 500 trace IDs, got %d", len(ids))
	}

	return nil
}
