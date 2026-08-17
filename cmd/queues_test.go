package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestQueueListSortFlags(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"queues list", queuesListCmd},
		{"queue-items list", queueItemsListCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, flagName := range []string{"sort-by", "sort-order"} {
				flag := tt.cmd.Flag(flagName)
				if flag == nil {
					t.Fatalf("flag %q not found", flagName)
				}
				if flag.DefValue != "" {
					t.Fatalf("flag %q default = %q, want empty", flagName, flag.DefValue)
				}
			}
		})
	}
}
