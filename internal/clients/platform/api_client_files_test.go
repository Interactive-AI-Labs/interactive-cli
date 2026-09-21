package platform

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func writeTempFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestCreateFile_DeclaresContentLength(t *testing.T) {
	content := []byte("hello world, this is the file body")
	localPath := writeTempFile(t, "report.txt", content)

	var (
		gotTransferEncoding []string
		gotContentLength    int64
		gotBodyLen          int
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTransferEncoding = r.TransferEncoding
		gotContentLength = r.ContentLength

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		gotBodyLen = len(body)

		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), "report.txt")
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	result, err := client.CreateFile(
		context.Background(), "org-1", "proj-1", localPath, "", nil,
	)
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	if result.ID != "file-1" {
		t.Errorf("result.ID = %q, want file-1", result.ID)
	}

	if len(gotTransferEncoding) != 0 {
		t.Errorf(
			"TransferEncoding = %v, want empty (declared length, not chunked)",
			gotTransferEncoding,
		)
	}
	if gotContentLength <= 0 {
		t.Fatalf("ContentLength = %d, want a positive declared length", gotContentLength)
	}
	if gotContentLength != int64(gotBodyLen) {
		t.Errorf(
			"ContentLength = %d, want it to equal the actual body length %d",
			gotContentLength, gotBodyLen,
		)
	}
}

func TestCreateFile_QuotedFilename(t *testing.T) {
	content := []byte("quoted filename body")
	localPath := writeTempFile(t, `weird"name.txt`, content)

	var gotFilename string
	var gotContent []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("failed to read file part: %v", err)
		}
		defer file.Close()
		gotFilename = header.Filename
		gotContent, err = io.ReadAll(file)
		if err != nil {
			t.Fatalf("failed to read file part body: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), gotFilename)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	_, err := client.CreateFile(context.Background(), "org-1", "proj-1", localPath, "", nil)
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}

	if gotFilename != `weird"name.txt` {
		t.Errorf("server-observed filename = %q, want %q", gotFilename, `weird"name.txt`)
	}
	if !bytes.Equal(gotContent, content) {
		t.Errorf("server-observed content = %q, want %q", gotContent, content)
	}
}

func TestCreateFile_Name(t *testing.T) {
	content := []byte("named upload")
	localPath := writeTempFile(t, "local-name.txt", content)

	var gotName string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		gotName = r.FormValue("name")

		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), gotName)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	result, err := client.CreateFile(
		context.Background(), "org-1", "proj-1", localPath, "custom-name.txt", nil,
	)
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}

	if gotName != "custom-name.txt" {
		t.Errorf("server-observed name field = %q, want custom-name.txt", gotName)
	}
	if result.Current.Name != "custom-name.txt" {
		t.Errorf("result.Current.Name = %q, want custom-name.txt", result.Current.Name)
	}
}

// Guards the hand-rolled header/footer splice around the streamed body: it proves
// the bytes arrive unaltered, not anything about server-side identity or dedup.
func TestCreateFile_FramingPreservesFileBytes(t *testing.T) {
	content := bytes.Repeat([]byte("interactiveai"), 1000)
	localPath := writeTempFile(t, "data.bin", content)
	wantSum := sha256.Sum256(content)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("failed to read file part: %v", err)
		}
		defer file.Close()
		got, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("failed to read file part body: %v", err)
		}
		gotSum := sha256.Sum256(got)
		if gotSum != wantSum {
			t.Errorf("sha256 of the received file part = %x, want %x", gotSum, wantSum)
		}

		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), "data.bin")
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	if _, err := client.CreateFile(
		context.Background(),
		"org-1",
		"proj-1",
		localPath,
		"",
		nil,
	); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
}

