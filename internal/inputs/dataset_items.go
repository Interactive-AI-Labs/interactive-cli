package inputs

import (
	"errors"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

var DefaultDatasetItemColumns = []string{
	"id",
	"name",
	"status",
	"dataset_name",
	"source_trace_id",
	"created_at",
}

var AllDatasetItemColumns = []string{
	"id",
	"name",
	"status",
	"dataset_name",
	"source_trace_id",
	"source_observation_id",
	"created_at",
	"updated_at",
}

func ValidateDatasetItemListOptions(opts platform.DatasetItemListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

type DatasetItemCreateInput struct {
	ID                  string
	Name                string
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
) (platform.DatasetItemCreateBody, error) {
	body := platform.DatasetItemCreateBody{
		ID:                  strings.TrimSpace(input.ID),
		Name:                strings.TrimSpace(input.Name),
		DatasetName:         strings.TrimSpace(input.DatasetName),
		SourceTraceID:       strings.TrimSpace(input.SourceTraceID),
		SourceObservationID: strings.TrimSpace(input.SourceObservationID),
		Status:              strings.TrimSpace(input.Status),
	}

	if strings.TrimSpace(input.InputJSON) != "" {
		v, err := parseJSONAny(input.InputJSON, "--input")
		if err != nil {
			return platform.DatasetItemCreateBody{}, err
		}
		body.Input = v
	}

	if strings.TrimSpace(input.ExpectedOutputJSON) != "" {
		v, err := parseJSONAny(input.ExpectedOutputJSON, "--expected-output")
		if err != nil {
			return platform.DatasetItemCreateBody{}, err
		}
		body.ExpectedOutput = v
	}

	if strings.TrimSpace(input.MetadataJSON) != "" {
		metadata, err := parseJSONObject(
			input.MetadataJSON,
			"--metadata-json",
		)
		if err != nil {
			return platform.DatasetItemCreateBody{}, err
		}
		body.Metadata = metadata
	}

	return body, nil
}

type DatasetItemUpdateInput struct {
	Name         *string
	Status       string
	MetadataJSON string
}

func BuildDatasetItemUpdateBody(
	input DatasetItemUpdateInput,
) (platform.DatasetItemUpdateBody, error) {
	body := platform.DatasetItemUpdateBody{}

	if input.Name != nil {
		if name := strings.TrimSpace(*input.Name); name != "" {
			body["name"] = name
		} else {
			body["name"] = nil
		}
	}

	if status := strings.TrimSpace(input.Status); status != "" {
		body["status"] = strings.ToUpper(status)
	}

	if strings.TrimSpace(input.MetadataJSON) != "" {
		metadata, err := parseJSONObject(input.MetadataJSON, "--metadata-json")
		if err != nil {
			return nil, err
		}
		body["metadata"] = metadata
	}

	if len(body) == 0 {
		return nil, errors.New("nothing to update: pass --name, --status or --metadata-json")
	}
	return body, nil
}
