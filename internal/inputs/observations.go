package inputs

import (
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var DefaultObservationColumns = []string{
	"id", "type", "name", "model", "latency_ms", "total_cost", "total_tokens",
}

var DefaultStandaloneObservationColumns = []string{
	"id",
	"trace_id",
	"type",
	"name",
	"model",
	"latency_ms",
	"total_cost",
	"total_tokens",
}

var AllObservationColumns = []string{
	"id",
	"trace_id",
	"type",
	"name",
	"start_time",
	"end_time",
	"parent_observation_id",
	"level",
	"status_message",
	"model",
	"input_tokens",
	"output_tokens",
	"total_tokens",
	"total_cost",
	"latency_ms",
}

var AllStandaloneObservationColumns = []string{
	"id",
	"trace_id",
	"type",
	"name",
	"model",
	"environment",
	"user_id",
	"version",
	"start_time",
	"end_time",
	"parent_observation_id",
	"level",
	"status_message",
	"input_tokens",
	"output_tokens",
	"total_tokens",
	"total_cost",
	"latency_ms",
}

func ValidateObservationSearchOptions(opts clients.ObservationSearchOptions) error {
	if err := validateTimestamp(opts.FromTimestamp, "from-timestamp"); err != nil {
		return err
	}
	if err := validateTimestamp(opts.ToTimestamp, "to-timestamp"); err != nil {
		return err
	}
	if opts.Limit < 0 {
		return fmt.Errorf("limit must be non-negative, got %d", opts.Limit)
	}
	return nil
}
