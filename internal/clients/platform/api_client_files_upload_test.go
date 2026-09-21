package platform

import (
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestFileUpload_UnavailableSendsOneRequestAndReportsServerMessage(t *testing.T) {
	content := uploadProgressContent()
	localPath := writeTempFile(t, "big.bin", content)
	size := int64(len(content))

	tests := []struct {
		name   string
		upload func(*APIClient) (*FileMetadata, error)
	}{
		{
			name: "create",
			upload: func(client *APIClient) (*FileMetadata, error) {
				return client.CreateFile(
					context.Background(), "org-1", "proj-1", localPath, "", nil,
				)
			},
		},
		{
			name: "update",
			upload: func(client *APIClient) (*FileMetadata, error) {
				return client.AddFileVersion(
					context.Background(), "org-1", "proj-1", "f-1", localPath, "", nil,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requests int32
			server := newUploadServer(t, http.StatusServiceUnavailable, size, &requests)
			defer server.Close()

			client := newEvalTestClient(t, server.URL)
			_, err := tt.upload(client)
			if err == nil {
				t.Fatal("upload error = nil, want an error when the store is unavailable")
			}

			if got := atomic.LoadInt32(&requests); got != 1 {
				t.Errorf("server saw %d requests, want exactly 1", got)
			}
			if !strings.Contains(err.Error(), "file storage unavailable") {
				t.Errorf(
					"upload error = %q, want it to contain the server message %q",
					err.Error(),
					"file storage unavailable",
				)
			}
			if strings.Contains(err.Error(), "attempt") {
				t.Errorf("upload error = %q, want no retry language", err.Error())
			}
		})
	}
}

func TestCreateFile_SuccessSendsOneRequestAndDecodesMetadata(t *testing.T) {
	content := uploadProgressContent()
	localPath := writeTempFile(t, "big.bin", content)
	size := int64(len(content))

	var requests int32
	server := newUploadServer(t, http.StatusCreated, size, &requests)
	defer server.Close()

	client := newEvalTestClient(t, server.URL)
	result, err := client.CreateFile(context.Background(), "org-1", "proj-1", localPath, "", nil)
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}

	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Errorf("server saw %d requests, want exactly 1", got)
	}
	if result.ID != "file-1" {
		t.Errorf("result.ID = %q, want file-1", result.ID)
	}
	if result.Current.VersionID != "v-1" {
		t.Errorf("result.Current.VersionID = %q, want v-1", result.Current.VersionID)
	}
	if result.Current.Size != size {
		t.Errorf("result.Current.Size = %d, want %d", result.Current.Size, size)
	}
}
