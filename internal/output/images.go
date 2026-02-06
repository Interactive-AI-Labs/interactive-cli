package output

import (
	"fmt"
	"io"
	"strings"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

func PrintImageList(out io.Writer, images []clients.ImageInfo) error {
	if len(images) == 0 {
		fmt.Fprintln(out, "No images found.")
		return nil
	}

	headers := []string{"NAME", "TAGS"}
	rows := make([][]string, len(images))
	for i, img := range images {
		rows[i] = []string{
			img.Name,
			strings.Join(img.Tags, ", "),
		}
	}

	return PrintTable(out, headers, rows)
}
