package output

import (
	"bytes"
	"testing"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func TestPrintImageList(t *testing.T) {
	tests := []struct {
		name   string
		images []clients.ImageInfo
		want   string
	}{
		{
			name:   "empty list prints message",
			images: []clients.ImageInfo{},
			want:   "No images found.\n",
		},
		{
			name:   "nil list prints message",
			images: nil,
			want:   "No images found.\n",
		},
		{
			name: "single image with one tag",
			images: []clients.ImageInfo{
				{Name: "myapp", Tags: []string{"latest"}},
			},
			want: "NAME    TAGS\n" +
				"myapp   latest\n",
		},
		{
			name: "image with multiple tags joined",
			images: []clients.ImageInfo{
				{Name: "api", Tags: []string{"v1.0", "v1.1", "latest"}},
			},
			want: "NAME   TAGS\n" +
				"api    v1.0, v1.1, latest\n",
		},
		{
			name: "image with no tags",
			images: []clients.ImageInfo{
				{Name: "untagged", Tags: []string{}},
			},
			want: "NAME       TAGS\n" +
				"untagged   \n",
		},
		{
			name: "multiple images",
			images: []clients.ImageInfo{
				{Name: "frontend", Tags: []string{"v2"}},
				{Name: "backend", Tags: []string{"v3", "stable"}},
			},
			want: "NAME       TAGS\n" +
				"frontend   v2\n" +
				"backend    v3, stable\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := PrintImageList(&buf, tt.images)
			if err != nil {
				t.Fatalf("PrintImageList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