func TestListFiles_QueryParams(t *testing.T) {
	const path = "/api/platform/v1/organizations/org-1/projects/proj-1/files"
	const body = `{"success":true,"data":{"files":[],"meta":{"cursor":null,"total_items":null,"limit":50}}}`

	intPtr := func(n int) *int { return &n }

	tests := []struct {
		name      string
		opts      FileListOptions
		wantQuery url.Values
	}{
		{
			name:      "omits everything unset",
			opts:      FileListOptions{},
			wantQuery: url.Values{},
		},
		{
			name:      "sends limit and cursor",
			opts:      FileListOptions{Limit: intPtr(10), Cursor: "abc"},
			wantQuery: url.Values{"limit": {"10"}, "cursor": {"abc"}},
		},
		{
			name:      "sends includeVersions",
			opts:      FileListOptions{IncludeVersions: true},
			wantQuery: url.Values{"includeVersions": {"true"}},
		},
		{
			name:      "sends an explicit zero limit as a written zero",
			opts:      FileListOptions{Limit: intPtr(0)},
			wantQuery: url.Values{"limit": {"0"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newListTestServer(t, path, tt.wantQuery, body)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			if _, _, _, err := client.ListFiles(
				context.Background(),
				"org-1",
				"proj-1",
				tt.opts,
			); err != nil {
				t.Fatalf("ListFiles() error = %v", err)
			}
		})
	}
}

func TestListFiles_DecodesNameAndEmail(t *testing.T) {
	const path = "/api/platform/v1/organizations/org-1/projects/proj-1/files"
	const body = `{"success":true,"data":{"files":[` +
		`{"id":"f-1","name":"report.pdf","current":{"versionId":"v-1","size":10,` +
		`"name":"report.pdf","contentType":"application/pdf","createdAt":"2026-01-01T00:00:00Z",` +
		`"createdBy":"user-1","createdByEmail":"a@example.com","isCurrent":true}},` +
		`{"id":"f-2","name":"notes.txt","current":{"versionId":"v-2","size":5,` +
		`"name":"notes.txt","contentType":"text/plain","createdAt":"2026-01-02T00:00:00Z",` +
		`"createdBy":"pk-abc","createdByEmail":null,"isCurrent":true}}` +
		`],"meta":{"cursor":"next-page","total_items":null,"limit":50}}}`

	server := newListTestServer(t, path, url.Values{}, body)
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	files, meta, _, err := client.ListFiles(
		context.Background(),
		"org-1",
		"proj-1",
		FileListOptions{},
	)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2", len(files))
	}
	if files[0].Name != "report.pdf" {
		t.Errorf("files[0].Name = %q, want report.pdf", files[0].Name)
	}
	if files[0].Current.CreatedByEmail == nil ||
		*files[0].Current.CreatedByEmail != "a@example.com" {
		t.Errorf(
			"files[0].Current.CreatedByEmail = %v, want a@example.com",
			files[0].Current.CreatedByEmail,
		)
	}
	if files[1].Current.CreatedByEmail != nil {
		t.Errorf("files[1].Current.CreatedByEmail = %v, want nil", *files[1].Current.CreatedByEmail)
	}
	if meta.Cursor == nil || *meta.Cursor != "next-page" {
		t.Errorf("meta.Cursor = %v, want next-page", meta.Cursor)
	}
}

func writeFileMetadataResponse(
	w http.ResponseWriter,
	id, versionID string,
	size int64,
	name string,
) {
	resp := struct {
		Success bool         `json:"success"`
		Data    FileMetadata `json:"data"`
	}{
		Success: true,
		Data: FileMetadata{
			ID: id,
			Current: FileVersion{
				VersionID:   versionID,
				Size:        size,
				Name:        name,
				ContentType: "application/octet-stream",
				CreatedAt:   time.Now().UTC(),
				CreatedBy:   "user-1",
				IsCurrent:   true,
			},
		},
	}
	b, err := json.Marshal(resp)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal test response: %v", err))
	}
	_, _ = w.Write(b)
}

func TestDownloadFile_ParsesFilenameAndStreamsBytes(t *testing.T) {
	content := []byte("the current version's bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/files/file-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().
			Set("Content-Disposition", `attachment; filename="evil.txt"; filename*=UTF-8''%2E%2E`)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	var buf bytes.Buffer
	var gotProgress int64
	filename, length, err := client.DownloadFile(
		context.Background(), "org-1", "proj-1", "file-1", &buf,
		func(n int64) { gotProgress += n }, nil,
	)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v", err)
	}
	if filename != ".." {
		t.Errorf("filename = %q, want %q (raw, unreduced)", filename, "..")
	}
	if length != int64(len(content)) {
		t.Errorf("length = %d, want %d", length, len(content))
	}
	if !bytes.Equal(buf.Bytes(), content) {
		t.Errorf("downloaded bytes = %q, want %q", buf.Bytes(), content)
	}
	if gotProgress != int64(len(content)) {
		t.Errorf("progress reported = %d, want %d", gotProgress, len(content))
	}
}

