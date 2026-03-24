package inputs

import (
	"strings"
	"testing"
)

func TestValidateObservationID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid hex ID", "8973e1e3d5b29474", false},
		{"valid UUID", "d1c7fb08-4cea-4afb-8d64-e3571bd3902d", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"too long", strings.Repeat("a", 257), true},
		{"max length is valid", strings.Repeat("a", 256), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObservationID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateObservationID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

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
			err := ValidateObservationColumns(tt.columns)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateObservationColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
