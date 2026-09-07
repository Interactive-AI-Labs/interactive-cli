package cmd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestSafeDownloadFilename(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		fileID  string
		want    string
		wantErr bool
	}{
		{name: "a normal name", raw: "report.pdf", fileID: "file-1", want: "report.pdf"},
		{
			name:   "a traversal path reduces to its base",
			raw:    "../../evil.txt",
			fileID: "file-1",
			want:   "evil.txt",
		},
		{name: "a bare dot-dot falls back to the id", raw: "..", fileID: "file-1", want: "file-1"},
		{name: "a bare dot falls back to the id", raw: ".", fileID: "file-1", want: "file-1"},
		{name: "empty falls back to the id", raw: "", fileID: "file-1", want: "file-1"},
		{name: "a bare separator falls back to the id", raw: "/", fileID: "file-1", want: "file-1"},
		{
			name:   "an absolute path falls back to the id",
			raw:    "/etc/passwd",
			fileID: "file-1",
			want:   "passwd",
		},
		{
			name:   "a subdirectory path reduces to its base",
			raw:    "reports/q3.pdf",
			fileID: "file-1",
			want:   "q3.pdf",
		},
		{
			name:   "a trailing separator reduces to the last element",
			raw:    "reports/",
			fileID: "file-1",
			want:   "reports",
		},

		{
			name:   "a traversal fallback reduces to its base",
			fileID: "../escaped.txt",
			want:   "escaped.txt",
		},
		{
			name:   "a subdirectory fallback reduces to its base",
			fileID: "reports/q3.pdf",
			want:   "q3.pdf",
		},
		{name: "an absolute fallback reduces to its base", fileID: "/tmp/abs.txt", want: "abs.txt"},
		{name: "a dot-dot fallback is not usable", fileID: "..", wantErr: true},
		{name: "a separator fallback is not usable", fileID: "/", wantErr: true},
		{name: "both empty is not usable", wantErr: true},

		{
			name:   "a carriage return in the name falls back",
			raw:    "a\rmalware.sh",
			fileID: "file-1",
			want:   "file-1",
		},
		{
			name:   "a newline in the name falls back",
			raw:    "a\nb.txt",
			fileID: "file-1",
			want:   "file-1",
		},
		{
			name:   "an escape in the name falls back",
			raw:    "a\x1b[2Jb.txt",
			fileID: "file-1",
			want:   "file-1",
		},
		{name: "a NUL in the name falls back", raw: "a\x00b.txt", fileID: "file-1", want: "file-1"},
		{name: "a DEL in the name falls back", raw: "a\x7fb.txt", fileID: "file-1", want: "file-1"},
		{
			name:    "control characters in both is not usable",
			raw:     "a\rb",
			fileID:  "c\x1bd",
			wantErr: true,
		},
		{name: "a tab in the name falls back", raw: "a\tb.txt", fileID: "file-1", want: "file-1"},
		{
			name:   "legitimate unicode is kept",
			raw:    "отчёт-第三.pdf",
			fileID: "file-1",
			want:   "отчёт-第三.pdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeDownloadFilename(tt.raw, tt.fileID)
			if tt.wantErr {
				if err == nil {
					t.Fatalf(
						"safeDownloadFilename(%q, %q) = %q, want an error",
						tt.raw,
						tt.fileID,
						got,
					)
				}
				return
			}
			if err != nil {
				t.Fatalf("safeDownloadFilename(%q, %q) error = %v", tt.raw, tt.fileID, err)
			}
			if got != tt.want {
				t.Errorf(
					"safeDownloadFilename(%q, %q) = %q, want %q",
					tt.raw,
					tt.fileID,
					got,
					tt.want,
				)
			}
		})
	}
}

type filesDownloadServer struct {
	body        []byte
	filename    string
	midTransfer func()
	// downloadHits counts only the files route, so lookups do not inflate it.
	downloadHits *int
}