func TestDownloadFileVersion_HitsVersionPath(t *testing.T) {
	content := []byte("an older version's bytes")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/files/file-1/versions/v-0" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().
			Set("Content-Disposition", `attachment; filename="report.txt"; filename*=UTF-8''report.txt`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	var buf bytes.Buffer
	filename, _, err := client.DownloadFileVersion(
		context.Background(), "org-1", "proj-1", "file-1", "v-0", &buf, nil, nil,
	)
	if err != nil {
		t.Fatalf("DownloadFileVersion() error = %v", err)
	}
	if filename != "report.txt" {
		t.Errorf("filename = %q, want report.txt", filename)
	}
	if !bytes.Equal(buf.Bytes(), content) {
		t.Errorf("downloaded bytes = %q, want %q", buf.Bytes(), content)
	}
}

func TestDownloadFile_OnStartReceivesContentLength(t *testing.T) {
	content := []byte("bytes whose length the server declares up front")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	var buf bytes.Buffer
	var onStartCalls int
	var gotTotal int64
	_, _, err := client.DownloadFile(
		context.Background(), "org-1", "proj-1", "file-1", &buf, nil,
		func(total int64) {
			onStartCalls++
			gotTotal = total
		},
	)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v", err)
	}
	if onStartCalls != 1 {
		t.Fatalf("onStart called %d times, want exactly 1", onStartCalls)
	}
	if gotTotal != int64(len(content)) {
		t.Errorf("onStart total = %d, want %d", gotTotal, len(content))
	}
}

func TestDownloadFile_OnStartSkippedWhenContentLengthUnknown(t *testing.T) {
	content := []byte("bytes sent without a declared content length")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			// Flushing before the body is written forces chunked transfer
			// encoding, so the client sees an unknown (-1) Content-Length.
			f.Flush()
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	var buf bytes.Buffer
	var onStartCalls int
	_, _, err := client.DownloadFile(
		context.Background(), "org-1", "proj-1", "file-1", &buf, nil,
		func(total int64) { onStartCalls++ },
	)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v", err)
	}
	if onStartCalls != 0 {
		t.Errorf("onStart called %d times, want 0 (content length was unknown)", onStartCalls)
	}
}

func TestDownloadFile_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":{"error":{"message":"file not found"}}}`))
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	var buf bytes.Buffer
	_, _, err := client.DownloadFile(
		context.Background(),
		"org-1",
		"proj-1",
		"missing",
		&buf,
		nil,
		nil,
	)
	if err == nil {
		t.Fatal("DownloadFile() error = nil, want a not-found error")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf(
			"DownloadFile() error = %q, want it to contain the server message %q",
			err,
			"file not found",
		)
	}
}

