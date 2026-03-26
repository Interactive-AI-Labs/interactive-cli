package inputs

import (
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var DefaultMetricsDailyColumns = []string{
	"date",
	"count_traces",
	"count_observations",
	"total_cost",
}

var AllMetricsDailyColumns = []string{
	"date",
	"count_traces",
	"count_observations",
	"total_cost",
	"total_tokens",
}

func ValidateMetricsDailyOptions(opts clients.MetricsDailyOptions) error {
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
