package inputs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ReplayInput struct {
	Dataset     string
	Scenarios   []string
	File        string
	RunID       string
	Repeat      int
	Concurrency int
}

// LoadScenarioFile reads a YAML or JSON scenario document into a generic map.
// The agent owns the schema; the only local default is the scenario name,
// taken from the file name when the document has none.
func LoadScenarioFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file %q: %w", path, err)
	}
	var doc any
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse scenario file %q: %w", path, err)
	}
	body, ok := doc.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("scenario file %q must be a YAML or JSON object", path)
	}
	if _, ok := body["scenario"]; !ok {
		base := filepath.Base(path)
		body["scenario"] = strings.TrimSuffix(base, filepath.Ext(base))
	}
	return body, nil
}
