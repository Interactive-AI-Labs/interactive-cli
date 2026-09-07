package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

func TestPrintFileRefCandidates(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	candidates := []clients.FileRefCandidate{
		{FileId: "f-1", Name: "report.pdf", Size: 1024, CreatedAt: createdAt},
		{FileId: "f-2", Name: "report.pdf", Size: 10, CreatedAt: createdAt},
	}

	var buf bytes.Buffer
	if err := PrintFileRefCandidates(&buf, "report.pdf", candidates); err != nil {
		t.Fatalf("PrintFileRefCandidates() error = %v", err)
	}

	want := "\"report.pdf\" matches more than one file:\n" +
		"ID    NAME         SIZE      CREATED AT\n" +
		"f-1   report.pdf   1.0 KiB   " + LocalTime(createdAt.Format(time.RFC3339)) + "\n" +
		"f-2   report.pdf   10 B      " + LocalTime(createdAt.Format(time.RFC3339)) + "\n"
	if got := buf.String(); got != want {
		t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, want)
	}
}

func TestPrintFileList(t *testing.T) {
	email := "a@example.com"
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	cursor := "next-page"

	tests := []struct {
		name    string
		files   []platform.FileMetadata
		meta    platform.FileListMeta
		columns []string
		want    string
	}{
		{
			name:    "empty list",
			columns: inputs.DefaultFileColumns,
			want:    "No files found.\n",
		},
		{
			name: "falls back to raw id without an email, prints cursor footer",
			files: []platform.FileMetadata{
				{
					Id:   "file-1",
					Name: "report.pdf",
					Current: platform.FileVersion{
						Size:           1024,
						CreatedAt:      createdAt,
						CreatedBy:      "user-1",
						CreatedByEmail: &email,
					},
				},
				{
					Id:   "file-2",
					Name: "notes.txt",
					Current: platform.FileVersion{
						Size:      10,
						CreatedAt: createdAt,
						CreatedBy: "pk-abc",
					},
				},
			},
			meta:    platform.FileListMeta{Cursor: &cursor},
			columns: inputs.DefaultFileColumns,
			want: "ID       NAME         SIZE      UPLOADED BY     CREATED AT\n" +
				"file-1   report.pdf   1.0 KiB   a@example.com   " +
				LocalTime(createdAt.Format(time.RFC3339)) + "\n" +
				"file-2   notes.txt    10 B      pk-abc          " +
				LocalTime(createdAt.Format(time.RFC3339)) + "\n" +
				"\nMore results — next page: --cursor next-page\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := PrintFileList(&buf, tt.files, tt.meta, tt.columns); err != nil {
				t.Fatalf("PrintFileList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
