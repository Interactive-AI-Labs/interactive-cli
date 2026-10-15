package cmd

import "testing"

func TestPromptDeleteTarget(t *testing.T) {
	refusal := func(name, segment string) string {
		return `cannot delete by --version or --label: "` + name + `" contains the path segment "` +
			segment + `", which the server may treat as a wider delete`
	}

	tests := []struct {
		name      string
		prompt    string
		version   int
		label     string
		byVersion bool
		byLabel   bool
		want      string
		wantErr   string
	}{
		{name: "whole prompt", prompt: "x", want: `routine "x" and all its versions`},
		{
			name:   "whole prompt ignores reserved segments",
			prompt: "folders/x",
			want:   `routine "folders/x" and all its versions`,
		},
		{
			name:      "version",
			prompt:    "x",
			version:   3,
			byVersion: true,
			want:      `version 3 of routine "x"`,
		},
		{
			name:    "label",
			prompt:  "x",
			label:   "prod",
			byLabel: true,
			want:    `versions labelled "prod" of routine "x"`,
		},
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
			name:      "folders segment",
			prompt:    "folders/x",
			version:   2,
			byVersion: true,
			wantErr:   refusal("folders/x", "folders"),
		},
		{
			name:    "by-name segment",
			prompt:  "a/by-name",
			label:   "prod",
			byLabel: true,
			wantErr: refusal("a/by-name", "by-name"),
		},
		{
			name:      "versions segment",
			prompt:    "a/versions/b",
			version:   1,
			byVersion: true,
			wantErr:   refusal("a/versions/b", "versions"),
		},
		{
			name:      "segment-like name",
			prompt:    "folders-x",
			version:   1,
			byVersion: true,
			want:      `version 1 of routine "folders-x"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := promptDeleteTarget(
				"routine",
				tt.prompt,
				tt.version,
				tt.label,
				tt.byVersion,
				tt.byLabel,
			)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if got != tt.want || gotErr != tt.wantErr {
				t.Errorf("got (%q, %q), want (%q, %q)", got, gotErr, tt.want, tt.wantErr)
			}
		})
	}
}
