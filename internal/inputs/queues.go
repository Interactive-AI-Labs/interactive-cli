package inputs

import (
	"fmt"
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
) (clients.AnnotationQueueCreateBody, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return clients.AnnotationQueueCreateBody{}, fmt.Errorf("queue name is required")
	}

	return clients.AnnotationQueueCreateBody{
		Name:           name,
		Description:    strings.TrimSpace(description),
		ScoreConfigIDs: scoreConfigIDs,
	}, nil
}
