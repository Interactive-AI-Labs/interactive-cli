package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var commentColumnMap = map[string]struct {
	Header string
	Value  func(c *clients.CommentInfo) string
}{
	"id":          {"ID", func(c *clients.CommentInfo) string { return c.ID }},
	"object_type": {"OBJECT TYPE", func(c *clients.CommentInfo) string { return c.ObjectType }},
	"object_id":   {"OBJECT ID", func(c *clients.CommentInfo) string { return c.ObjectID }},
	"content": {
		"CONTENT",
		func(c *clients.CommentInfo) string { return truncateString(c.Content, 80) },
	},
	"author_user_id": {
		"AUTHOR USER ID",
		func(c *clients.CommentInfo) string { return c.AuthorUserID },
	},
	"created_at": {
		"CREATED AT",
		func(c *clients.CommentInfo) string { return LocalTime(c.CreatedAt) },
	},
	"updated_at": {
		"UPDATED AT",
		func(c *clients.CommentInfo) string { return LocalTime(c.UpdatedAt) },
	},
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

func PrintCommentList(
	out io.Writer,
	comments []clients.CommentInfo,
	meta clients.PageMeta,
	columns []string,
) error {
	if len(comments) == 0 {
		fmt.Fprintln(out, "No comments found.")
		return nil
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		if def, ok := commentColumnMap[col]; ok {
			headers[i] = def.Header
		}
	}

	rows := make([][]string, len(comments))
	for i, comment := range comments {
		row := make([]string, len(columns))
		for j, col := range columns {
			if def, ok := commentColumnMap[col]; ok {
				row[j] = def.Value(&comment)
			}
		}
		rows[i] = row
	}

	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	PrintPageMeta(out, meta.Page, meta.TotalPages, meta.TotalItems)
	return nil
}

func PrintCommentDetail(out io.Writer, c *clients.CommentInfo) error {
	fmt.Fprintf(out, "ID:             %s\n", c.ID)
	fmt.Fprintf(out, "Object Type:    %s\n", c.ObjectType)
	fmt.Fprintf(out, "Object ID:      %s\n", c.ObjectID)
	fmt.Fprintf(out, "Content:        %s\n", c.Content)
	if c.AuthorUserID != "" {
		fmt.Fprintf(out, "Author User ID: %s\n", c.AuthorUserID)
	}
	fmt.Fprintf(out, "Created At:     %s\n", LocalTime(c.CreatedAt))
	fmt.Fprintf(out, "Updated At:     %s\n", LocalTime(c.UpdatedAt))
	return nil
}

func PrintCommentCreateResult(out io.Writer, c *clients.CommentInfo) error {
	fmt.Fprintf(out, "Created comment %q.\n", c.ID)
	fmt.Fprintf(out, "Object Type:    %s\n", c.ObjectType)
	fmt.Fprintf(out, "Object ID:      %s\n", c.ObjectID)
	fmt.Fprintf(out, "Content:        %s\n", c.Content)
	fmt.Fprintf(out, "Created At:     %s\n", LocalTime(c.CreatedAt))
	return nil
}
