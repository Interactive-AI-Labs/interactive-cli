package platform

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
