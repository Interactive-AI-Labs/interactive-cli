package inputs

import (
	"fmt"
	"strings"
)

// ValidateSecretValue validates a secret value.
// It checks that the value is not empty/whitespace and is not wrapped in quotes.
func ValidateSecretValue(key, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value for key %q cannot be empty", key)
	}
	if isQuoted(value, '"') {
		return fmt.Errorf("value for key %q should not be wrapped in double quotes", key)
	}
	if isQuoted(value, '\'') {
		return fmt.Errorf("value for key %q should not be wrapped in single quotes", key)
	}
	return nil
}

// IsValidEnvVarName validates that a string is a valid environment variable name.
// Environment variable names must start with a letter or underscore,
// and contain only letters, digits, or underscores.
func IsValidEnvVarName(name string) bool {
	if len(name) == 0 {
		return false
	}

	firstChar := name[0]
	if !((firstChar >= 'A' && firstChar <= 'Z') || (firstChar >= 'a' && firstChar <= 'z') || firstChar == '_') {
		return false
	}

	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}

// isQuoted returns true if the string starts and ends with the given quote character
// and is at least 2 characters long.
func isQuoted(s string, quote byte) bool {
	if len(s) < 2 {
		return false
	}
	return s[0] == quote && s[len(s)-1] == quote
}
