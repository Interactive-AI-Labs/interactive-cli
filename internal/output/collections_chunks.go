package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// PrintChunkUpsertResult summarizes an upsert response: a per-status count plus
// any per-chunk errors.
func PrintChunkUpsertResult(out io.Writer, r *clients.ChunkUpsertResult) error {
	byStatus := map[string]int{}
	for _, res := range r.Results {
		byStatus[res.Status]++
	}
	statuses := make([]string, 0, len(byStatus))
	for status := range byStatus {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	for _, status := range statuses {
		fmt.Fprintf(out, "%d %s\n", byStatus[status], status)
	}
	if len(r.Errors) > 0 {
		fmt.Fprintf(out, "\n%d error(s):\n", len(r.Errors))
		for _, e := range r.Errors {
			b, _ := json.Marshal(e)
			fmt.Fprintf(out, "  %s\n", string(b))
		}
	}
	return nil
}

// PrintChunkList renders a page of chunks as a table.
func PrintChunkList(out io.Writer, list *clients.ChunkList) error {
	if len(list.Chunks) == 0 {
		fmt.Fprintln(out, "No chunks found.")
		return nil
	}

	headers := []string{"ID", "DOCUMENT", "TEXT"}
	rows := make([][]string, len(list.Chunks))
	for i, c := range list.Chunks {
		rows[i] = []string{c.ID, c.DocumentID, truncate(c.Text, 60)}
	}
	if err := PrintTable(out, headers, rows); err != nil {
		return err
	}

	if list.HasMore && list.NextCursor != nil {
		fmt.Fprintf(out, "\nMore results — next page: --cursor %s\n", *list.NextCursor)
	}
	return nil
}

// PrintChunk renders a single chunk's detail.
func PrintChunk(out io.Writer, c *clients.Chunk) error {
	fmt.Fprintf(out, "ID:        %s\n", c.ID)
	fmt.Fprintf(out, "Document:  %s\n", c.DocumentID)
	fmt.Fprintf(out, "Text:      %s\n", c.Text)

	if len(c.Metadata) > 0 {
		b, _ := json.MarshalIndent(c.Metadata, "           ", "  ")
		fmt.Fprintf(out, "Metadata:  %s\n", string(b))
	}

	slots := make([]string, 0, len(c.Vectors))
	for slot := range c.Vectors {
		slots = append(slots, slot)
	}
	sort.Strings(slots)
	for _, slot := range slots {
		fmt.Fprintf(out, "Vector[%s]: %s\n", slot, vectorSummary(c.Vectors[slot]))
	}
	if len(c.Vector) > 0 {
		fmt.Fprintf(out, "Vector:    %d dims\n", len(c.Vector))
	}
	return nil
}

// vectorSummary describes a stored vector without dumping its values: dense
// dimension count, sparse fill, or binary bit length.
func vectorSummary(raw json.RawMessage) string {
	var dense []float64
	if err := json.Unmarshal(raw, &dense); err == nil {
		return fmt.Sprintf("%d dims", len(dense))
	}
	var sparse struct {
		Indices []int `json:"indices"`
		Dim     int   `json:"dim"`
	}
	if err := json.Unmarshal(raw, &sparse); err == nil && len(sparse.Indices) > 0 {
		return fmt.Sprintf("sparse, %d of %d dims set", len(sparse.Indices), sparse.Dim)
	}
	var bits string
	if err := json.Unmarshal(raw, &bits); err == nil {
		return fmt.Sprintf("binary, %d bits", len(bits))
	}
	return "present (unrecognized format)"
}

// PrintBulkDeleteResult renders a bulk-delete response. The deleted id list is
// only useful for --filter/--all deletes (the caller already knows the ids for
// --ids), so it's printed one per line when present.
func PrintBulkDeleteResult(out io.Writer, r *clients.BulkDeleteResult) error {
	fmt.Fprintf(out, "Deleted %d chunk(s)\n", r.DeletedCount)
	if len(r.DeletedIds) > 0 {
		for _, id := range r.DeletedIds {
			fmt.Fprintf(out, "  %s\n", id)
		}
	}
	return nil
}

// truncate shortens s to n runes, appending an ellipsis when cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
