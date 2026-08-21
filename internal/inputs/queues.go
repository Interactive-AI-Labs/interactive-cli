package inputs

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
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
	"count_completed_items",
	"count_pending_items",
}

func ValidateQueueListOptions(opts api.AnnotationQueueListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildQueueCreateBody(
	name, description string,
	scoreConfigIDs []string,
) api.AnnotationQueueCreateBody {
	return api.AnnotationQueueCreateBody{
		Name:           strings.TrimSpace(name),
		Description:    strings.TrimSpace(description),
		ScoreConfigIDs: scoreConfigIDs,
	}
}
