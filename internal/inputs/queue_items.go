package inputs

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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

func ValidateQueueItemListOptions(opts clients.QueueItemListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildQueueItemCreateBody(
	objectID, objectType, status string,
) (clients.QueueItemCreateBody, error) {
	objectID = strings.TrimSpace(objectID)
	if objectID == "" {
		return clients.QueueItemCreateBody{}, fmt.Errorf("--object-id is required")
	}

	objectType = strings.TrimSpace(objectType)
	if objectType == "" {
		return clients.QueueItemCreateBody{}, fmt.Errorf("--object-type is required")
	}

	return clients.QueueItemCreateBody{
		ObjectID:   objectID,
		ObjectType: objectType,
		Status:     strings.TrimSpace(status),
	}, nil
}
