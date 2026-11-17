package inputs

import (
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

// PrepareRouterModelListOptions cannot reuse ValidatePagination: page is 0-indexed, limit max 100.
func PrepareRouterModelListOptions(
	opts platform.RouterModelListOptions,
) (platform.RouterModelListOptions, error) {
	if opts.Page < 0 {
		return opts, fmt.Errorf("page must be >= 0, got %d", opts.Page)
	}
	if opts.Limit < 1 || opts.Limit > 100 {
		return opts, fmt.Errorf("limit must be between 1 and 100, got %d", opts.Limit)
	}
	switch opts.Region {
	case "", platform.RouterModelRegionUS, platform.RouterModelRegionEU:
	case "global":
		opts.Region = platform.RouterModelRegionUS
	default:
		return opts, fmt.Errorf("invalid region %q: must be one of global, eu", opts.Region)
	}
	return opts, nil
}
