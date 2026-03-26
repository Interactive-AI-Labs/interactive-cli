package inputs

import (
	"fmt"
	"slices"
	"strings"
)

// ValidateColumns checks that every entry in columns exists in the allowed set.
func ValidateColumns(columns, allowed []string) error {
	for _, col := range columns {
		if !slices.Contains(allowed, col) {
			return fmt.Errorf(
				"unknown column %q (available: %s)",
				col,
				strings.Join(allowed, ", "),
			)
		}
	}
	return nil
}