func TestDownloadFile_AmbiguousRef(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"detail":{"success":false,"error":{"code":"FILE_REF_AMBIGUOUS",` +
			`"message":"ambiguous","details":{"candidates":[` +
			`{"fileId":"f-1","name":"report.pdf","size":10,"createdAt":"2026-01-01T00:00:00Z"},` +
			`{"fileId":"f-2","name":"report.pdf","size":20,"createdAt":"2026-01-02T00:00:00Z"}` +
			`]}}}}`))
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	var buf bytes.Buffer
	_, _, err := client.DownloadFile(
		context.Background(),
		"org-1",
		"proj-1",
		"report.pdf",
		&buf,
		nil,
		nil,
	)

	var ambiguous *FileRefAmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("DownloadFile() error = %v, want a *FileRefAmbiguousError", err)
	}
	if len(ambiguous.Candidates) != 2 {
		t.Fatalf("len(Candidates) = %d, want 2", len(ambiguous.Candidates))
	}
	if ambiguous.Ref != "report.pdf" {
		t.Errorf("Ref = %q, want report.pdf", ambiguous.Ref)
	}
}

func TestGetFileMetadata_DecodesDetailEnvelope(t *testing.T) {
	const body = `{"success":true,"data":{"file":{"id":"f-1","name":"report.pdf",` +
		`"current":{"versionId":"v-3","size":30,"name":"report.pdf","contentType":"text/plain",` +
		`"createdAt":"2026-01-03T00:00:00Z","createdBy":"user-1","createdByEmail":null,"isCurrent":true},` +
		`"versions":[` +
		`{"versionId":"v-1","size":10,"name":"report.pdf","contentType":"text/plain",` +
		`"createdAt":"2026-01-01T00:00:00Z","createdBy":"user-1","createdByEmail":null,"isCurrent":false},` +
		`{"versionId":"v-2","size":20,"name":"report.pdf","contentType":"text/plain",` +
		`"createdAt":"2026-01-02T00:00:00Z","createdBy":"user-1","createdByEmail":null,"isCurrent":false},` +
		`{"versionId":"v-3","size":30,"name":"report.pdf","contentType":"text/plain",` +
		`"createdAt":"2026-01-03T00:00:00Z","createdBy":"user-1","createdByEmail":null,"isCurrent":true}` +
		`]}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/files/f-1/metadata" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	meta, _, err := client.GetFileMetadata(context.Background(), "org-1", "proj-1", "f-1")
	if err != nil {
		t.Fatalf("GetFileMetadata() error = %v", err)
	}

	ids := map[string]bool{}
	current := 0
	for _, v := range meta.Versions {
		ids[v.VersionID] = true
		if v.IsCurrent {
			current++
		}
	}
	if want := map[string]bool{"v-1": true, "v-2": true, "v-3": true}; len(ids) != len(want) {
		t.Errorf("version ids = %v, want %v", ids, want)
	}
	if current != 1 {
		t.Errorf("current count = %d, want 1", current)
	}
}

func TestListFiles_DecodesVersionsWhenIncluded(t *testing.T) {
	const body = `{"success":true,"data":{"files":[{"id":"f-1","name":"report.pdf",` +
		`"current":{"versionId":"v-3","size":30,"name":"report.pdf","contentType":"text/plain",` +
		`"createdAt":"2026-01-03T00:00:00Z","createdBy":"user-1","createdByEmail":null,"isCurrent":true},` +
		`"versions":[` +
		`{"versionId":"v-1","size":10,"name":"report.pdf","contentType":"text/plain",` +
		`"createdAt":"2026-01-01T00:00:00Z","createdBy":"user-1","createdByEmail":null,"isCurrent":false},` +
		`{"versionId":"v-2","size":20,"name":"report.pdf","contentType":"text/plain",` +
		`"createdAt":"2026-01-02T00:00:00Z","createdBy":"user-1","createdByEmail":null,"isCurrent":false},` +
		`{"versionId":"v-3","size":30,"name":"report.pdf","contentType":"text/plain",` +
		`"createdAt":"2026-01-03T00:00:00Z","createdBy":"user-1","createdByEmail":null,"isCurrent":true}` +
		`]}],"meta":{"cursor":null,"total_items":null,"limit":50}}}`

	server := newListTestServer(t, "/api/platform/v1/organizations/org-1/projects/proj-1/files",
		url.Values{"includeVersions": {"true"}}, body)
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	files, _, _, err := client.ListFiles(
		context.Background(), "org-1", "proj-1", FileListOptions{IncludeVersions: true},
	)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("len(files) = %d, want 1", len(files))
	}

	ids := map[string]bool{}
	current := 0
	for _, v := range files[0].Versions {
		ids[v.VersionID] = true
		if v.IsCurrent {
			current++
		}
	}
	if want := map[string]bool{"v-1": true, "v-2": true, "v-3": true}; len(ids) != len(want) {
		t.Errorf("version ids = %v, want %v", ids, want)
	}
	if current != 1 {
		t.Errorf("current count = %d, want 1", current)
	}
}

func TestAddFileVersion_NoNameByDefault(t *testing.T) {
	content := []byte("new version bytes")
	localPath := writeTempFile(t, "local-name.txt", content)

	var gotHasNameKey bool
	var gotContentLength int64
	var gotBodyLen int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/files/f-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotContentLength = r.ContentLength

		counted := &countingReader{r: r.Body, onRead: func(n int64) { gotBodyLen += int(n) }}
		r.Body = io.NopCloser(counted)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		_, gotHasNameKey = r.MultipartForm.Value["name"]

		w.WriteHeader(http.StatusOK)
		writeFileMetadataResponse(w, "f-1", "v-2", int64(len(content)), "stored-name.txt")
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	result, err := client.AddFileVersion(
		context.Background(),
		"org-1",
		"proj-1",
		"f-1",
		localPath,
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("AddFileVersion() error = %v", err)
	}
	if result.ID != "f-1" {
		t.Errorf("result.ID = %q, want f-1", result.ID)
	}
	if gotHasNameKey {
		t.Error("multipart form carries a name part, want none")
	}
	if gotContentLength <= 0 || gotContentLength != int64(gotBodyLen) {
		t.Errorf(
			"ContentLength = %d, want it to equal the actual body length %d",
			gotContentLength,
			gotBodyLen,
		)
	}
}

