package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode"
)

// hostileServerName is a name an uploader can choose that steers the terminal.
const hostileServerName = "wipe\x1b[2J\x1b[Hscreen\rforged.pdf"

// assertNoControlBytes allows only the expected newlines among Unicode control characters.
func assertNoControlBytes(t *testing.T, got string, wantLines int) {
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

func filesHostileNameServer(t *testing.T) *httptest.Server {
	t.Helper()

	payload, err := json.Marshal(map[string]any{
		"success": true,
		"data": map[string]any{
			"id":   "f-1",
			"name": hostileServerName,
			"current": map[string]any{
				"versionId":   "v-2",
				"size":        1024,
				"name":        hostileServerName,
				"contentType": "application/pdf",
				"createdAt":   "2026-01-01T00:00:00Z",
				"createdBy":   "user-1",
				"isCurrent":   true,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	base := "/api/platform/v1/organizations/org-1/projects/proj-1/files"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeFilesTestSessionResponse(w, r) {
			return
		}
		switch {
		case strings.HasPrefix(r.URL.Path, base):
			_, _ = w.Write(payload)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

// TestFilesUpdateRename_SanitizesServerName catches the success line printing the
// server-supplied name verbatim.
func TestFilesUpdateRename_SanitizesServerName(t *testing.T) {
	server := filesHostileNameServer(t)
	out := setupFilesUpdateTest(t, server)

	if err := filesUpdateCmd.Flags().Set("name", "renamed.pdf"); err != nil {
		t.Fatalf("set --name: %v", err)
	}
	if err := filesUpdateCmd.RunE(filesUpdateCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files update --name: %v", err)
	}

	assertNoControlBytes(t, out.String(), 2)
}
