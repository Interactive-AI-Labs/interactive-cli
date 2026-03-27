package inputs

import (
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
	body := clients.ScoreConfigCreateBody{
		Name:        strings.TrimSpace(input.Name),
		DataType:    strings.ToUpper(strings.TrimSpace(input.DataType)),
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