func TestAddFileVersion_WithName(t *testing.T) {
	content := []byte("new version bytes")
	localPath := writeTempFile(t, "local-name.txt", content)

	var gotName string
	var gotContentLength int64
	var gotBodyLen int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentLength = r.ContentLength
		counted := &countingReader{r: r.Body, onRead: func(n int64) { gotBodyLen += int(n) }}
		r.Body = io.NopCloser(counted)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}
		gotName = r.FormValue("name")

		w.WriteHeader(http.StatusOK)
		writeFileMetadataResponse(w, "f-1", "v-2", int64(len(content)), gotName)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	result, err := client.AddFileVersion(
		context.Background(), "org-1", "proj-1", "f-1", localPath, "renamed.txt", nil,
	)
	if err != nil {
		t.Fatalf("AddFileVersion() error = %v", err)
	}
	if gotName != "renamed.txt" {
		t.Errorf("server-observed name field = %q, want renamed.txt", gotName)
	}
	if result.Current.Name != "renamed.txt" {
		t.Errorf("result.Current.Name = %q, want renamed.txt", result.Current.Name)
	}
	if gotContentLength <= 0 || gotContentLength != int64(gotBodyLen) {
		t.Errorf(
			"ContentLength = %d, want it to equal the actual body length %d",
			gotContentLength,
			gotBodyLen,
		)
	}
}

func TestAddFileVersion_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":{"error":{"message":"File 'missing' not found"}}}`))
	}))
	defer server.Close()

	localPath := writeTempFile(t, "f.txt", []byte("x"))
	client := newEvalTestClient(t, server.URL)
	_, err := client.AddFileVersion(
		context.Background(),
		"org-1",
		"proj-1",
		"missing",
		localPath,
		"",
		nil,
	)
	if err == nil {
		t.Fatal("AddFileVersion() error = nil, want a not-found error")
	}
	var ambiguous *FileRefAmbiguousError
	if errors.As(err, &ambiguous) {
		t.Fatalf("AddFileVersion() error = %v, want a plain error, not ambiguous", err)
	}
	if err.Error() != "File 'missing' not found" {
		t.Errorf("error message = %q, want the server's message", err.Error())
	}
}

