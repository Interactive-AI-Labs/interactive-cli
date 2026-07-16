package inputs

import (
	"fmt"
	"slices"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var validRouterModelRegions = []string{"us", "eu"}

// ValidateRouterModelListOptions validates router model list options. This
// endpoint is 0-indexed (page >= 0) with a max limit of 100, so it cannot reuse
// ValidatePagination.
func ValidateRouterModelListOptions(opts clients.RouterModelListOptions) error {
	if opts.Page < 0 {
		return fmt.Errorf("page must be >= 0, got %d", opts.Page)
	}
	if opts.Limit < 1 || opts.Limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100, got %d", opts.Limit)
	}
	if opts.Region != "" && !slices.Contains(validRouterModelRegions, opts.Region) {
		return fmt.Errorf("invalid region %q: must be one of us, eu", opts.Region)
	}
	return nil
}
