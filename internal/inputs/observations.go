package inputs

import (
	"fmt"
	"slices"
	"strings"
)

const maxObservationIDLength = 256

var DefaultObservationColumns = []string{
	"id", "type", "name", "model", "latency_ms", "total_cost", "total_tokens",
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

func ValidateObservationID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("observation ID cannot be empty")
	}
	if len(id) > maxObservationIDLength {
		return fmt.Errorf("observation ID is too long (max %d characters)", maxObservationIDLength)
	}
	return nil
}

func ValidateObservationColumns(columns []string) error {
	for _, col := range columns {
		if !slices.Contains(AllObservationColumns, col) {
			return fmt.Errorf(
				"unknown column %q (available: %s)",
				col,
				strings.Join(AllObservationColumns, ", "),
			)
		}
	}
	return nil
}
