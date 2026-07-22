package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestImageDeleteValidation(t *testing.T) {
	originalTag := imageDeleteTag
	t.Cleanup(func() { imageDeleteTag = originalTag })

	tests := []struct {
		name string
		args []string
		tag  string
		want string
	}{
		{
			name: "empty image name",
			args: []string{" "},
			tag:  "v1",
			want: "image name is required",
		},
		{
			name: "empty tag",
			args: []string{"app"},
			tag:  " ",
			want: "tag is required; please provide --tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imageDeleteTag = tt.tag
			err := imageDeleteCmd.RunE(imageDeleteCmd, tt.args)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if diff := cmp.Diff(tt.want, err.Error()); diff != "" {
				t.Errorf("error mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