func TestAddFileVersion_AmbiguousRef(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"detail":{"success":false,"error":{"code":"FILE_REF_AMBIGUOUS",` +
			`"message":"ambiguous","details":{"candidates":[` +
			`{"fileId":"f-1","name":"report.pdf","size":10,"createdAt":"2026-01-01T00:00:00Z"},` +
			`{"fileId":"f-2","name":"report.pdf","size":20,"createdAt":"2026-01-02T00:00:00Z"}` +
			`]}}}}`))
	}))
	defer server.Close()

	localPath := writeTempFile(t, "f.txt", []byte("x"))
	client := newEvalTestClient(t, server.URL)
	_, err := client.AddFileVersion(
		context.Background(),
		"org-1",
		"proj-1",
		"report.pdf",
		localPath,
		"",
		nil,
	)

	var ambiguous *FileRefAmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("AddFileVersion() error = %v, want a *FileRefAmbiguousError", err)
	}
	if ambiguous.Ref != "report.pdf" {
		t.Errorf("Ref = %q, want report.pdf", ambiguous.Ref)
	}
}

func TestFileMutations_MethodPathAndResponse(t *testing.T) {
	const (
		filePath    = "/api/platform/v1/organizations/org-1/projects/proj-1/files/f-1"
		versionPath = filePath + "/versions/v-1"
	)
	tests := []struct {
		name            string
		method          string
		path            string
		requestName     string
		response        string
		invoke          func(*APIClient) (id, versionID, name, currentName string, err error)
		wantVersionID   string
		wantName        string
		wantCurrentName string
	}{
		{
			name:     "delete file",
			method:   http.MethodDelete,
			path:     filePath,
			response: `{"success":true,"data":{"id":"f-1","versionId":null}}`,
			invoke: func(client *APIClient) (string, string, string, string, error) {
				id, err := client.DeleteFile(context.Background(), "org-1", "proj-1", "f-1")
				return id, "", "", "", err
			},
		},
		{
			name:     "delete version",
			method:   http.MethodDelete,
			path:     versionPath,
			response: `{"success":true,"data":{"id":"f-1","versionId":"v-1"}}`,
			invoke: func(client *APIClient) (string, string, string, string, error) {
				id, err := client.DeleteFileVersion(
					context.Background(), "org-1", "proj-1", "f-1", "v-1",
				)
				return id, "", "", "", err
			},
		},
		{
			name:   "restore version",
			method: http.MethodPost,
			path:   versionPath,
			response: `{"success":true,"data":{"id":"f-1","name":"report.pdf",` +
				`"current":{"versionId":"v-2","size":42,"name":"report.pdf",` +
				`"contentType":"application/pdf","createdAt":"2026-01-02T00:00:00Z",` +
				`"createdBy":"user-1","isCurrent":true}}}`,
			invoke: func(client *APIClient) (string, string, string, string, error) {
				result, err := client.RestoreVersion(
					context.Background(), "org-1", "proj-1", "f-1", "v-1",
				)
				if err != nil {
					return "", "", "", "", err
				}
				return result.ID, result.Current.VersionID, result.Name, result.Current.Name, nil
			},
			wantVersionID:   "v-2",
			wantName:        "report.pdf",
			wantCurrentName: "report.pdf",
		},
		{
			name:        "rename file",
			method:      http.MethodPatch,
			path:        filePath,
			requestName: "renamed.pdf",
			response: `{"success":true,"data":{"id":"f-1","name":"renamed.pdf",` +
				`"current":{"versionId":"v-1","size":10,"name":"report.pdf",` +
				`"contentType":"application/pdf","createdAt":"2026-01-01T00:00:00Z",` +
				`"createdBy":"user-1","isCurrent":true}}}`,
			invoke: func(client *APIClient) (string, string, string, string, error) {
				result, err := client.RenameFile(
					context.Background(), "org-1", "proj-1", "f-1", "renamed.pdf",
				)
				if err != nil {
					return "", "", "", "", err
				}
				return result.ID, result.Current.VersionID, result.Name, result.Current.Name, nil
			},
			wantVersionID:   "v-1",
			wantName:        "renamed.pdf",
			wantCurrentName: "report.pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != tt.method {
						t.Errorf("method = %s, want %s", r.Method, tt.method)
					}
					if r.URL.Path != tt.path {
						t.Errorf("path = %s, want %s", r.URL.Path, tt.path)
					}
					if tt.requestName != "" {
						var body struct {
							Name string `json:"name"`
						}
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
							t.Fatalf("decode request body: %v", err)
						}
						if body.Name != tt.requestName {
							t.Errorf("name = %q, want %q", body.Name, tt.requestName)
						}
					}
					fmt.Fprint(w, tt.response)
				}),
			)
			defer server.Close()

			id, versionID, name, currentName, err := tt.invoke(newEvalTestClient(t, server.URL))
			if err != nil {
				t.Fatalf("mutation error = %v", err)
			}
			if id != "f-1" {
				t.Errorf("id = %q, want f-1", id)
			}
			if versionID != tt.wantVersionID {
				t.Errorf("versionID = %q, want %q", versionID, tt.wantVersionID)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if currentName != tt.wantCurrentName {
				t.Errorf("currentName = %q, want %q", currentName, tt.wantCurrentName)
			}
		})
	}
}

func TestRestoreVersion_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":{"success":false,"error":{"code":"FILE_NOT_FOUND",` +
			`"message":"Version 'v-9' not found"}}}`))
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	_, err := client.RestoreVersion(context.Background(), "org-1", "proj-1", "f-1", "v-9")
	if err == nil || err.Error() != "Version 'v-9' not found" {
		t.Errorf("err = %v, want the server's message", err)
	}
}