func filesDownloadTestServer(t *testing.T, cfg filesDownloadServer) *httptest.Server {
	t.Helper()
	base := "/api/platform/v1/organizations/org-1/projects/proj-1/files"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeFilesTestSessionResponse(w, r) {
			return
		}
		switch {
		case strings.HasPrefix(r.URL.Path, base):
			if cfg.downloadHits != nil {
				*cfg.downloadHits++
			}
			if cfg.filename != "" {
				w.Header().Set("Content-Disposition", `attachment; filename="`+cfg.filename+`"`)
			}
			if cfg.midTransfer == nil {
				w.Write(cfg.body)
				return
			}
			half := len(cfg.body) / 2
			w.Write(cfg.body[:half])
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Errorf("response writer is not an http.Flusher")
				return
			}
			flusher.Flush()
			cfg.midTransfer()
			w.Write(cfg.body[half:])
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func setupFilesDownloadTest(t *testing.T, server *httptest.Server) {
	t.Helper()
	setupFilesCommandTest(t, server, filesDownloadCmd)
	origTimeout := filesDownloadTimeout
	t.Cleanup(func() {
		filesDownloadOrg, filesDownloadProject = "", ""
		filesDownloadVersion, filesDownloadOutput = "", ""
		filesDownloadForce = false
		filesDownloadTimeout = origTimeout
		for _, name := range []string{"version", "output", "force", "timeout"} {
			filesDownloadCmd.Flags().Lookup(name).Changed = false
		}
	})
	filesDownloadOrg, filesDownloadProject = "acme", "alunafi"
}

func stagedTempEntries(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var found []string
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".iai-files-download-") {
			found = append(found, entry.Name())
		}
	}
	return found, nil
}

func assertNoStagedTemps(t *testing.T, dir string) {
	t.Helper()
	found, err := stagedTempEntries(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	if len(found) != 0 {
		t.Errorf("leftover temp files in %s: %v", dir, found)
	}
}

func TestFilesDownload_OutputElsewhereWithReadOnlyCwd(t *testing.T) {
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{body: body, filename: "stored.pdf"})

	cwd := t.TempDir()
	targetDir := t.TempDir()
	t.Chdir(cwd)
	t.Cleanup(func() { os.Chmod(cwd, 0o755) })
	if err := os.Chmod(cwd, 0o555); err != nil {
		t.Fatalf("chmod cwd read-only: %v", err)
	}

	setupFilesDownloadTest(t, server)
	target := filepath.Join(targetDir, "report.pdf")
	if err := filesDownloadCmd.Flags().Set("output", target); err != nil {
		t.Fatalf("set --output: %v", err)
	}

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download --output %s: %v", target, err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("target contents = %q, want %q", got, body)
	}
}

func TestFilesDownload_StagesTempInTargetDirectory(t *testing.T) {
	cwd := t.TempDir()
	targetDir := t.TempDir()
	t.Chdir(cwd)

	var midTargetTemps, midCwdTemps []string
	var midErr error
	body := []byte("some reasonably long body so the halves are non-empty")
	server := filesDownloadTestServer(t, filesDownloadServer{
		body:     body,
		filename: "stored.pdf",
		midTransfer: func() {
			midTargetTemps, midErr = stagedTempEntries(targetDir)
			if midErr != nil {
				return
			}
			midCwdTemps, midErr = stagedTempEntries(cwd)
		},
	})

	setupFilesDownloadTest(t, server)
	target := filepath.Join(targetDir, "report.pdf")
	if err := filesDownloadCmd.Flags().Set("output", target); err != nil {
		t.Fatalf("set --output: %v", err)
	}

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download --output %s: %v", target, err)
	}
	if midErr != nil {
		t.Fatalf("mid-transfer scan: %v", midErr)
	}

	if len(midTargetTemps) != 1 {
		t.Errorf("mid-transfer temps in target dir = %v, want exactly 1", midTargetTemps)
	}
	if len(midCwdTemps) != 0 {
		t.Errorf("mid-transfer temps in cwd = %v, want none", midCwdTemps)
	}
	assertNoStagedTemps(t, targetDir)
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_OutputBareFilenameLandsInCwd(t *testing.T) {
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{body: body, filename: "stored.pdf"})

	cwd := t.TempDir()
	t.Chdir(cwd)
	setupFilesDownloadTest(t, server)
	if err := filesDownloadCmd.Flags().Set("output", "report.pdf"); err != nil {
		t.Fatalf("set --output: %v", err)
	}

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download --output report.pdf: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(cwd, "report.pdf"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("target contents = %q, want %q", got, body)
	}
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_NoOutputUsesServerFilename(t *testing.T) {
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{body: body, filename: "stored.pdf"})

	cwd := t.TempDir()
	t.Chdir(cwd)
	setupFilesDownloadTest(t, server)

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(cwd, "stored.pdf"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("target contents = %q, want %q", got, body)
	}
	assertNoStagedTemps(t, cwd)
}

func captureDownloadRename(t *testing.T) *[][2]string {
	t.Helper()
	var seen [][2]string
	orig := renameDownload
	t.Cleanup(func() { renameDownload = orig })
	renameDownload = func(oldpath, newpath string) error {
		seen = append(seen, [2]string{oldpath, newpath})
		return orig(oldpath, newpath)
	}
	return &seen
}

