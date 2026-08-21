package inputs

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

var DefaultDatasetItemColumns = []string{
	"id",
	"status",
	"dataset_name",
	"source_trace_id",
	"created_at",
}

var AllDatasetItemColumns = []string{
	"id",
	"status",
	"dataset_name",
	"source_trace_id",
	"source_observation_id",
	"created_at",
	"updated_at",
}

func ValidateDatasetItemListOptions(opts api.DatasetItemListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

type DatasetItemCreateInput struct {
	ID                  string
	DatasetName         string
	InputJSON           string
	ExpectedOutputJSON  string
	MetadataJSON        string
	SourceTraceID       string
	SourceObservationID string
	Status              string
}

func BuildDatasetItemCreateBody(
	input DatasetItemCreateInput,
) (api.DatasetItemCreateBody, error) {
	body := api.DatasetItemCreateBody{
		ID:                  strings.TrimSpace(input.ID),
		DatasetName:         strings.TrimSpace(input.DatasetName),
		SourceTraceID:       strings.TrimSpace(input.SourceTraceID),
		SourceObservationID: strings.TrimSpace(input.SourceObservationID),
		Status:              strings.TrimSpace(input.Status),
	}

	if strings.TrimSpace(input.InputJSON) != "" {
		v, err := parseJSONAny(input.InputJSON, "--input")
		if err != nil {
			return api.DatasetItemCreateBody{}, err
		}
		body.Input = v
	}

	if strings.TrimSpace(input.ExpectedOutputJSON) != "" {
		v, err := parseJSONAny(input.ExpectedOutputJSON, "--expected-output")
		if err != nil {
			return api.DatasetItemCreateBody{}, err
		}
		body.ExpectedOutput = v
	}

	if strings.TrimSpace(input.MetadataJSON) != "" {
		metadata, err := parseJSONObject(
			input.MetadataJSON,
			"--metadata-json",
		)
		if err != nil {
			return api.DatasetItemCreateBody{}, err
		}
		body.Metadata = metadata
	}

	return body, nil
}
