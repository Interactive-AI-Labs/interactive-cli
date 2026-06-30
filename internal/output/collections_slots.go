package output

import (
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// PrintSlotAddResult renders the result of adding a slot.
func PrintSlotAddResult(out io.Writer, r *clients.SlotAddResult) error {
	fmt.Fprintf(out, "Slot:          %s\n", r.Slot)
	fmt.Fprintf(out, "Type:          %s\n", r.Type)
	fmt.Fprintf(out, "Dimension:     %d\n", r.Dimension)
	fmt.Fprintf(out, "Distance:      %s\n", r.Distance)
	fmt.Fprintf(out, "Index status:  %s\n", r.IndexStatus)
	return nil
}

// PrintSlotIndexProgress renders a slot's index build progress.
func PrintSlotIndexProgress(out io.Writer, p *clients.SlotIndexProgress) error {
	fmt.Fprintf(out, "Slot:        %s\n", p.Slot)
	fmt.Fprintf(out, "Index type:  %s\n", p.IndexType)
	fmt.Fprintf(out, "Status:      %s\n", p.Status)
	return nil
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
