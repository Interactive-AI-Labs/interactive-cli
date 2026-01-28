package inputs

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateSecretValue validates a secret value.
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

// ValidateSecretKey validates a secret key name.
func ValidateSecretKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("key name cannot be empty")
	}

	if !isValidEnvVarName(key) {
		return fmt.Errorf("key name %q is not a valid environment variable name", key)
	}

	return nil
}

// envVarNameRegex matches valid POSIX environment variable names:
// must start with a letter or underscore, followed by letters, digits, or underscores.
var envVarNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// isValidEnvVarName validates that a string is a valid environment variable name.
func isValidEnvVarName(name string) bool {
	return envVarNameRegex.MatchString(name)
}

// isQuoted returns true if the string starts and ends with the given quote character
func isQuoted(s string, quote byte) bool {
	if len(s) < 2 {
		return false
	}
	return s[0] == quote && s[len(s)-1] == quote
}
