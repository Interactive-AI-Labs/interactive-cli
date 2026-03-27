package inputs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var DefaultScoreConfigColumns = []string{
	"id",
	"name",
	"data_type",
	"is_archived",
	"created_at",
}

var AllScoreConfigColumns = []string{
	"id",
	"name",
	"data_type",
	"is_archived",
	"min_value",
	"max_value",
	"description",
	"created_at",
	"updated_at",
}

func ValidateScoreConfigListOptions(opts clients.ScoreConfigListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func parseJSONArray(raw, flagName string) (json.RawMessage, error) {
	var value []any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("invalid %s: must be a valid JSON array", flagName)
	}
	if value == nil {
		return nil, fmt.Errorf("invalid %s: must be a valid JSON array", flagName)
	}

	return json.RawMessage(raw), nil
}

type ScoreConfigCreateInput struct {
	Name           string
	DataType       string
	MinValue       *float64
	MaxValue       *float64
	CategoriesJSON string
	Description    string
}

func BuildScoreConfigCreateBody(
	input ScoreConfigCreateInput,
) (clients.ScoreConfigCreateBody, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return clients.ScoreConfigCreateBody{}, fmt.Errorf("--name is required")
	}

	dataType := strings.ToUpper(strings.TrimSpace(input.DataType))
	if dataType == "" {
		return clients.ScoreConfigCreateBody{}, fmt.Errorf("--data-type is required")
	}

	body := clients.ScoreConfigCreateBody{
		Name:        name,
		DataType:    dataType,
		MinValue:    input.MinValue,
		MaxValue:    input.MaxValue,
		Description: strings.TrimSpace(input.Description),
	}

	if strings.TrimSpace(input.CategoriesJSON) != "" {
		categories, err := parseJSONArray(input.CategoriesJSON, "--categories")
		if err != nil {
			return clients.ScoreConfigCreateBody{}, err
		}
		body.Categories = categories
	}

	return body, nil
}

type ScoreConfigUpdateInput struct {
	Description    *string
	IsArchived     *bool
	MinValue       *float64
	MaxValue       *float64
	CategoriesJSON string
}

func BuildScoreConfigUpdateBody(
	input ScoreConfigUpdateInput,
) (clients.ScoreConfigUpdateBody, error) {
	body := clients.ScoreConfigUpdateBody{
		Description: input.Description,
		IsArchived:  input.IsArchived,
		MinValue:    input.MinValue,
		MaxValue:    input.MaxValue,
	}

	if strings.TrimSpace(input.CategoriesJSON) != "" {
		categories, err := parseJSONArray(input.CategoriesJSON, "--categories")
		if err != nil {
			return clients.ScoreConfigUpdateBody{}, err
		}
		body.Categories = categories
	}

	return body, nil
}
