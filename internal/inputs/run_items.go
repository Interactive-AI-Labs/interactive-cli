package inputs

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

var DefaultRunItemColumns = []string{
	"id",
	"dataset_run_name",
	"dataset_item_id",
	"trace_id",
	"observation_id",
	"created_at",
}

var AllRunItemColumns = []string{
	"id",
	"dataset_run_name",
	"dataset_item_id",
	"trace_id",
	"observation_id",
	"created_at",
	"updated_at",
}

func ValidateRunItemListOptions(opts api.DatasetRunItemListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

type RunItemCreateInput struct {
	RunName        string
	RunDescription string
	DatasetItemID  string
	TraceID        string
	ObservationID  string
	MetadataJSON   string
}

func BuildRunItemCreateBody(
	input RunItemCreateInput,
) (api.DatasetRunItemCreateBody, error) {
	body := api.DatasetRunItemCreateBody{
		RunName:        strings.TrimSpace(input.RunName),
		RunDescription: strings.TrimSpace(input.RunDescription),
		DatasetItemID:  strings.TrimSpace(input.DatasetItemID),
		TraceID:        strings.TrimSpace(input.TraceID),
		ObservationID:  strings.TrimSpace(input.ObservationID),
	}

	if strings.TrimSpace(input.MetadataJSON) != "" {
		metadata, err := parseJSONObject(
			input.MetadataJSON,
			"--metadata-json",
		)
		if err != nil {
			return api.DatasetRunItemCreateBody{}, err
		}
		body.Metadata = metadata
	}

	return body, nil
}
