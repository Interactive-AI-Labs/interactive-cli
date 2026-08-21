package inputs

import (
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"
)

var DefaultQueueItemColumns = []string{
	"id",
	"object_id",
	"object_type",
	"status",
	"created_at",
}

var AllQueueItemColumns = []string{
	"id",
	"object_id",
	"object_type",
	"status",
	"completed_at",
	"created_at",
	"updated_at",
}

func ValidateQueueItemListOptions(opts api.QueueItemListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildQueueItemCreateBody(
	objectID, objectType, status string,
) api.QueueItemCreateBody {
	return api.QueueItemCreateBody{
		ObjectID:   strings.TrimSpace(objectID),
		ObjectType: strings.TrimSpace(objectType),
		Status:     strings.TrimSpace(status),
	}
}