func TestFileRefsAreEscapedInPaths(t *testing.T) {
	const base = "/api/platform/v1/organizations/org-1/projects/proj-1/files/"
	cases := []struct {
		name     string
		ref      string
		wantPath string
	}{
		{"space", "Q3 Report.pdf", base + "Q3 Report.pdf"},
		{"percent", "50%.pdf", base + "50%.pdf"},
		{"slash", "reports/q3.pdf", base + "reports/q3.pdf"},
		{"hash", "a#b.pdf", base + "a#b.pdf"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var gotPath, gotRaw string
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					gotPath, gotRaw = r.URL.Path, r.URL.EscapedPath()
					fmt.Fprint(w, `{"success":true,"data":{"id":"f-1","versionId":null}}`)
				}),
			)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			if _, err := client.DeleteFile(
				context.Background(),
				"org-1",
				"proj-1",
				c.ref,
			); err != nil {
				t.Fatalf("DeleteFile(%q) error = %v", c.ref, err)
			}
			if gotPath != c.wantPath {
				t.Errorf("decoded path = %q, want %q (raw %q)", gotPath, c.wantPath, gotRaw)
			}
		})
	}
}

// progressRecorder is mutex-guarded because onProgress fires on the transport's
// body-writing goroutine, not the caller's.
type progressRecorder struct {
	mu    sync.Mutex
	total int64
	peak  int64
}

func (p *progressRecorder) add(n int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total += n
	if p.total > p.peak {
		p.peak = p.total
	}
}

func (p *progressRecorder) snapshot() (total, peak int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.total, p.peak
}

func newUploadServer(
	t *testing.T,
	status int,
	size int64,
	requests *int32,
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(requests, 1)
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			t.Errorf("failed to drain request body for request %d: %v", n, err)
		}
		w.WriteHeader(status)
		if status == http.StatusServiceUnavailable {
			_, _ = w.Write(
				[]byte(
					`{"success":false,"detail":{"error":{"message":"file storage unavailable"}}}`,
				),
			)
			return
		}
		writeFileMetadataResponse(w, "file-1", "v-1", size, "big.bin")
	}))
}

// uploadProgressContent is comfortably over the transport's 4 KiB write buffer,
// so the body is reported through several onProgress deltas.
func uploadProgressContent() []byte {
	return bytes.Repeat([]byte("y"), 8192)
}

func TestCreateFile_ProgressMatchesFileSize(t *testing.T) {
	content := uploadProgressContent()
	localPath := writeTempFile(t, "big.bin", content)
	size := int64(len(content))

	var requests int32
	server := newUploadServer(t, http.StatusCreated, size, &requests)
	defer server.Close()

	var rec progressRecorder
	client := newEvalTestClient(t, server.URL)
	if _, err := client.CreateFile(
		context.Background(), "org-1", "proj-1", localPath, "", rec.add,
	); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}

	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("server saw %d requests, want 1", got)
	}
	total, peak := rec.snapshot()
	if peak > size {
		t.Errorf("progress peaked at %d bytes, want never above the file size %d", peak, size)
	}
	if total != size {
		t.Errorf("progress total = %d, want exactly the file size %d", total, size)
	}
}

type receivedPart struct {
	name        string
	filename    string
	contentType string
	body        []byte
}

// Reads the parts as the backend does: from the part headers, never by scanning the raw body.
func readUploadParts(t *testing.T, r *http.Request) []receivedPart {
	t.Helper()
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("failed to parse request Content-Type %q: %v", r.Header.Get("Content-Type"), err)
	}
	boundary, ok := params["boundary"]
	if !ok {
		t.Fatalf("request Content-Type %q carries no boundary", r.Header.Get("Content-Type"))
	}

	mr := multipart.NewReader(r.Body, boundary)
	var parts []receivedPart
	for {
		p, err := mr.NextRawPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("failed to read next multipart part: %v", err)
		}
		body, err := io.ReadAll(p)
		if err != nil {
			t.Fatalf("failed to read part %q body: %v", p.FormName(), err)
		}
		parts = append(parts, receivedPart{
			name:        p.FormName(),
			filename:    p.FileName(),
			contentType: p.Header.Get("Content-Type"),
			body:        body,
		})
	}
	return parts
}

