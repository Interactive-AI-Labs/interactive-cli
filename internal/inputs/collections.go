package inputs

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ReadCollectionBodyFile reads a YAML or JSON file and returns it as JSON bytes
// suitable for a collection create/patch request body. YAML is a superset of
// JSON, so both parse; yaml.v3 decodes mappings into string-keyed maps, so the
// result re-marshals cleanly to JSON.
func ReadCollectionBodyFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", path, err)
	}

	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse %q as YAML/JSON: %w", path, err)
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to encode config: %w", err)
	}
	return body, nil
}

// BuildChunkCountBody builds the count request body from an optional metadata
// filter (a JSON object string) and an optional id prefix.
func BuildChunkCountBody(filterJSON, prefix string) ([]byte, error) {
	body := map[string]any{}
	if filterJSON != "" {
		var filter map[string]any
		if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
			return nil, fmt.Errorf("--filter must be a JSON object: %w", err)
		}
		body["filter"] = filter
	}
	if prefix != "" {
		body["prefix"] = prefix
	}
	return json.Marshal(body)
}

// BuildBulkDeleteBody builds the bulk-delete body from exactly one selector:
// ids, a metadata filter (JSON object string), or all.
func BuildBulkDeleteBody(ids []string, filterJSON string, all bool) ([]byte, error) {
	set := 0
	if len(ids) > 0 {
		set++
	}
	if filterJSON != "" {
		set++
	}
	if all {
		set++
	}
	if set == 0 {
		return nil, fmt.Errorf("provide exactly one of --ids, --filter, or --all")
	}
	if set > 1 {
		return nil, fmt.Errorf(
			"provide exactly one of --ids, --filter, or --all (got %d selectors)",
			set,
		)
	}

	switch {
	case len(ids) > 0:
		return json.Marshal(map[string]any{"ids": ids})
	case filterJSON != "":
		var filter map[string]any
		if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
			return nil, fmt.Errorf("--filter must be a JSON object: %w", err)
		}
		return json.Marshal(map[string]any{"filter": filter})
	default:
		return json.Marshal(map[string]any{"all": true})
	}
}

// BuildAddSlotBody builds an add-slot body from flags (a raw vector slot).
func BuildAddSlotBody(slotType string, dimension int, distance string) ([]byte, error) {
	if dimension <= 0 {
		return nil, fmt.Errorf("--dimension must be a positive integer (or use --file)")
	}
	body := map[string]any{"type": slotType, "dimension": dimension}
	if distance != "" {
		body["distance"] = distance
	}
	return json.Marshal(body)
}
