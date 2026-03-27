package inputs

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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

func ValidateDatasetItemListOptions(opts clients.DatasetItemListOptions) error {
	if strings.TrimSpace(opts.DatasetName) == "" {
		return fmt.Errorf("--dataset-name is required")
	}
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
) (clients.DatasetItemCreateBody, error) {
	datasetName := strings.TrimSpace(input.DatasetName)
	if datasetName == "" {
		return clients.DatasetItemCreateBody{}, fmt.Errorf("--dataset-name is required")
	}

	body := clients.DatasetItemCreateBody{
		ID:                  strings.TrimSpace(input.ID),
		DatasetName:         datasetName,
		SourceTraceID:       strings.TrimSpace(input.SourceTraceID),
		SourceObservationID: strings.TrimSpace(input.SourceObservationID),
		Status:              strings.TrimSpace(input.Status),
	}

	if strings.TrimSpace(input.InputJSON) != "" {
		v, err := parseJSONAny(input.InputJSON, "--input")
		if err != nil {
			return clients.DatasetItemCreateBody{}, err
		}
		body.Input = v
	}

	if strings.TrimSpace(input.ExpectedOutputJSON) != "" {
		v, err := parseJSONAny(input.ExpectedOutputJSON, "--expected-output")
		if err != nil {
			return clients.DatasetItemCreateBody{}, err
		}
		body.ExpectedOutput = v
	}

	if strings.TrimSpace(input.MetadataJSON) != "" {
		metadata, err := parseJSONObject(input.MetadataJSON)
		if err != nil {
			return clients.DatasetItemCreateBody{}, err
		}
		body.Metadata = metadata
	}

	return body, nil
}
