package cmd

import (
	"slices"
	"testing"
)

func TestStackGetOutputFlags(t *testing.T) {
	tests := []struct {
		name      string
		flags     []string
		wantError string
	}{
		{"default stdout", nil, ""},
		{"file", []string{"file"}, ""},
		{"json", []string{"json"}, ""},
		{"yaml", []string{"yaml"}, ""},
		{
			"file and json",
			[]string{"file", "json"},
			"if any flags in the group [file json yaml] are set none of the others can be; [file json] were all set",
		},
		{
			"file and yaml",
			[]string{"file", "yaml"},
			"if any flags in the group [file json yaml] are set none of the others can be; [file yaml] were all set",
		},
		{
			"json and yaml",
			[]string{"json", "yaml"},
			"if any flags in the group [file json yaml] are set none of the others can be; [json yaml] were all set",
		},
		{
			"all output modes",
			[]string{"file", "json", "yaml"},
			"if any flags in the group [file json yaml] are set none of the others can be; [file json yaml] were all set",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, name := range []string{"file", "json", "yaml"} {
				flag := stackGetCmd.Flags().Lookup(name)
				previous := flag.Changed
				flag.Changed = slices.Contains(tt.flags, name)
				t.Cleanup(func() { flag.Changed = previous })
			}
			err := stackGetCmd.ValidateFlagGroups()
			message := ""
			if err != nil {
				message = err.Error()
			}
			if message != tt.wantError {
				t.Fatalf("error = %q, want %q", message, tt.wantError)
			}
		})
	}
}
