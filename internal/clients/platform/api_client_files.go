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
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

type FileVersion struct {
	VersionID      string    `json:"versionId"`
	Size           int64     `json:"size"`
	Name           string    `json:"name"`
	ContentType    string    `json:"contentType"`
	CreatedAt      time.Time `json:"createdAt"`
	CreatedBy      string    `json:"createdBy"`
	CreatedByEmail *string   `json:"createdByEmail"`
	IsCurrent      bool      `json:"isCurrent"`
}

type FileMetadata struct {
	ID       string        `json:"id"`
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
	Cursor *string `json:"cursor"`
}

type fileListData struct {
	Files []FileMetadata `json:"files"`
	Meta  FileListMeta   `json:"meta"`
}

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

// quoteEscaper matches mime/multipart's escaping for content dispositions.
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"", "\r", "%0D", "\n", "%0A")

func filePartHeader(filename string) (textproto.MIMEHeader, error) {
	if strings.ContainsFunc(filename, func(r rune) bool { return r < 0x20 || r == 0x7f }) {
		return nil, fmt.Errorf(
			"%s contains a control character; use --name to choose a stored name",
			filename,
		)
	}

	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	h := make(textproto.MIMEHeader)
	h.Set(
		"Content-Disposition",
		fmt.Sprintf(`form-data; name="file"; filename="%s"`, quoteEscaper.Replace(filename)),
	)
	h.Set("Content-Type", contentType)
	return h, nil
}

func buildFileUploadFraming(
	name, filename string,
) (header, footer []byte, contentType string, err error) {
	// Keep the file body out of the buffer so it can stream directly from disk.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	if name != "" {
		if err := mw.WriteField("name", name); err != nil {
			return nil, nil, "", fmt.Errorf("failed to write name field: %w", err)
		}
	}
	part, err := filePartHeader(filename)
	if err != nil {
		return nil, nil, "", err
	}
	if _, err := mw.CreatePart(part); err != nil {
		return nil, nil, "", fmt.Errorf("failed to write file field: %w", err)
	}

	contentType = mw.FormDataContentType()
	header = append([]byte(nil), buf.Bytes()...)
	footer = []byte("\r\n--" + mw.Boundary() + "--\r\n")
	return header, footer, contentType, nil
}

func (c *APIClient) CreateFile(
	ctx context.Context,
	orgID, projectID string,
	localPath string,
	name string,
	onProgress func(n int64),
) (*FileMetadata, error) {
	path := evalBasePath(orgID, projectID) + "/files"
	return c.putOrPostFile(
		ctx,
		http.MethodPost,
		path,
		"",
		localPath,
		name,
		onProgress,
		http.StatusCreated,
		"upload file",
	)
}

func (c *APIClient) AddFileVersion(
	ctx context.Context,
	orgID, projectID, ref string,
	localPath string,
	name string,
	onProgress func(n int64),
) (*FileMetadata, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + url.PathEscape(ref)
	return c.putOrPostFile(
		ctx,
		http.MethodPut,
		path,
		ref,
		localPath,
		name,
		onProgress,
		http.StatusOK,
		"add file version",
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

	if resp.StatusCode != wantStatus {
		return nil, fileRefError(ref, respBody, resp.Status)
	}

	data, err := decodeSuccess[FileMetadata](respBody, action)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

type FileRefAmbiguousError struct {
	Ref        string
	Candidates []clients.FileRefCandidate
}

func (e *FileRefAmbiguousError) Error() string {
	return fmt.Sprintf("%q matches more than one file", e.Ref)
}

func fileRefError(ref string, body []byte, status string) error {
	if candidates := clients.ExtractFileRefCandidates(body); len(candidates) > 0 {
		return &FileRefAmbiguousError{Ref: ref, Candidates: candidates}
	}
	if msg := clients.ExtractServerMessage(body); msg != "" {
		return errors.New(msg)
	}
	return fmt.Errorf("server returned %s", status)
}

func (c *APIClient) DownloadFile(
	ctx context.Context,
	orgID, projectID, fileID string,
	dest io.Writer,
	onProgress func(n int64),
	onStart func(total int64),
) (string, int64, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + url.PathEscape(fileID)
	return c.downloadTo(ctx, path, fileID, dest, onProgress, onStart)
}

func (c *APIClient) DownloadFileVersion(
	ctx context.Context,
	orgID, projectID, fileID, versionID string,
	dest io.Writer,
	onProgress func(n int64),
	onStart func(total int64),
) (string, int64, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + url.PathEscape(fileID) +
		"/versions/" + url.PathEscape(versionID)
	return c.downloadTo(ctx, path, fileID, dest, onProgress, onStart)
}

func (c *APIClient) downloadTo(
	ctx context.Context,
	path, ref string,
	dest io.Writer,
	onProgress func(n int64),
	onStart func(total int64),
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

	if onStart != nil && resp.ContentLength >= 0 {
		onStart(resp.ContentLength)
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

func (c *APIClient) GetFileMetadata(
	ctx context.Context,
	orgID, projectID, ref string,
) (*FileMetadata, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + url.PathEscape(ref) + "/metadata"
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
	ID string `json:"id"`
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
	return data.ID, nil
}

// RestoreVersion asks the platform to restore an earlier version's contents.
func (c *APIClient) RestoreVersion(
	ctx context.Context,
	orgID, projectID, ref, versionID string,
) (*FileMetadata, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + url.PathEscape(ref) +
		"/versions/" + url.PathEscape(versionID)
	req, err := c.newRequest(ctx, http.MethodPost, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to restore version: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fileRefError(ref, body, resp.Status)
	}

	data, err := decodeSuccess[FileMetadata](body, "restore version")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

type renameFileRequest struct {
	Name string `json:"name"`
}

func (c *APIClient) RenameFile(
	ctx context.Context,
	orgID, projectID, ref, name string,
) (*FileMetadata, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + url.PathEscape(ref)
	req, err := c.newJSONRequest(ctx, http.MethodPatch, path, renameFileRequest{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to rename file: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fileRefError(ref, body, resp.Status)
	}

	data, err := decodeSuccess[FileMetadata](body, "rename file")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (c *APIClient) DeleteFile(ctx context.Context, orgID, projectID, ref string) (string, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + url.PathEscape(ref)
	return c.deleteFileRef(ctx, path, ref, "delete file")
}

func (c *APIClient) DeleteFileVersion(
	ctx context.Context,
	orgID, projectID, ref, versionID string,
) (string, error) {
	path := evalBasePath(orgID, projectID) + "/files/" + url.PathEscape(ref) +
		"/versions/" + url.PathEscape(versionID)
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
