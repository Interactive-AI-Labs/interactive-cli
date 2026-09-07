package platform

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
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
	if result.Id != "file-1" {
		t.Errorf("result.Id = %q, want file-1", result.Id)
	}

	if len(gotTransferEncoding) != 0 {
		t.Errorf("TransferEncoding = %v, want empty (declared length, not chunked)", gotTransferEncoding)
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

func TestCreateFile_RetriesOn503(t *testing.T) {
	content := bytes.Repeat([]byte("x"), 4096)
	localPath := writeTempFile(t, "big.bin", content)

	origDelay := uploadRetryDelay
	defer func() { uploadRetryDelay = origDelay }()
	uploadRetryDelay = func(time.Duration) {}

	var attempt int32
	var firstContentLength, secondContentLength int64
	var firstBody, secondBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempt, 1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		if n == 1 {
			firstContentLength = r.ContentLength
			firstBody = body
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"success":false,"detail":{"error":{"message":"file storage unavailable"}}}`))
			return
		}

		secondContentLength = r.ContentLength
		secondBody = body
		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), "big.bin")
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	result, err := client.CreateFile(context.Background(), "org-1", "proj-1", localPath, "", nil)
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	if result.Id != "file-1" {
		t.Errorf("result.Id = %q, want file-1", result.Id)
	}

	if atomic.LoadInt32(&attempt) != 2 {
		t.Fatalf("server saw %d attempts, want 2", attempt)
	}
	if firstContentLength != secondContentLength {
		t.Errorf(
			"declared length changed across retry: first=%d second=%d",
			firstContentLength, secondContentLength,
		)
	}
	if !bytes.Equal(firstBody, secondBody) {
		t.Errorf("retry resent a different body than the first attempt")
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

func TestCreateFile_SHA256Identity(t *testing.T) {
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
			t.Errorf("server-observed sha256 = %x, want %x", gotSum, wantSum)
		}

		w.WriteHeader(http.StatusCreated)
		writeFileMetadataResponse(w, "file-1", "v-1", int64(len(content)), "data.bin")
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	if _, err := client.CreateFile(context.Background(), "org-1", "proj-1", localPath, "", nil); err != nil {
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
			if _, _, _, err := client.ListFiles(context.Background(), "org-1", "proj-1", tt.opts); err != nil {
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
	files, meta, _, err := client.ListFiles(context.Background(), "org-1", "proj-1", FileListOptions{})
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("len(files) = %d, want 2", len(files))
	}
	if files[0].Name != "report.pdf" {
		t.Errorf("files[0].Name = %q, want report.pdf", files[0].Name)
	}
	if files[0].Current.CreatedByEmail == nil || *files[0].Current.CreatedByEmail != "a@example.com" {
		t.Errorf("files[0].Current.CreatedByEmail = %v, want a@example.com", files[0].Current.CreatedByEmail)
	}
	if files[1].Current.CreatedByEmail != nil {
		t.Errorf("files[1].Current.CreatedByEmail = %v, want nil", *files[1].Current.CreatedByEmail)
	}
	if meta.Cursor == nil || *meta.Cursor != "next-page" {
		t.Errorf("meta.Cursor = %v, want next-page", meta.Cursor)
	}
	if meta.Limit != 50 {
		t.Errorf("meta.Limit = %d, want 50", meta.Limit)
	}
}

func writeFileMetadataResponse(w http.ResponseWriter, id, versionID string, size int64, name string) {
	resp := struct {
		Success bool         `json:"success"`
		Data    FileMetadata `json:"data"`
	}{
		Success: true,
		Data: FileMetadata{
			Id: id,
			Current: FileVersion{
				VersionId:   versionID,
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
		w.Header().Set("Content-Disposition", `attachment; filename="evil.txt"; filename*=UTF-8''%2E%2E`)
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
		func(n int64) { gotProgress += n },
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
		w.Header().Set("Content-Disposition", `attachment; filename="report.txt"; filename*=UTF-8''report.txt`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	var buf bytes.Buffer
	filename, _, err := client.DownloadFileVersion(
		context.Background(), "org-1", "proj-1", "file-1", "v-0", &buf, nil,
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

func TestDownloadFile_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":{"error":{"message":"file not found"}}}`))
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	var buf bytes.Buffer
	_, _, err := client.DownloadFile(context.Background(), "org-1", "proj-1", "missing", &buf, nil)
	if err == nil {
		t.Fatal("DownloadFile() error = nil, want a not-found error")
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
	_, _, err := client.DownloadFile(context.Background(), "org-1", "proj-1", "report.pdf", &buf, nil)

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
		ids[v.VersionId] = true
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

func TestGetFileMetadata_AmbiguousRef(t *testing.T) {
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
	_, _, err := client.GetFileMetadata(context.Background(), "org-1", "proj-1", "report.pdf")

	var ambiguous *FileRefAmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("GetFileMetadata() error = %v, want a *FileRefAmbiguousError", err)
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
		ids[v.VersionId] = true
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
	result, err := client.AddFileVersion(context.Background(), "org-1", "proj-1", "f-1", localPath, "", nil)
	if err != nil {
		t.Fatalf("AddFileVersion() error = %v", err)
	}
	if result.Id != "f-1" {
		t.Errorf("result.Id = %q, want f-1", result.Id)
	}
	if gotHasNameKey {
		t.Error("multipart form carries a name part, want none")
	}
	if gotContentLength <= 0 || gotContentLength != int64(gotBodyLen) {
		t.Errorf("ContentLength = %d, want it to equal the actual body length %d", gotContentLength, gotBodyLen)
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
		t.Errorf("ContentLength = %d, want it to equal the actual body length %d", gotContentLength, gotBodyLen)
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
	_, err := client.AddFileVersion(context.Background(), "org-1", "proj-1", "missing", localPath, "", nil)
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
	_, err := client.AddFileVersion(context.Background(), "org-1", "proj-1", "report.pdf", localPath, "", nil)

	var ambiguous *FileRefAmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("AddFileVersion() error = %v, want a *FileRefAmbiguousError", err)
	}
	if ambiguous.Ref != "report.pdf" {
		t.Errorf("Ref = %q, want report.pdf", ambiguous.Ref)
	}
}

func TestDeleteFile_HitsWholeFilePath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/files/f-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"success":true,"data":{"id":"f-1","versionId":null}}`)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	id, err := client.DeleteFile(context.Background(), "org-1", "proj-1", "f-1")
	if err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}
	if id != "f-1" {
		t.Errorf("id = %q, want f-1", id)
	}
}

func TestDeleteFileVersion_HitsVersionPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/platform/v1/organizations/org-1/projects/proj-1/files/f-1/versions/v-1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"success":true,"data":{"id":"f-1","versionId":"v-1"}}`)
	}))
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	id, err := client.DeleteFileVersion(context.Background(), "org-1", "proj-1", "f-1", "v-1")
	if err != nil {
		t.Fatalf("DeleteFileVersion() error = %v", err)
	}
	if id != "f-1" {
		t.Errorf("id = %q, want f-1", id)
	}
}

func TestDeleteFile_AmbiguousRef(t *testing.T) {
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
	_, err := client.DeleteFile(context.Background(), "org-1", "proj-1", "report.pdf")

	var ambiguous *FileRefAmbiguousError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("DeleteFile() error = %v, want a *FileRefAmbiguousError", err)
	}
}
