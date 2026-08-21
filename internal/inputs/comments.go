package inputs

import "github.com/Interactive-AI-Labs/interactive-cli/internal/clients/api"

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

func ValidateCommentListOptions(opts api.CommentListOptions) error {
	return ValidatePagination(opts.Page, opts.Limit)
}

func BuildCommentCreateBody(
	objectType, objectID, content, authorUserID string,
) api.CommentCreateBody {
	return api.CommentCreateBody{
		ObjectType:   objectType,
		ObjectID:     objectID,
		Content:      content,
		AuthorUserID: authorUserID,
	}
}
