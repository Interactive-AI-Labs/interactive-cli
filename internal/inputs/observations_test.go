package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestValidateObservationColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		wantErr bool
	}{
		{"all valid columns", AllObservationColumns, false},
		{"default columns valid", DefaultObservationColumns, false},
		{"single valid column", []string{"id"}, false},
		{"unknown column", []string{"id", "nonexistent"}, true},
		{"empty list is valid", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColumns(tt.columns, AllObservationColumns)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateObservationColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStandaloneObservationColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		wantErr bool
	}{
		{"all valid columns", AllStandaloneObservationColumns, false},
		{"default columns valid", DefaultStandaloneObservationColumns, false},
		{"single valid column", []string{"id"}, false},
		{"unknown column", []string{"id", "nonexistent"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateColumns(tt.columns, AllStandaloneObservationColumns)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ValidateStandaloneObservationColumns() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestValidateObservationSearchOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    clients.ObservationSearchOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: clients.ObservationSearchOptions{
				FromTimestamp: "2025-01-01T00:00:00Z",
				Limit:         20,
			},
			wantErr: false,
		},
		{
			name:    "invalid from timestamp",
			opts:    clients.ObservationSearchOptions{FromTimestamp: "2025-01-01"},
			wantErr: true,
		},
		{
			name: "negative limit",
			opts: clients.ObservationSearchOptions{
				FromTimestamp: "2025-01-01T00:00:00Z",
				Limit:         -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObservationSearchOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"ValidateObservationSearchOptions() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
