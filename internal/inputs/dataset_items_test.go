package inputs

import (
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/google/go-cmp/cmp"
)

func TestBuildDatasetItemUpdateBody(t *testing.T) {
	tests := []struct {
		name string
		in   DatasetItemUpdateInput
		want platform.DatasetItemUpdateBody
		err  string
	}{
		{
			name: "name is trimmed",
			in:   DatasetItemUpdateInput{Name: utils.ToPtr(" scenario-1 ")},
			want: platform.DatasetItemUpdateBody{"name": "scenario-1"},
		},
		{
			name: "empty name clears it",
			in:   DatasetItemUpdateInput{Name: utils.ToPtr("")},
			want: platform.DatasetItemUpdateBody{"name": nil},
		},
		{
			name: "status is upper-cased",
			in:   DatasetItemUpdateInput{Status: "archived"},
			want: platform.DatasetItemUpdateBody{"status": "ARCHIVED"},
		},
		{
			name: "metadata object",
			in:   DatasetItemUpdateInput{MetadataJSON: `{"k":1}`},
			want: platform.DatasetItemUpdateBody{"metadata": map[string]any{"k": float64(1)}},
		},
		{
			name: "all three",
			in: DatasetItemUpdateInput{
				Name:         utils.ToPtr("a"),
				Status:       "ACTIVE",
				MetadataJSON: `{}`,
			},
			want: platform.DatasetItemUpdateBody{
				"name":     "a",
				"status":   "ACTIVE",
				"metadata": map[string]any{},
			},
		},
		{
			name: "metadata must be an object",
			in:   DatasetItemUpdateInput{MetadataJSON: `[1]`},
			err:  "invalid --metadata-json: must be a valid JSON object",
		},
		{
			name: "nothing to update",
			in:   DatasetItemUpdateInput{},
			err:  "nothing to update: pass --name, --status or --metadata-json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildDatasetItemUpdateBody(tt.in)
			gotErr := ""
			if err != nil {
				gotErr = err.Error()
			}
			if diff := cmp.Diff(tt.err, gotErr); diff != "" {
				t.Fatalf("error mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildDatasetItemCreateBodyName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "trimmed", in: " scenario-1 ", want: "scenario-1"},
		{name: "omitted", in: "", want: ""},
		{name: "blank", in: "   ", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildDatasetItemCreateBody(
				DatasetItemCreateInput{DatasetName: "ds", Name: tt.in},
			)
			if err != nil {
				t.Fatal(err)
			}
			want := platform.DatasetItemCreateBody{DatasetName: "ds", Name: tt.want}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("body mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
