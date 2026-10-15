package inputs

import (
	"fmt"
	"strings"
)

// ValidatePromptDelete rejects a --version/--label delete that would reach a wider target.
// The server decodes %2F, so these name segments route to other delete handlers.
func ValidatePromptDelete(name string, version int, label string, byVersion, byLabel bool) error {
	if byVersion && version < 1 {
		return fmt.Errorf("--version must be 1 or greater")
	}
	if byLabel && strings.TrimSpace(label) == "" {
		return fmt.Errorf("--label must not be empty")
	}
	if !byVersion && !byLabel {
		return nil
	}
	for _, segment := range strings.Split(name, "/") {
		switch segment {
		case "folders", "by-name", "versions":
			return fmt.Errorf(
				"cannot delete %q by --version or --label: its %q segment routes to a wider delete",
				name, segment,
			)
		}
	}
	return nil
}
