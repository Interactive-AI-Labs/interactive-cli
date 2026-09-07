package cmd

import "testing"

func TestSafeDownloadFilename(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		fileID string
		want   string
	}{
		{"a normal name", "report.pdf", "file-1", "report.pdf"},
		{"a traversal path reduces to its base", "../../evil.txt", "file-1", "evil.txt"},
		{"a bare dot-dot falls back to the id", "..", "file-1", "file-1"},
		{"a bare dot falls back to the id", ".", "file-1", "file-1"},
		{"empty falls back to the id", "", "file-1", "file-1"},
		{"a bare separator falls back to the id", "/", "file-1", "file-1"},
		{"an absolute path falls back to the id", "/etc/passwd", "file-1", "passwd"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safeDownloadFilename(tt.raw, tt.fileID); got != tt.want {
				t.Errorf("safeDownloadFilename(%q, %q) = %q, want %q", tt.raw, tt.fileID, got, tt.want)
			}
		})
	}
}
