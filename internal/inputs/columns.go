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

// ValidatePagination validates common page-based pagination parameters.
func ValidatePagination(page, limit int) error {
	if page < 1 {
		return fmt.Errorf("page must be >= 1, got %d", page)
	}
	if limit < 0 {
		return fmt.Errorf("limit must be non-negative, got %d", limit)
	}
	return nil
}
