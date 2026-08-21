package inputs

import "github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"

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

func ValidateCommentListOptions(opts platform.CommentListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildCommentCreateBody(
	objectType, objectID, content, authorUserID string,
) platform.CommentCreateBody {
	return platform.CommentCreateBody{
		ObjectType:   objectType,
		ObjectID:     objectID,
		Content:      content,
		AuthorUserID: authorUserID,
	}
}
