package inputs

import "testing"

func TestValidatePromptDelete(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		version   int
		label     string
		byVersion bool
		byLabel   bool
		wantErr   string
	}{
		{name: "whole prompt", prompt: "x"},
		{name: "whole prompt with reserved segment", prompt: "folders/x"},
		{name: "version", prompt: "x", version: 3, byVersion: true},
		{name: "label", prompt: "x", label: "prod", byLabel: true},
		{name: "segment-like name", prompt: "folders-x", version: 1, byVersion: true},
		{
			name:      "version 0",
			prompt:    "x",
			byVersion: true,
			wantErr:   "--version must be 1 or greater",
		},
		{
			name:      "negative version",
			prompt:    "x",
			version:   -1,
			byVersion: true,
			wantErr:   "--version must be 1 or greater",
		},
		{name: "empty label", prompt: "x", byLabel: true, wantErr: "--label must not be empty"},
		{
			name:    "blank label",
			prompt:  "x",
			label:   " ",
			byLabel: true,
			wantErr: "--label must not be empty",
		},
		{
			name:      "folders segment",
			prompt:    "folders/x",
			version:   2,
			byVersion: true,
			wantErr:   `cannot delete "folders/x" by --version or --label: its "folders" segment routes to a wider delete`,
		},
		{
			name:    "by-name segment",
			prompt:  "a/by-name",
			label:   "prod",
			byLabel: true,
			wantErr: `cannot delete "a/by-name" by --version or --label: its "by-name" segment routes to a wider delete`,
		},
		{
			name:      "versions segment",
			prompt:    "a/versions/b",
			version:   1,
			byVersion: true,
			wantErr:   `cannot delete "a/versions/b" by --version or --label: its "versions" segment routes to a wider delete`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := ""
			if err := ValidatePromptDelete(
				tt.prompt,
				tt.version,
				tt.label,
				tt.byVersion,
				tt.byLabel,
			); err != nil {
				gotErr = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("error = %q, want %q", gotErr, tt.wantErr)
			}
		})
	}
}
