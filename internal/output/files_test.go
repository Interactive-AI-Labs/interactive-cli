package output

import (
	"bytes"
	"strings"
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
			if err := PrintFileList(&buf, tt.files, tt.meta, tt.columns, false); err != nil {
				t.Fatalf("PrintFileList() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("output mismatch\ngot:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestPrintFileVersions(t *testing.T) {
	t1 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC)
	versions := []platform.FileVersion{
		{VersionId: "v-1", Size: 10, Name: "report.pdf", CreatedAt: t1, CreatedBy: "user-1", IsCurrent: false},
		{VersionId: "v-2", Size: 20, Name: "report.pdf", CreatedAt: t2, CreatedBy: "user-1", IsCurrent: false},
		{VersionId: "v-3", Size: 30, Name: "report.pdf", CreatedAt: t3, CreatedBy: "user-1", IsCurrent: true},
	}

	var buf bytes.Buffer
	if err := PrintFileVersions(&buf, versions); err != nil {
		t.Fatalf("PrintFileVersions() error = %v", err)
	}
	got := buf.String()

	ids := map[string]bool{"v-1": false, "v-2": false, "v-3": false}
	for id := range ids {
		if !strings.Contains(got, id) {
			t.Errorf("output missing version id %q:\n%s", id, got)
		}
	}
	if n := strings.Count(got, "yes"); n != 1 {
		t.Errorf("current marker count = %d, want exactly 1 in:\n%s", n, got)
	}
}

func TestPrintFileDetail(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	meta := &platform.FileMetadata{
		Id:   "f-1",
		Name: "report.pdf",
		Current: platform.FileVersion{
			VersionId: "v-1", Size: 1024, CreatedAt: createdAt, CreatedBy: "user-1",
		},
		Versions: []platform.FileVersion{
			{VersionId: "v-1", Size: 1024, Name: "report.pdf", CreatedAt: createdAt, CreatedBy: "user-1", IsCurrent: true},
		},
	}

	var buf bytes.Buffer
	if err := PrintFileDetail(&buf, meta); err != nil {
		t.Fatalf("PrintFileDetail() error = %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "f-1") || !strings.Contains(got, "report.pdf") || !strings.Contains(got, "v-1") {
		t.Errorf("detail output missing expected fields:\n%s", got)
	}
}
