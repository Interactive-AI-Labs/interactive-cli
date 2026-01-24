package inputs

import (
	"fmt"
	"strings"
)

// ValidateSecretValue checks that a secret value does not start and end with quotes.
// This prevents user confusion since Kubernetes treats quoted values as literal strings
// (including the quotes), unlike shell environment variables which strip them.
func ValidateSecretValue(key, value string) error {
	if isQuoted(value, '"') {
		return fmt.Errorf("secret value for key %q should not be wrapped in double quotes", key)
	}
	if isQuoted(value, '\'') {
		return fmt.Errorf("secret value for key %q should not be wrapped in single quotes", key)
	}
	return nil
}

// isQuoted returns true if the string starts and ends with the given quote character
// and is at least 2 characters long.
func isQuoted(s string, quote byte) bool {
	if len(s) < 2 {
		return false
	}
	return s[0] == quote && s[len(s)-1] == quote
}

// ValidateSecretData validates all key-value pairs in a secret data map.
func ValidateSecretData(data map[string]string) error {
	var errors []string
	for key, value := range data {
		if err := ValidateSecretValue(key, value); err != nil {
			errors = append(errors, err.Error())
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("invalid secret values:\n  %s", strings.Join(errors, "\n  "))
	}
	return nil
}