func TestFilesDownload_RenameNeverCrossesDirectories(t *testing.T) {
	tests := []struct {
		name          string
		outputInOther bool
		output        string
	}{
		{"output in another directory", true, ""},
		{"output as a bare filename", false, "report.pdf"},
		{"no output at all", false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := []byte("report bytes")
			server := filesDownloadTestServer(
				t,
				filesDownloadServer{body: body, filename: "stored.pdf"},
			)

			cwd := t.TempDir()
			t.Chdir(cwd)
			setupFilesDownloadTest(t, server)
			seen := captureDownloadRename(t)

			output := tt.output
			if tt.outputInOther {
				output = filepath.Join(t.TempDir(), "report.pdf")
			}
			if output != "" {
				if err := filesDownloadCmd.Flags().Set("output", output); err != nil {
					t.Fatalf("set --output: %v", err)
				}
			}

			if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
				t.Fatalf("files download: %v", err)
			}

			if len(*seen) != 1 {
				t.Fatalf("renames = %v, want exactly 1", *seen)
			}
			from, to := (*seen)[0][0], (*seen)[0][1]
			if filepath.Dir(from) != filepath.Dir(to) {
				t.Errorf("renamed across directories: %q -> %q; EXDEV is reachable", from, to)
			}
		})
	}
}

func deviceOf(t *testing.T, path string) uint64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Skipf("%s: no syscall.Stat_t on this platform", path)
	}
	return uint64(st.Dev)
}

func TestFilesDownload_TargetOnAnotherFilesystem(t *testing.T) {
	otherFS := os.Getenv("IAI_TEST_XDEV_DIR")
	if otherFS == "" {
		t.Skip("set IAI_TEST_XDEV_DIR to a writable directory on a different filesystem")
	}

	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{body: body, filename: "stored.pdf"})

	cwd := t.TempDir()
	t.Chdir(cwd)
	if deviceOf(t, cwd) == deviceOf(t, otherFS) {
		t.Fatalf(
			"IAI_TEST_XDEV_DIR (%s) is on the same filesystem as the test cwd; this test would prove nothing",
			otherFS,
		)
	}

	setupFilesDownloadTest(t, server)
	target := filepath.Join(otherFS, "iai-xdev-download.pdf")
	t.Cleanup(func() { os.Remove(target) })
	if err := filesDownloadCmd.Flags().Set("output", target); err != nil {
		t.Fatalf("set --output: %v", err)
	}

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download --output %s: %v", target, err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("target contents = %q, want %q", got, body)
	}
	assertNoStagedTemps(t, otherFS)
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_ExistingOutputTargetIsRefusedBeforeTransfer(t *testing.T) {
	var hits int
	server := filesDownloadTestServer(t, filesDownloadServer{
		body:         []byte("new bytes"),
		filename:     "stored.pdf",
		downloadHits: &hits,
	})

	cwd := t.TempDir()
	targetDir := t.TempDir()
	t.Chdir(cwd)

	target := filepath.Join(targetDir, "report.pdf")
	existing := []byte("existing bytes")
	if err := os.WriteFile(target, existing, 0o644); err != nil {
		t.Fatalf("write existing target: %v", err)
	}

	setupFilesDownloadTest(t, server)
	if err := filesDownloadCmd.Flags().Set("output", target); err != nil {
		t.Fatalf("set --output: %v", err)
	}

	err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"})
	if err == nil {
		t.Fatalf("files download --output %s: want an error, got nil", target)
	}
	if !strings.Contains(err.Error(), "already exists; use --force to overwrite") {
		t.Errorf("error = %q, want it to mention already exists; use --force to overwrite", err)
	}

	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(got) != string(existing) {
		t.Errorf("target contents = %q, want the untouched %q", got, existing)
	}
	assertNoStagedTemps(t, targetDir)
	assertNoStagedTemps(t, cwd)

	if hits != 0 {
		t.Errorf("download route hits = %d, want 0; the body was transferred and thrown away", hits)
	}
}

func TestRefuseExistingTarget_StatFailureNamesTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "self-referential-link")
	if err := os.Symlink(filepath.Base(target), target); err != nil {
		t.Skipf("create self-referential symlink: %v", err)
	}
	if _, err := os.Stat(target); err == nil || errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stat self-referential symlink error = %v, want a non-ENOENT error", err)
	}

	err := refuseExistingTarget(target, false)
	if err == nil {
		t.Fatal("refuseExistingTarget error = nil, want the stat failure")
	}
	if !strings.Contains(err.Error(), target) {
		t.Errorf("error = %q, want it to name target %q", err, target)
	}
	if !strings.Contains(err.Error(), "failed to inspect") {
		t.Errorf("error = %q, want stat context", err)
	}
}

