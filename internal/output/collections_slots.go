package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// PrintSlotAddResult renders the result of adding a slot.
func PrintSlotAddResult(out io.Writer, r *clients.SlotAddResult) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Slot:\t%s\n", r.Slot)
	fmt.Fprintf(w, "Type:\t%s\n", r.Type)
	fmt.Fprintf(w, "Dimension:\t%d\n", r.Dimension)
	fmt.Fprintf(w, "Distance:\t%s\n", r.Distance)
	fmt.Fprintf(w, "Index status:\t%s\n", r.IndexStatus)
	return w.Flush()
}

// PrintSlotIndexProgress renders a slot's index build progress.
func PrintSlotIndexProgress(out io.Writer, p *clients.SlotIndexProgress) error {
	w := NewDescribeWriter(out)
	fmt.Fprintf(w, "Slot:\t%s\n", p.Slot)
	fmt.Fprintf(w, "Index type:\t%s\n", p.IndexType)
	fmt.Fprintf(w, "Status:\t%s\n", p.Status)
	return w.Flush()
}

// PrintSlotOpResult renders a reindex/vacuum result (whichever status is set).
func PrintSlotOpResult(out io.Writer, r *clients.SlotOpResult) error {
	status := r.Status
	if status == "" {
		status = r.IndexStatus
	}
	fmt.Fprintf(out, "Slot %q: %s\n", r.Slot, status)
	return nil
}
