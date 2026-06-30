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
	if path == "" {
		return nil, fmt.Errorf("a config file is required; provide --file")
	}

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
