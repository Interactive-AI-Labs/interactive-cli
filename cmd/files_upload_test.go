package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

const filesUploadResponseBody = `{"success":true,"data":{"id":"f-1","name":"report.txt",` +
	`"current":{"versionId":"v-1","size":12,"name":"report.txt","contentType":"text/plain",` +
	`"createdAt":"2024-01-01T00:00:00Z","createdBy":"u-1","createdByEmail":null,"isCurrent":true}}}`

// Guarded because an aborted upload can leave the handler running past RunE.
type filesRequestLog struct {
	mu   sync.Mutex
	seen []string
}

func (l *filesRequestLog) add(entry string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seen = append(l.seen, entry)
}

func (l *filesRequestLog) all() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.seen...)
}

func filesUploadTestServer(t *testing.T, requests *filesRequestLog) *httptest.Server {
	t.Helper()
	base := "/api/platform/v1/organizations/org-1/projects/proj-1/files"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeFilesTestSessionResponse(w, r) {
			return
		}
		switch {
		case strings.HasPrefix(r.URL.Path, base):
			requests.add(r.Method + " " + r.URL.Path)
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			}
			fmt.Fprint(w, filesUploadResponseBody)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func setupFilesUploadTest(t *testing.T, server *httptest.Server) {
	t.Helper()
	setupFilesCommandTest(t, server, filesUploadCmd)
	origTimeout := filesUploadTimeout
	t.Cleanup(func() {
		filesUploadOrg, filesUploadProject = "", ""
		filesUploadName = ""
		filesUploadTimeout = origTimeout
		for _, name := range []string{"name", "timeout", "organization", "project"} {
			filesUploadCmd.Flags().Lookup(name).Changed = false
		}
	})
	filesUploadOrg, filesUploadProject = "acme", "alunafi"
}

func setupFilesUpdateTest(t *testing.T, server *httptest.Server) *strings.Builder {
	t.Helper()
	out := setupFilesCommandTest(t, server, filesUpdateCmd)
	origTimeout := filesUpdateTimeout
	t.Cleanup(func() {
		filesUpdateOrg, filesUpdateProject = "", ""
		filesUpdateName = ""
		filesUpdateTimeout = origTimeout
		for _, name := range []string{"name", "timeout", "organization", "project"} {
			filesUpdateCmd.Flags().Lookup(name).Changed = false
		}
	})
	filesUpdateOrg, filesUpdateProject = "acme", "alunafi"
	return out
}

func TestFilesUpload_RegularFileSendsOneRequest(t *testing.T) {
	requests := new(filesRequestLog)
	server := filesUploadTestServer(t, requests)
	setupFilesUploadTest(t, server)

	localPath := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(localPath, []byte("report bytes"), 0o644); err != nil {
		t.Fatalf("write local file: %v", err)
	}

	if err := filesUploadCmd.RunE(filesUploadCmd, []string{localPath}); err != nil {
		t.Fatalf("files upload %s: %v", localPath, err)
	}

	got := requests.all()
	if len(got) != 1 {
		t.Fatalf("requests = %v, want exactly 1", got)
	}
	if got[0] != "POST /api/platform/v1/organizations/org-1/projects/proj-1/files" {
		t.Errorf("request = %q, want a files POST", got[0])
	}
}

func TestPrintFileResult_UsesLogicalFileName(t *testing.T) {
	var out strings.Builder
	printFileResult(&out, "Restored", &platform.FileMetadata{
		ID:   "f-1",
		Name: "current-name.txt",
		Current: platform.FileVersion{
			VersionID: "v-2",
			Name:      "historical-name.txt",
			Size:      12,
		},
	})

	want := "Restored current-name.txt\n" +
		"  id:      f-1\n" +
		"  version: v-2\n" +
		"  size:    12 B\n"
	if got := out.String(); got != want {
		t.Errorf("printFileResult output = %q, want %q", got, want)
	}
}

func TestFilesUploadAndUpdate_DirectoryOperandIsRefusedBeforeTransfer(t *testing.T) {
	tests := []struct {
		name   string
		update bool
	}{
		{name: "upload"},
		{name: "update", update: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := new(filesRequestLog)
			server := filesUploadTestServer(t, requests)
			dir := t.TempDir()

			var err error
			if tt.update {
				setupFilesUpdateTest(t, server)
				err = filesUpdateCmd.RunE(filesUpdateCmd, []string{"f-1", dir})
			} else {
				setupFilesUploadTest(t, server)
				err = filesUploadCmd.RunE(filesUploadCmd, []string{dir})
			}

			if err == nil {
				t.Fatalf("files %s %s: want an error, got nil", tt.name, dir)
			}
			if !strings.Contains(err.Error(), dir) {
				t.Errorf("error = %q, want it to name %q", err, dir)
			}
			if !strings.Contains(err.Error(), "not a regular file") {
				t.Errorf("error = %q, want it to reject the operand as not a regular file", err)
			}
			if got := requests.all(); len(got) != 0 {
				t.Errorf("requests = %v, want none; a partial upload reached the server", got)
			}
		})
	}
}

func TestFilesUpdate_NameOnlySendsRenameRequest(t *testing.T) {
	requests := new(filesRequestLog)
	server := filesUploadTestServer(t, requests)
	setupFilesUpdateTest(t, server)

	if err := filesUpdateCmd.Flags().Set("name", "renamed.pdf"); err != nil {
		t.Fatalf("set --name: %v", err)
	}
	args := []string{"f-1"}
	if err := filesUpdateCmd.Args(filesUpdateCmd, args); err != nil {
		t.Fatalf("validate files update --name: %v", err)
	}
	if err := filesUpdateCmd.RunE(filesUpdateCmd, args); err != nil {
		t.Fatalf("files update f-1 --name renamed.pdf: %v", err)
	}

	got := requests.all()
	if len(got) != 1 {
		t.Fatalf("requests = %v, want exactly 1", got)
	}
	if got[0] != "PATCH /api/platform/v1/organizations/org-1/projects/proj-1/files/f-1" {
		t.Errorf("request = %q, want a files PATCH", got[0])
	}
}

func TestFilesUpdate_NameIsRequiredWithoutLocalFile(t *testing.T) {
	requests := new(filesRequestLog)
	server := filesUploadTestServer(t, requests)
	setupFilesUpdateTest(t, server)

	err := filesUpdateCmd.Args(filesUpdateCmd, []string{"f-1"})
	if err == nil {
		t.Fatal("files update f-1: want an error, got nil")
	}
	if !strings.Contains(err.Error(), "--name is required") {
		t.Errorf("error = %q, want it to require --name", err)
	}
	if got := requests.all(); len(got) != 0 {
		t.Errorf("requests = %v, want none", got)
	}
}

// A FIFO blocks in os.Open until a writer appears, so the deadline distinguishes
// "rejected by the guard" from "hung before the guard".
func TestFilesUpload_FifoOperandIsRefusedWithoutOpening(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "pipe")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Skipf("mkfifo unsupported here: %v", err)
	}

	requests := &filesRequestLog{}
	server := filesUploadTestServer(t, requests)
	setupFilesUploadTest(t, server)

	done := make(chan error, 1)
	go func() { done <- filesUploadCmd.RunE(filesUploadCmd, []string{fifo}) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatalf("files upload %s: want an error, got nil", fifo)
		}
		if !strings.Contains(err.Error(), "is not a regular file") {
			t.Errorf("error = %q, want it to reject the operand as not a regular file", err)
		}
		if seen := requests.all(); len(seen) != 0 {
			t.Errorf("requests = %v, want none", seen)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("files upload blocked on a FIFO; the operand is being opened before it is checked")
	}
}
