package inputs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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

func ValidateDatasetListOptions(opts clients.DatasetListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildDatasetCreateBody(
	name, description, metadataJSON string,
) (clients.DatasetCreateBody, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return clients.DatasetCreateBody{}, fmt.Errorf("dataset name is required")
	}

	body := clients.DatasetCreateBody{
		Name:        name,
		Description: strings.TrimSpace(description),
	}

	if strings.TrimSpace(metadataJSON) != "" {
		metadata, err := parseJSONObject(metadataJSON)
		if err != nil {
			return clients.DatasetCreateBody{}, err
		}
		body.Metadata = metadata
	}

	return body, nil
}

func parseJSONAny(raw, flagName string) (any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("invalid %s: must be valid JSON", flagName)
	}
	return value, nil
}
