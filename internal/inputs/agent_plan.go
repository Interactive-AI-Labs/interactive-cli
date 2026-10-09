package inputs

import (
	"slices"
	"strings"
)

// PendingMcpRefs lists changing MCP references that prevent planning an agent against live state.
// Dependency lookup follows the server: only bare names and ref attach an MCP, not an inline entry's id.
func PendingMcpRefs(agentConfig any, changing []string) []string {
	config, _ := agentConfig.(map[string]any)
	entries, _ := config["mcps"].([]any)
	var pending []string
	for _, entry := range entries {
		var name string
		switch value := entry.(type) {
		case string:
			name = value
		case map[string]any:
			name, _ = value["ref"].(string)
		}
		name = strings.TrimSpace(name)
		if name != "" && slices.Contains(changing, name) {
			pending = append(pending, name)
		}
	}
	slices.Sort(pending)
	return slices.Compact(pending)
}
