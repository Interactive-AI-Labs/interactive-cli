package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// PrintDocumentList renders a documents list as a table.
func PrintDocumentList(out io.Writer, list *clients.DocumentList) error {
	if len(list.Documents) == 0 {
		fmt.Fprintln(out, "No documents found.")
		return nil
	}

	headers := []string{"DOCUMENT", "CHUNKS"}
	rows := make([][]string, len(list.Documents))
	for i, d := range list.Documents {
		rows[i] = []string{d.DocumentID, fmt.Sprintf("%d", d.ChunkCount)}
	}
	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	if list.HasMore && list.NextCursor != nil {
		fmt.Fprintf(out, "\nMore results — next page: --cursor %s\n", *list.NextCursor)
	}
	return nil
}

// PrintDocumentChunks renders one document's chunks: a header plus a chunk table.
func PrintDocumentChunks(out io.Writer, doc *clients.DocumentChunks) error {
	fmt.Fprintf(out, "Document:  %s\n\n", doc.DocumentID)
	if len(doc.Chunks) == 0 {
		fmt.Fprintln(out, "No chunks found.")
		return nil
	}

	headers := []string{"ID", "TEXT"}
	rows := make([][]string, len(doc.Chunks))
	for i, c := range doc.Chunks {
		rows[i] = []string{c.ID, truncateString(c.Text, 70)}
	}
	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	if doc.HasMore && doc.NextCursor != nil {
		fmt.Fprintf(out, "\nMore chunks — next page: --cursor %s\n", *doc.NextCursor)
	}
	return nil
}
