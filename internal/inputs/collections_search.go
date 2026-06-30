package inputs

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ParseVector parses a comma-separated list of floats (e.g. "0.1,0.2,0.3").
func ParseVector(s string) ([]float64, error) {
	parts := strings.Split(s, ",")
	vec := make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		f, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid vector component %q: %w", p, err)
		}
		vec = append(vec, f)
	}
	if len(vec) == 0 {
		return nil, fmt.Errorf("--vector is empty")
	}
	return vec, nil
}

// BuildSearchBody builds a single-lane search body from exactly one of query
// (text, embedded server-side) or vector.
func BuildSearchBody(
	query string,
	vector []float64,
	using string,
	limit int,
	filterJSON string,
) ([]byte, error) {
	if (query == "") == (len(vector) == 0) {
		return nil, fmt.Errorf("provide exactly one of --query or --vector")
	}

	body := map[string]any{}
	if query != "" {
		body["query"] = query
	} else {
		body["vector"] = vector
	}
	if using != "" {
		body["using"] = using
	}
	if limit > 0 {
		body["limit"] = limit
	}
	if err := addFilter(body, filterJSON); err != nil {
		return nil, err
	}
	return json.Marshal(body)
}

// BuildQueryByIDBody builds a query-by-id body.
func BuildQueryByIDBody(
	id, using string,
	limit int,
	excludeSelf bool,
	filterJSON string,
) ([]byte, error) {
	if id == "" {
		return nil, fmt.Errorf("--id is required")
	}

	body := map[string]any{"id": id, "exclude_self": excludeSelf}
	if using != "" {
		body["using"] = using
	}
	if limit > 0 {
		body["limit"] = limit
	}
	if err := addFilter(body, filterJSON); err != nil {
		return nil, err
	}
	return json.Marshal(body)
}

// addFilter parses a JSON object string into body["filter"] when non-empty.
func addFilter(body map[string]any, filterJSON string) error {
	if filterJSON == "" {
		return nil
	}
	var filter map[string]any
	if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
		return fmt.Errorf("--filter must be a JSON object: %w", err)
	}
	body["filter"] = filter
	return nil
}
