package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func dirEntryNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names
}

func assertNoNewEntries(t *testing.T, dir string, before []string) {
	t.Helper()
	known := make(map[string]bool, len(before))
	for _, name := range before {
		known[name] = true
	}
	for _, name := range dirEntryNames(t, dir) {
		if !known[name] {
			t.Errorf("%s gained entry %q; the download escaped its working directory", dir, name)
		}
	}
}

func TestFilesDownload_UnnamedResponseWithTraversalRef(t *testing.T) {
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{body: body})

	cwd := t.TempDir()
	t.Chdir(cwd)
	setupFilesDownloadTest(t, server)

	parent := filepath.Dir(cwd)
	before := dirEntryNames(t, parent)

	err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"../escaped.txt"})
	if err != nil {
		t.Fatalf("files download ../escaped.txt: %v", err)
	}

	assertNoNewEntries(t, parent, before)
	if _, statErr := os.Stat(filepath.Join(cwd, "escaped.txt")); statErr != nil {
		t.Errorf("files download reported success but %s holds no sanitised file: %v", cwd, statErr)
	}
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_UnnamedResponseWithSubdirectoryRef(t *testing.T) {
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{body: body})

	cwd := t.TempDir()
	t.Chdir(cwd)
	reports := filepath.Join(cwd, "reports")
	if err := os.Mkdir(reports, 0o755); err != nil {
		t.Fatalf("mkdir reports: %v", err)
	}
	setupFilesDownloadTest(t, server)
	seen := captureDownloadRename(t)

	err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"reports/q3.pdf"})
	if err != nil {
		t.Fatalf("files download reports/q3.pdf: %v", err)
	}

	if names := dirEntryNames(t, reports); len(names) != 0 {
		t.Errorf("reports/ gained %v; the ref was used as a local path", names)
	}
	got, readErr := os.ReadFile(filepath.Join(cwd, "q3.pdf"))
	if readErr != nil {
		t.Errorf("files download reported success but %s holds no q3.pdf: %v", cwd, readErr)
	} else if string(got) != string(body) {
		t.Errorf("q3.pdf contents = %q, want %q", got, body)
	}
	for _, pair := range *seen {
		if filepath.Dir(pair[0]) != filepath.Dir(pair[1]) {
			t.Errorf("renamed across directories: %q -> %q; EXDEV is reachable", pair[0], pair[1])
		}
	}
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_UnnamedResponseWithAbsoluteRef(t *testing.T) {
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{body: body})

	cwd := t.TempDir()
	outside := t.TempDir()
	t.Chdir(cwd)
	setupFilesDownloadTest(t, server)

	ref := filepath.Join(outside, "iai-red-abs.txt")
	err := filesDownloadCmd.RunE(filesDownloadCmd, []string{ref})
	if err != nil {
		t.Fatalf("files download %s: %v", ref, err)
	}

	if _, statErr := os.Stat(ref); statErr == nil {
		t.Errorf("%s was created; an absolute ref was used as the download target", ref)
	}
	if names := dirEntryNames(t, outside); len(names) != 0 {
		t.Errorf("%s gained %v; the download escaped its working directory", outside, names)
	}
	if _, statErr := os.Stat(filepath.Join(cwd, "iai-red-abs.txt")); statErr != nil {
		t.Errorf("files download reported success but %s holds no sanitised file: %v", cwd, statErr)
	}
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_OutputDirectoryIsRefused(t *testing.T) {
	for _, tt := range []struct {
		name  string
		force bool
	}{
		{name: "without force"},
		{name: "with force", force: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var hits int
			server := filesDownloadTestServer(t, filesDownloadServer{
				body:         []byte("report bytes"),
				filename:     "stored.pdf",
				downloadHits: &hits,
			})

			cwd := t.TempDir()
			t.Chdir(cwd)
			target := filepath.Join(cwd, "somedir")
			if err := os.Mkdir(target, 0o755); err != nil {
				t.Fatalf("mkdir target: %v", err)
			}
			setupFilesDownloadTest(t, server)
			if err := filesDownloadCmd.Flags().Set("output", target); err != nil {
				t.Fatalf("set --output: %v", err)
			}
			if tt.force {
				if err := filesDownloadCmd.Flags().Set("force", "true"); err != nil {
					t.Fatalf("set --force: %v", err)
				}
			}

			err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"})
			if err == nil {
				t.Fatalf("files download --output %s: want an error, got nil", target)
			}
			if !tt.force && strings.Contains(err.Error(), "use --force to overwrite") {
				t.Errorf("error = %q, want no --force advice for a directory target", err)
			}
			if !strings.Contains(err.Error(), "directory") {
				t.Errorf("error = %q, want it to say the target is a directory", err)
			}
			if hits != 0 {
				t.Errorf(
					"download route hits = %d, want 0; the body was transferred and thrown away",
					hits,
				)
			}
			assertNoStagedTemps(t, cwd)
		})
	}
}

func TestFilesDownload_OutputInMissingDirectoryReportsUserPath(t *testing.T) {
	server := filesDownloadTestServer(t, filesDownloadServer{
		body:     []byte("report bytes"),
		filename: "stored.pdf",
	})

	cwd := t.TempDir()
	t.Chdir(cwd)
	setupFilesDownloadTest(t, server)
	if err := filesDownloadCmd.Flags().Set("output", "nope/report.pdf"); err != nil {
		t.Fatalf("set --output: %v", err)
	}

	err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"})
	if err == nil {
		t.Fatalf("files download --output nope/report.pdf: want an error, got nil")
	}
	if !strings.Contains(err.Error(), "nope/report.pdf") {
		t.Errorf("error = %q, want it to name the path the user typed", err)
	}
	if strings.Contains(err.Error(), ".iai-files-download-") {
		t.Errorf("error = %q, want no staging-file path the user never typed", err)
	}
	assertNoStagedTemps(t, cwd)
}

func TestFilesDownload_UnnamedResponseWithSafeRefLandsInCwd(t *testing.T) {
	body := []byte("report bytes")
	server := filesDownloadTestServer(t, filesDownloadServer{body: body})

	cwd := t.TempDir()
	t.Chdir(cwd)
	setupFilesDownloadTest(t, server)

	if err := filesDownloadCmd.RunE(filesDownloadCmd, []string{"f-1"}); err != nil {
		t.Fatalf("files download f-1: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(cwd, "f-1"))
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("target contents = %q, want %q", got, body)
	}
	assertNoStagedTemps(t, cwd)
}
