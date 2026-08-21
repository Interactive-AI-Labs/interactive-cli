package inputs

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

var DefaultDatasetColumns = []string{
	"id",
	"name",
	"description",
	"created_at",
}

var AllDatasetColumns = []string{
	"id",
	"name",
	"description",
	"created_at",
	"updated_at",
}

func ValidateDatasetListOptions(opts api.DatasetListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildDatasetCreateBody(
	name, description, metadataJSON string,
) (api.DatasetCreateBody, error) {
	body := api.DatasetCreateBody{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
	}

	if strings.TrimSpace(metadataJSON) != "" {
		metadata, err := parseJSONObject(
			metadataJSON,
			"--metadata-json",
		)
		if err != nil {
			return api.DatasetCreateBody{}, err
		}
		body.Metadata = metadata
	}

	return body, nil
}
