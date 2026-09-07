package output

import (
	"bytes"
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

// hostileFileNames are names an uploader can choose that steer the terminal:
// a forged row, a forged row prefix, a screen wipe and a forged column.
var hostileFileNames = []string{
	"forge\ninjected-row.pdf",
	"forge\rf-666   spoofed.pdf",
	"wipe\x1b[2J\x1b[Hscreen.pdf",
	"split\tcolumn.pdf",
}

// assertRenderedLines checks the rendering emits exactly wantLines lines and
// that the only control character present is the line terminator itself —
// including multi-byte Unicode controls (e.g. NEL U+0085, CSI U+009B), not
// just single-byte ASCII ones.
func assertRenderedLines(t *testing.T, got string, wantLines int) {
	t.Helper()

	if n := strings.Count(got, "\n"); n != wantLines {
		t.Errorf("line count = %d, want %d; a name forged a line boundary:\n%q", n, wantLines, got)
	}
	for i, r := range got {
		if r == '\n' {
			continue
		}
		if unicode.IsControl(r) {
			t.Errorf(
				"rune at byte %d = %q reaches the terminal as a control character:\n%q",
				i,
				r,
				got,
			)
		}
	}
}

func hostileFiles(t *testing.T) []platform.FileMetadata {
	t.Helper()

	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	files := make([]platform.FileMetadata, len(hostileFileNames))
	for i, name := range hostileFileNames {
		email := "user\x1b[31m@example.com"
		files[i] = platform.FileMetadata{
			ID:   "f-" + string(rune('1'+i)),
			Name: name,
			Current: platform.FileVersion{
				VersionID:      "v-1",
				Size:           1024,
				Name:           name,
				ContentType:    "application/\x1b[7mpdf",
				CreatedAt:      createdAt,
				CreatedBy:      "user-1",
				CreatedByEmail: &email,
				IsCurrent:      true,
			},
		}
	}
	return files
}

func hostileVersions(t *testing.T) []platform.FileVersion {
	t.Helper()

	versions := make([]platform.FileVersion, 0, len(hostileFileNames))
	for _, f := range hostileFiles(t) {
		versions = append(versions, f.Current)
	}
	return versions
}

func TestFileRenderers_SanitizeServerValues(t *testing.T) {
	files := hostileFiles(t)
	versions := hostileVersions(t)
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	candidates := make([]clients.FileRefCandidate, len(hostileFileNames))
	for i, name := range hostileFileNames {
		candidates[i] = clients.FileRefCandidate{
			FileID: "f-" + string(rune('1'+i)), Name: name, Size: 1024, CreatedAt: createdAt,
		}
	}

	tests := []struct {
		name      string
		wantLines int
		render    func(*bytes.Buffer) error
	}{
		{
			name:      "file list",
			wantLines: 1 + len(files),
			render: func(buf *bytes.Buffer) error {
				return PrintFileList(
					buf,
					files,
					platform.FileListMeta{},
					inputs.AllFileColumns,
					false,
				)
			},
		},
		{
			name:      "file list cursor and versions header",
			wantLines: 2 + 2 + 2 + 2,
			render: func(buf *bytes.Buffer) error {
				file := files[0]
				file.Versions = []platform.FileVersion{file.Current}
				cursor := "cur\x1b[2Jsor\n"
				return PrintFileList(
					buf,
					[]platform.FileMetadata{file},
					platform.FileListMeta{Cursor: &cursor},
					inputs.DefaultFileColumns,
					true,
				)
			},
		},
		{
			name:      "file versions",
			wantLines: 1 + len(versions),
			render:    func(buf *bytes.Buffer) error { return PrintFileVersions(buf, versions) },
		},
		{
			name:      "file detail",
			wantLines: 6 + 2 + 2,
			render: func(buf *bytes.Buffer) error {
				file := files[0]
				file.Versions = []platform.FileVersion{file.Current}
				return PrintFileDetail(buf, &file)
			},
		},
		{
			name:      "file reference candidates",
			wantLines: 1 + 1 + len(candidates),
			render: func(buf *bytes.Buffer) error {
				return PrintFileRefCandidates(buf, "report.pdf", candidates)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tt.render(&buf); err != nil {
				t.Fatalf("render() error = %v", err)
			}
			assertRenderedLines(t, buf.String(), tt.wantLines)
		})
	}
}

// TestPrintFiles_LeavesLegitimateUnicodeIntact guards against a sanitizer broad
// enough to mangle real non-ASCII filenames.
func TestPrintFiles_LeavesLegitimateUnicodeIntact(t *testing.T) {
	names := []string{"отчёт-за-третий-квартал.pdf", "報告書-2026年.pdf"}
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	files := make([]platform.FileMetadata, len(names))
	for i, name := range names {
		files[i] = platform.FileMetadata{
			ID:   "f-" + string(rune('1'+i)),
			Name: name,
			Current: platform.FileVersion{
				VersionID: "v-1", Size: 1024, Name: name,
				ContentType: "application/pdf", CreatedAt: createdAt, CreatedBy: "user-1",
				IsCurrent: true,
			},
			Versions: nil,
		}
	}

	var list bytes.Buffer
	if err := PrintFileList(
		&list,
		files,
		platform.FileListMeta{},
		inputs.AllFileColumns,
		false,
	); err != nil {
		t.Fatalf("PrintFileList() error = %v", err)
	}

	for _, name := range names {
		if !strings.Contains(list.String(), name) {
			t.Errorf("PrintFileList output lost the name %q:\n%s", name, list.String())
		}
	}
	if strings.Contains(list.String(), "\uFFFD") {
		t.Errorf("PrintFileList replaced legitimate runes:\n%s", list.String())
	}

	for i := range files {
		files[i].Versions = []platform.FileVersion{files[i].Current}

		var detail bytes.Buffer
		if err := PrintFileDetail(&detail, &files[i]); err != nil {
			t.Fatalf("PrintFileDetail() error = %v", err)
		}
		if !strings.Contains(detail.String(), names[i]) {
			t.Errorf("PrintFileDetail output lost the name %q:\n%s", names[i], detail.String())
		}
		if strings.Contains(detail.String(), "\uFFFD") {
			t.Errorf("PrintFileDetail replaced legitimate runes:\n%s", detail.String())
		}
	}
}