func findPart(t *testing.T, parts []receivedPart, name string) receivedPart {
	t.Helper()
	for _, p := range parts {
		if p.name == name {
			return p
		}
	}
	t.Fatalf("no %q part in received parts %+v", name, parts)
	return receivedPart{}
}

func TestCreateFile_FilePartContentTypeFromExtension(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{filename: "report.pdf", want: "application/pdf"},
		{filename: "photo.png", want: "image/png"},
		{filename: "notes.txt", want: "text/plain; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			content := []byte("body of " + tt.filename)
			localPath := writeTempFile(t, tt.filename, content)

			var got receivedPart
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					got = findPart(t, readUploadParts(t, r), "file")

					w.WriteHeader(http.StatusCreated)
					writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), tt.filename)
				}),
			)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			if _, err := client.CreateFile(
				context.Background(), "org-1", "proj-1", localPath, "", nil,
			); err != nil {
				t.Fatalf("CreateFile() error = %v", err)
			}

			if got.contentType != tt.want {
				t.Errorf("file part Content-Type = %q, want %q", got.contentType, tt.want)
			}
			if got.filename != tt.filename {
				t.Errorf("file part filename = %q, want %q", got.filename, tt.filename)
			}
			if !bytes.Equal(got.body, content) {
				t.Errorf("file part body = %q, want %q", got.body, content)
			}
		})
	}
}

func TestCreateFile_UnmappedExtensionFallsBackToOctetStream(t *testing.T) {
	// "blob" has no extension and ".bin" is in Go's builtin table, so neither depends on system mime files.
	for _, filename := range []string{"blob", "data.bin"} {
		t.Run(filename, func(t *testing.T) {
			content := []byte("opaque bytes")
			localPath := writeTempFile(t, filename, content)

			var got receivedPart
			server := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					got = findPart(t, readUploadParts(t, r), "file")

					w.WriteHeader(http.StatusCreated)
					writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), filename)
				}),
			)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			if _, err := client.CreateFile(
				context.Background(), "org-1", "proj-1", localPath, "", nil,
			); err != nil {
				t.Fatalf("CreateFile() error = %v", err)
			}

			if got.contentType != "application/octet-stream" {
				t.Errorf(
					"file part Content-Type = %q, want application/octet-stream",
					got.contentType,
				)
			}
		})
	}
}

func TestCreateFile_NamePartPrecedesFilePart(t *testing.T) {
	content := []byte("ordered parts")
	localPath := writeTempFile(t, "local-name.txt", content)

	var gotOrder []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, p := range readUploadParts(t, r) {
			gotOrder = append(gotOrder, p.name)
		}

		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), "custom-name.txt")
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	if _, err := client.CreateFile(
		context.Background(), "org-1", "proj-1", localPath, "custom-name.txt", nil,
	); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}

	want := []string{"name", "file"}
	if !slices.Equal(gotOrder, want) {
		t.Errorf("part order = %v, want %v", gotOrder, want)
	}
}

func TestCreateFile_BackslashFilename(t *testing.T) {
	content := []byte("backslash filename body")
	localPath := writeTempFile(t, `weird\name.txt`, content)

	var got receivedPart
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = findPart(t, readUploadParts(t, r), "file")

		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), got.filename)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	if _, err := client.CreateFile(
		context.Background(), "org-1", "proj-1", localPath, "", nil,
	); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}

	if got.filename != `weird\name.txt` {
		t.Errorf("file part filename = %q, want %q", got.filename, `weird\name.txt`)
	}
	if !bytes.Equal(got.body, content) {
		t.Errorf("file part body = %q, want %q", got.body, content)
	}
}

func TestCreateFile_RefusesControlCharacterFilename(t *testing.T) {
	content := []byte("header injection attempt")
	localPath := writeTempFile(t, "bad\nname.txt", content)

	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), "bad-name.txt")
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	_, err := client.CreateFile(context.Background(), "org-1", "proj-1", localPath, "", nil)
	if err == nil {
		t.Error("CreateFile() error = nil, want a refusal for a filename containing a newline")
	} else if !strings.Contains(err.Error(), "--name") {
		t.Errorf("CreateFile() error = %q, want it to suggest --name", err.Error())
	}
	if got := atomic.LoadInt32(&requests); got != 0 {
		t.Errorf("server received %d requests, want 0", got)
	}
}
