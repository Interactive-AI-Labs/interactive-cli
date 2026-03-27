package inputs

import (
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var DefaultCommentColumns = []string{
	"id",
	"object_type",
	"object_id",
	"content",
	"author_user_id",
	"created_at",
}

var AllCommentColumns = []string{
	"id",
	"object_type",
	"object_id",
	"content",
	"author_user_id",
	"created_at",
	"updated_at",
}

func ValidateCommentListOptions(opts clients.CommentListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildCommentCreateBody(
	objectType, objectID, content, authorUserID string,
) clients.CommentCreateBody {
	return clients.CommentCreateBody{
		ObjectType:   objectType,
		ObjectID:     objectID,
		Content:      content,
		AuthorUserID: authorUserID,
	}
}
