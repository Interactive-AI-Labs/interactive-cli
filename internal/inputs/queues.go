package inputs

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var DefaultQueueColumns = []string{
	"id",
	"name",
	"description",
	"created_at",
}

var AllQueueColumns = []string{
	"id",
	"name",
	"description",
	"score_config_ids",
	"created_at",
	"updated_at",
}

func ValidateQueueListOptions(opts clients.AnnotationQueueListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildQueueCreateBody(
	name, description string,
	scoreConfigIDs []string,
) clients.AnnotationQueueCreateBody {
	return clients.AnnotationQueueCreateBody{
		Name:           strings.TrimSpace(name),
		Description:    strings.TrimSpace(description),
		ScoreConfigIDs: scoreConfigIDs,
	}
}
