package inputs

import (
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var DefaultSessionColumns = []string{
	"id",
	"created_at",
	"environment",
	"trace_count",
	"duration_seconds",
	"total_cost",
	"total_tokens",
}

var AllSessionColumns = []string{
	"id",
	"created_at",
	"updated_at",
	"environment",
	"user_id",
	"trace_count",
	"duration_seconds",
	"total_cost",
	"input_tokens",
	"output_tokens",
	"total_tokens",
}

func ValidateSessionListOptions(opts clients.SessionListOptions) error {
	if err := validateTimestamp(opts.FromTimestamp, "from-timestamp"); err != nil {
		return err
	}
	if err := validateTimestamp(opts.ToTimestamp, "to-timestamp"); err != nil {
		return err
	}
	if opts.Page < 1 {
		return fmt.Errorf("page must be >= 1, got %d", opts.Page)
	}
	if opts.Limit < 0 {
		return fmt.Errorf("limit must be non-negative, got %d", opts.Limit)
	}
	return nil
}
