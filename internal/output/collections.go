package output

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// PrintCollectionList renders a collections list as a table.
func PrintCollectionList(out io.Writer, collections []clients.CollectionSummary) error {
	if len(collections) == 0 {
		fmt.Fprintln(out, "No collections found.")
		return nil
	}

	headers := []string{"NAME", "CREATED", "UPDATED"}
	rows := make([][]string, len(collections))
	for i, c := range collections {
		rows[i] = []string{c.Name, LocalTime(c.CreatedAt), LocalTime(c.UpdatedAt)}
	}
	return PrintTable(out, headers, rows)
}

// PrintCollectionDescribe renders a collection's config: header fields plus a
// per-slot table.
func PrintCollectionDescribe(out io.Writer, c *clients.DescribeCollectionResponse) error {
	ft := "disabled"
	if c.Config.FullText != nil && c.Config.FullText.Enabled {
		ft = fmt.Sprintf("enabled (%s)", c.Config.FullText.Language)
	}
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Name:\t%s\n", c.Name)
	fmt.Fprintf(w, "Created:\t%s\n", LocalTime(c.CreatedAt))
	fmt.Fprintf(w, "Updated:\t%s\n", LocalTime(c.UpdatedAt))
	fmt.Fprintf(w, "Full-text:\t%s\n", ft)
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintln(out)

	headers := []string{"SLOT", "TYPE", "DIMENSION", "DISTANCE", "EMBEDDING MODEL", "INDEX"}
	rows := make([][]string, 0, len(c.Config.Vectors))
	for _, name := range sortedSlotNames(c.Config.Vectors) {
		slot := c.Config.Vectors[name]
		rows = append(rows, []string{
			name,
			slot.Type,
			fmt.Sprintf("%d", slot.Dimension),
			slot.Distance,
			embeddingModel(slot),
			indexType(slot),
		})
	}
	return PrintTable(out, headers, rows)
}

// PrintCollectionStats renders a collection's operational stats.
func PrintCollectionStats(out io.Writer, s *clients.CollectionStats) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Chunks:\t%d\n", s.ChunkCount)
	fmt.Fprintf(w, "Size:\t%s\n", humanBytes(s.SizeBytes))
	if err := w.Flush(); err != nil {
		return err
	}

	if len(s.IndexValid) == 0 {
		return nil
	}
	fmt.Fprintln(out)
	headers := []string{"SLOT", "INDEX VALID"}
	rows := make([][]string, 0, len(s.IndexValid))
	names := make([]string, 0, len(s.IndexValid))
	for name := range s.IndexValid {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		rows = append(rows, []string{name, fmt.Sprintf("%t", s.IndexValid[name])})
	}
	return PrintTable(out, headers, rows)
}

func sortedSlotNames(slots map[string]clients.CollectionSlot) []string {
	names := make([]string, 0, len(slots))
	for name := range slots {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func embeddingModel(slot clients.CollectionSlot) string {
	if slot.Embedding != nil && slot.Embedding.Model != "" {
		return slot.Embedding.Model
	}
	return "-"
}

func indexType(slot clients.CollectionSlot) string {
	idx := slot.Index
	if idx == nil || idx.Type == "" {
		return "deferred"
	}
	params := []string{}
	if idx.M != 0 {
		params = append(params, fmt.Sprintf("m=%d", idx.M))
	}
	if idx.EfConstruction != 0 {
		params = append(params, fmt.Sprintf("ef_construction=%d", idx.EfConstruction))
	}
	if idx.Lists != nil {
		params = append(params, fmt.Sprintf("lists=%d", *idx.Lists))
	}
	if idx.EfSearchDefault != nil {
		params = append(params, fmt.Sprintf("ef_search_default=%d", *idx.EfSearchDefault))
	}
	if len(params) == 0 {
		return idx.Type
	}
	return fmt.Sprintf("%s (%s)", idx.Type, strings.Join(params, ", "))
}

// humanBytes renders a byte count in the largest unit that keeps it >= 1.
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	val := float64(n) / float64(div)
	if val >= 1023.95 { // %.1f would render 1024.0; step up to the next unit
		val /= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", val, "KMGTPE"[exp])
}
