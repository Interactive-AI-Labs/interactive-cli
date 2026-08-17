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

var QueueSortFields = []string{
	"name",
	"description",
	"created_at",
	"updated_at",
	"count_completed_items",
	"count_pending_items",
}

func ValidateQueueListOptions(opts clients.AnnotationQueueListOptions) error {
	if err := ValidatePagination(opts.Page, opts.Limit); err != nil {
		return err
	}
	return ValidateSorting(opts.SortBy, opts.SortOrder, QueueSortFields)
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
