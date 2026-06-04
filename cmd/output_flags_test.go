package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateTableOnlyColumns(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		asJSON  bool
		asYAML  bool
		wantErr bool
	}{
		{
			name: "columns with table output is allowed",
			args: []string{"--columns", "id,name"},
		},
		{
			name:    "columns with json is rejected",
			args:    []string{"--columns", "id,name"},
			asJSON:  true,
			wantErr: true,
		},
		{
			name:    "columns with yaml is rejected",
			args:    []string{"--columns", "id,name"},
			asYAML:  true,
			wantErr: true,
		},
		{
			name:   "json without columns is allowed",
			asJSON: true,
		},
		{
			name:   "yaml without columns is allowed",
			asYAML: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			var columns []string
			cmd.Flags().StringSliceVar(&columns, "columns", nil, "Columns to display")
			cmd.SetArgs(tt.args)

			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			err := validateTableOnlyColumns(cmd, tt.asJSON, tt.asYAML)
			if tt.wantErr && err == nil {
				t.Fatal("validateTableOnlyColumns() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validateTableOnlyColumns() error = %v", err)
			}
		})
	}
}
