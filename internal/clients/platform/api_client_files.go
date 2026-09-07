package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// FileVersion is one version of a stored file, as reported by the files API.
type FileVersion struct {
	VersionId      string    `json:"versionId"`
	Size           int64     `json:"size"`
	Name           string    `json:"name"`
	ContentType    string    `json:"contentType"`
	CreatedAt      time.Time `json:"createdAt"`
	CreatedBy      string    `json:"createdBy"`
	CreatedByEmail *string   `json:"createdByEmail"`
	IsCurrent      bool      `json:"isCurrent"`
}

// FileMetadata is a stored file and its current version, as reported by the files API.
type FileMetadata struct {
	Id       string        `json:"id"`
	Name     string        `json:"name"`
	Current  FileVersion   `json:"current"`
	Versions []FileVersion `json:"versions,omitempty"`
}

type FileListOptions struct {
	Limit           *int   `url:"limit,omitempty"`
	Cursor          string `url:"cursor,omitempty"`
	IncludeVersions bool   `url:"includeVersions,omitempty"`
}

type FileListMeta struct {
	Cursor     *string `json:"cursor"`
	TotalItems *int    `json:"total_items"`
	Limit      int     `json:"limit"`
}

type fileListData struct {
	Files []FileMetadata `json:"files"`
	Meta  FileListMeta   `json:"meta"`
}

// ListFiles retrieves one page of a project's files.
func (c *APIClient) ListFiles(
	ctx context.Context,
	orgID, projectID string,
	opts FileListOptions,
) ([]FileMetadata, FileListMeta, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/files"
	data, raw, err := doList[fileListData](c, ctx, path, opts, "list files")
	if err != nil {
		return nil, FileListMeta{}, nil, err
	}
	return data.Files, data.Meta, raw, nil
}

// filesUploadMaxAttempts bounds retries against a persistently unavailable store.
const filesUploadMaxAttempts = 4

// uploadRetryDelay sleeps between upload retries; overridable in tests.
var uploadRetryDelay = time.Sleep

// countingReader reports every successful Read to onRead.
type countingReader struct {
	r      io.Reader
	onRead func(n int64)
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 && c.onRead != nil {
		c.onRead(int64(n))
	}
	return n, err
}

// buildFileUploadFraming builds the multipart header/footer around an omitted file body, so it can stream straight from disk.
func buildFileUploadFraming(name, filename string) (header, footer []byte, contentType string, err error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	if name != "" {
		if err := mw.WriteField("name", name); err != nil {
			return nil, nil, "", fmt.Errorf("failed to write name field: %w", err)
		}
	}
	if _, err := mw.CreateFormFile("file", filename); err != nil {
		return nil, nil, "", fmt.Errorf("failed to write file field: %w", err)
	}

	contentType = mw.FormDataContentType()
	header = append([]byte(nil), buf.Bytes()...)
	footer = []byte("\r\n--" + mw.Boundary() + "--\r\n")
	return header, footer, contentType, nil
}

// CreateFile uploads localPath as a new file, using name as the stored name if set and reporting bytes read via onProgress.
func (c *APIClient) CreateFile(
	ctx context.Context,
	orgID, projectID string,
	localPath string,
	name string,
	onProgress func(n int64),
) (*FileMetadata, error) {
	path := evalBasePath(orgID, projectID) + "/files"
	return c.putOrPostFile(
		ctx, http.MethodPost, path, "", localPath, name, onProgress, http.StatusCreated, "upload file",
	)
}

// AddFileVersion uploads localPath as a new version of ref, keeping its stored name unless name is given.
func (c *APIClient) AddFileVersion(
	ctx context.Context,
	orgID, projectID, ref string,
	localPath string,
	name string,
	onProgress func(n int64),
) (*FileMetadata, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + ref
	return c.putOrPostFile(
		ctx, http.MethodPut, path, ref, localPath, name, onProgress, http.StatusOK, "add file version",
	)
}

func (c *APIClient) putOrPostFile(
	ctx context.Context,
	method, path, ref string,
	localPath string,
	name string,
	onProgress func(n int64),
	wantStatus int,
	action string,
) (*FileMetadata, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", localPath, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat %s: %w", localPath, err)
	}
	size := info.Size()

	header, footer, contentType, err := buildFileUploadFraming(name, filepath.Base(localPath))
	if err != nil {
		return nil, err
	}
	contentLength := int64(len(header)) + size + int64(len(footer))

	var lastErr error
	for attempt := 1; attempt <= filesUploadMaxAttempts; attempt++ {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to rewind %s: %w", localPath, err)
		}

		body := io.MultiReader(
			bytes.NewReader(header),
			&countingReader{r: f, onRead: onProgress},
			bytes.NewReader(footer),
		)

		req, err := c.newRequest(ctx, method, path)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Body = io.NopCloser(body)
		req.ContentLength = contentLength
		req.Header.Set("Content-Type", contentType)

		resp, err := c.do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to %s: %w", action, err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read %s response: %w", action, err)
		}

		if resp.StatusCode == http.StatusServiceUnavailable && attempt < filesUploadMaxAttempts {
			lastErr = errors.New("file storage unavailable")
			if msg := clients.ExtractServerMessage(respBody); msg != "" {
				lastErr = errors.New(msg)
			}
			uploadRetryDelay(retryAfterDelay(resp.Header.Get("Retry-After")))
			continue
		}

		if resp.StatusCode != wantStatus {
			return nil, fileRefError(ref, respBody, resp.Status)
		}

		data, err := decodeSuccess[FileMetadata](respBody, action)
		if err != nil {
			return nil, err
		}
		return &data, nil
	}

	return nil, fmt.Errorf("failed to %s after %d attempts: %w", action, filesUploadMaxAttempts, lastErr)
}