func TestFilesDownload_ExistingOutputTargetWithForceOverwrites(t *testing.T) {
	var hits int
	body := []byte("new bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{
		body:         body,
		filename:     "stored.pdf",
		downloadHits: &hits,
	})

	cwd := t.TempDir()
	targetDir := t.TempDir()
	t.Chdir(cwd)

	target := filepath.Join(targetDir, "report.pdf")
	if err := os.WriteFile(target, []byte("existing bytes"), 0o644); err != nil {
		t.Fatalf("write existing target: %v", err)
	}

	setupFilesDownloadTest(t, server)
	if err := filesDownloadCmd.Flags().Set("output", target); err != nil {
		t.Fatalf("set --output: %v", err)
	}
	if err := filesDownloadCmd.Flags().Set("force", "true"); err != nil {
		t.Fatalf("set --force: %v", err)
	}

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download --output %s --force: %v", target, err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("target contents = %q, want %q", got, body)
	}
	if hits != 1 {
		t.Errorf("download route hits = %d, want exactly 1", hits)
	}
	assertNoStagedTemps(t, targetDir)
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_ExistingServerNamedTargetIsRefusedAfterTransfer(t *testing.T) {
	var hits int
	server := filesDownloadTestServer(t, filesDownloadServer{
		body:         []byte("new bytes"),
		filename:     "stored.pdf",
		downloadHits: &hits,
	})

	cwd := t.TempDir()
	t.Chdir(cwd)

	target := filepath.Join(cwd, "stored.pdf")
	existing := []byte("existing bytes")
	if err := os.WriteFile(target, existing, 0o644); err != nil {
		t.Fatalf("write existing target: %v", err)
	}

	setupFilesDownloadTest(t, server)

	err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"})
	if err == nil {
		t.Fatalf("files download: want an error, got nil")
	}
	if !strings.Contains(err.Error(), "already exists; use --force to overwrite") {
		t.Errorf("error = %q, want it to mention already exists; use --force to overwrite", err)
	}

	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("read target: %v", readErr)
	}
	if string(got) != string(existing) {
		t.Errorf("target contents = %q, want the untouched %q", got, existing)
	}
	if hits != 1 {
		t.Errorf(
			"download route hits = %d, want exactly 1; the name is only known from the response",
			hits,
		)
	}
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_MissingOutputTargetStillDownloads(t *testing.T) {
	var hits int
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{
		body:         body,
		filename:     "stored.pdf",
		downloadHits: &hits,
	})

	cwd := t.TempDir()
	targetDir := t.TempDir()
	t.Chdir(cwd)
	setupFilesDownloadTest(t, server)

	target := filepath.Join(targetDir, "report.pdf")
	if err := filesDownloadCmd.Flags().Set("output", target); err != nil {
		t.Fatalf("set --output: %v", err)
	}

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download --output %s: %v", target, err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("target contents = %q, want %q", got, body)
	}
	if hits != 1 {
		t.Errorf("download route hits = %d, want exactly 1", hits)
	}
	assertNoStagedTemps(t, targetDir)
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_StdoutUnaffectedByFileNamedDash(t *testing.T) {
	var hits int
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{
		body:         body,
		filename:     "stored.pdf",
		downloadHits: &hits,
	})

	cwd := t.TempDir()
	t.Chdir(cwd)

	dashPath := filepath.Join(cwd, "-")
	existing := []byte("existing bytes")
	if err := os.WriteFile(dashPath, existing, 0o644); err != nil {
		t.Fatalf("write file named -: %v", err)
	}

	setupFilesDownloadTest(t, server)
	stdout := new(strings.Builder)
	filesDownloadCmd.SetOut(stdout)
	if err := filesDownloadCmd.Flags().Set("output", "-"); err != nil {
		t.Fatalf("set --output -: %v", err)
	}

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download --output -: %v", err)
	}

	if stdout.String() != string(body) {
		t.Errorf("stdout = %q, want %q", stdout.String(), body)
	}
	got, readErr := os.ReadFile(dashPath)
	if readErr != nil {
		t.Fatalf("read file named -: %v", readErr)
	}
	if string(got) != string(existing) {
		t.Errorf("file named - contents = %q, want the untouched %q", got, existing)
	}
	if hits != 1 {
		t.Errorf("download route hits = %d, want exactly 1", hits)
	}
	assertNoStagedTemps(t, cwd)
}
