package inputs

import (
	"strings"
	"testing"
)

func TestValidateTraceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid 32-char hex", "5778886310644bbba99b55ea6a3d40ba", false},
		{"valid UUID with hyphens", "d1c7fb08-4cea-4afb-8d64-e3571bd3902d", false},
		{"valid UUID without hyphens", "d1c7fb084cea4afb8d64e3571bd3902d", false},
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"contains spaces", "5778886310644bbb a99b55ea6a3d40ba", true},
		{"contains special chars", "abc123!@#", true},
		{"too long", strings.Repeat("a", 65), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTraceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTraceID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTraceColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		wantErr bool
	}{
		{"all valid columns", AllTraceColumns, false},
		{"default columns valid", DefaultTraceColumns, false},
		{"single valid column", []string{"id"}, false},
		{"unknown column", []string{"id", "nonexistent"}, true},
		{"empty list is valid", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTraceColumns(tt.columns)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTraceColumns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