// retryAfterDelay parses Retry-After (seconds), or falls back to a short delay.
func retryAfterDelay(header string) time.Duration {
	if header == "" {
		return time.Second
	}
	seconds, err := strconv.Atoi(header)
	if err != nil || seconds < 0 {
		return time.Second
	}
	return time.Duration(seconds) * time.Second
}

// FileRefAmbiguousError is returned when a name matches more than one file.
type FileRefAmbiguousError struct {
	Ref        string
	Candidates []clients.FileRefCandidate
}

func (e *FileRefAmbiguousError) Error() string {
	return fmt.Sprintf("%q matches more than one file", e.Ref)
}

// fileRefError builds the error for a non-2xx response on a ref-addressed file route.
func fileRefError(ref string, body []byte, status string) error {
	if candidates := clients.ExtractFileRefCandidates(body); len(candidates) > 0 {
		return &FileRefAmbiguousError{Ref: ref, Candidates: candidates}
	}
	if msg := clients.ExtractServerMessage(body); msg != "" {
		return errors.New(msg)
	}
	return fmt.Errorf("server returned %s", status)
}

// DownloadFile streams a file's current version into dest, reporting progress via onProgress.
func (c *APIClient) DownloadFile(
	ctx context.Context,
	orgID, projectID, fileID string,
	dest io.Writer,
	onProgress func(n int64),
) (string, int64, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + fileID
	return c.downloadTo(ctx, path, fileID, dest, onProgress)
}

// DownloadFileVersion streams one specific version's bytes into dest.
func (c *APIClient) DownloadFileVersion(
	ctx context.Context,
	orgID, projectID, fileID, versionID string,
	dest io.Writer,
	onProgress func(n int64),
) (string, int64, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + fileID + "/versions/" + versionID
	return c.downloadTo(ctx, path, fileID, dest, onProgress)
}

func (c *APIClient) downloadTo(
	ctx context.Context,
	path, ref string,
	dest io.Writer,
	onProgress func(n int64),
) (string, int64, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fileRefError(ref, body, resp.Status)
	}

	filename := parseDispositionFilename(resp.Header.Get("Content-Disposition"))

	src := io.Reader(resp.Body)
	if onProgress != nil {
		src = &countingReader{r: resp.Body, onRead: onProgress}
	}
	written, err := io.Copy(dest, src)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read download body: %w", err)
	}

	return filename, written, nil
}

type fileDetailData struct {
	File FileMetadata `json:"file"`
}

// GetFileMetadata reads a file's current version and full version history.
func (c *APIClient) GetFileMetadata(
	ctx context.Context,
	orgID, projectID, ref string,
) (*FileMetadata, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + ref + "/metadata"
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fileRefError(ref, body, resp.Status)
	}

	data, err := decodeSuccess[fileDetailData](body, "get file metadata")
	if err != nil {
		return nil, nil, err
	}
	return &data.File, json.RawMessage(body), nil
}

type fileDeleteData struct {
	Id        string  `json:"id"`
	VersionId *string `json:"versionId"`
}

func (c *APIClient) deleteFileRef(ctx context.Context, path, ref, action string) (string, error) {
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("failed to %s: %w", action, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fileRefError(ref, body, resp.Status)
	}

	data, err := decodeSuccess[fileDeleteData](body, action)
	if err != nil {
		return "", err
	}
	return data.Id, nil
}

// DeleteFile deletes a file and all its versions.
func (c *APIClient) DeleteFile(ctx context.Context, orgID, projectID, ref string) (string, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + ref
	return c.deleteFileRef(ctx, path, ref, "delete file")
}

// DeleteFileVersion deletes one superseded version of a file.
func (c *APIClient) DeleteFileVersion(
	ctx context.Context,
	orgID, projectID, ref, versionID string,
) (string, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + ref + "/versions/" + versionID
	return c.deleteFileRef(ctx, path, ref, "delete file version")
}

func parseDispositionFilename(header string) string {
	if header == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(header)
	if err != nil {
		return ""
	}
	return params["filename"]
}
