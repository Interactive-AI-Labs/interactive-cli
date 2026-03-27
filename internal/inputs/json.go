package inputs

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parseJSONObject parses a JSON string into a map. Returns nil if raw is empty.
func parseJSONObject(raw, flagName string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf(
			"invalid %s: must be a valid JSON object",
			flagName,
		)
	}
	if value == nil {
		return nil, fmt.Errorf(
			"invalid %s: must be a JSON object",
			flagName,
		)
	}
	return value, nil
}

// parseJSONAny parses a JSON string into any Go value. Returns nil if raw is empty.
func parseJSONAny(raw, flagName string) (any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf(
			"invalid %s: must be valid JSON",
			flagName,
		)
	}
	return value, nil
}

// parseJSONArray parses a JSON string into an array. Returns nil if raw is empty.
func parseJSONArray(raw, flagName string) (json.RawMessage, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var value []any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf(
			"invalid %s: must be a valid JSON array",
			flagName,
		)
	}
	if value == nil {
		return nil, fmt.Errorf(
			"invalid %s: must be a valid JSON array",
			flagName,
		)
	}
	return json.RawMessage(raw), nil
}
