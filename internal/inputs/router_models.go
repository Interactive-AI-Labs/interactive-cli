package inputs

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

// RouterModelRegions are the advertised --region values; "us" is still accepted for "global".
var RouterModelRegions = []string{regionGlobal, platform.RouterModelRegionEU}

const regionGlobal = "global"

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
	case regionGlobal:
		opts.Region = platform.RouterModelRegionUS
	default:
		return opts, fmt.Errorf(
			"invalid region %q: must be one of %s",
			opts.Region,
			strings.Join(RouterModelRegions, ", "),
		)
	}
	return opts, nil
}
