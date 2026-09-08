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

// ValidateReplayInput checks the flag combination and the ranges the agent
// enforces. Pairwise exclusions are declared on the cobra command; this covers
// what cobra cannot express.
func ValidateReplayInput(in ReplayInput) error {
	if len(in.Scenarios) > 0 && in.Dataset == "" {
		return fmt.Errorf("--scenarios requires --dataset")
	}
	if in.Dataset == "" && in.File == "" && in.RunID == "" {
		return fmt.Errorf("one of --dataset, --file, or --run-id is required")
	}
	// The agent enforces the same ranges and returns 422 when they are exceeded;
	// these copies exist for early feedback and must stay in sync with it.
	if in.Repeat < 1 || in.Repeat > 20 {
		return fmt.Errorf("--repeat must be between 1 and 20")
	}
	if in.Concurrency < 1 || in.Concurrency > 32 {
		return fmt.Errorf("--concurrency must be between 1 and 32")
	}
	return nil
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

// RuntimeAPIKeyRef returns the env var name referenced by
// agentConfig.runtime.api_key (e.g. "AGENT_API_KEY" for "${AGENT_API_KEY}"),
// or "" when the config has no such reference.
func RuntimeAPIKeyRef(agentConfig any) string {
	cfg, ok := agentConfig.(map[string]any)
	if !ok {
		return ""
	}
	runtime, ok := cfg["runtime"].(map[string]any)
	if !ok {
		return ""
	}
	ref, ok := runtime["api_key"].(string)
	if !ok || !strings.HasPrefix(ref, "${") || !strings.HasSuffix(ref, "}") {
		return ""
	}
	return strings.TrimSuffix(strings.TrimPrefix(ref, "${"), "}")
}
