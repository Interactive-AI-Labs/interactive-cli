package inputs

import "github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"

var DefaultDatasetRunColumns = []string{
	"id",
	"name",
	"dataset_name",
	"created_at",
	"updated_at",
}

var AllDatasetRunColumns = []string{
	"id",
	"name",
	"description",
	"dataset_name",
	"created_at",
	"updated_at",
}

func ValidateDatasetRunListOptions(opts platform.DatasetRunListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}
