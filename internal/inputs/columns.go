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

var validSortOrders = []string{"asc", "desc"}

func ValidateSorting(sortBy, sortOrder string, allowedFields []string) error {
	if sortBy != "" && !slices.Contains(allowedFields, sortBy) {
		return fmt.Errorf(
			"invalid sort-by %q: must be one of %s",
			sortBy,
			strings.Join(allowedFields, ", "),
		)
	}
	if sortOrder != "" && !slices.Contains(validSortOrders, sortOrder) {
		return fmt.Errorf(
			"invalid sort-order %q: must be one of %s",
			sortOrder,
			strings.Join(validSortOrders, ", "),
		)
	}
	return nil
}
