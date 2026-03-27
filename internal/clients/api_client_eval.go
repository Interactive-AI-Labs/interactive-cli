package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

// doAndRead executes a request and returns the raw response body.
// It handles status-code checking and server-message extraction.
func (c *APIClient) doAndRead(req *http.Request, action string) ([]byte, error) {
	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to %s: %w", action, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(body); msg != "" {
			return nil, errors.New(msg)
		}
		return nil, fmt.Errorf("failed to %s: server returned %s", action, resp.Status)
	}

	return body, nil
}

// successEnvelope is the common response wrapper used by the platform API.
type successEnvelope[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

// decodeSuccess unmarshals a response body into {success, data} and checks success.
func decodeSuccess[T any](body []byte, action string) (T, error) {
	var envelope successEnvelope[T]
	if err := json.Unmarshal(body, &envelope); err != nil {
		var zero T
		return zero, fmt.Errorf("failed to decode %s response: %w", action, err)
	}
	if !envelope.Success {
		var zero T
		if msg := ExtractServerMessage(body); msg != "" {
			return zero, errors.New(msg)
		}
		return zero, fmt.Errorf("%s returned success=false", action)
	}
	return envelope.Data, nil
}

// doList is a convenience for list endpoints: builds query string, executes, decodes.
func doList[T any](
	c *APIClient,
	ctx context.Context,
	path string,
	opts any,
	action string,
) (T, json.RawMessage, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		var zero T
		return zero, nil, fmt.Errorf("failed to create request: %w", err)
	}

	q, err := query.Values(opts)
	if err != nil {
		var zero T
		return zero, nil, fmt.Errorf("failed to encode query parameters: %w", err)
	}
	req.URL.RawQuery = q.Encode()

	body, err := c.doAndRead(req, action)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	data, err := decodeSuccess[T](body, action)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	return data, json.RawMessage(body), nil
}

// doGet is a convenience for single-resource GET endpoints.
func doGet[T any](
	c *APIClient,
	ctx context.Context,
	path, action string,
) (T, json.RawMessage, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		var zero T
		return zero, nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := c.doAndRead(req, action)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	data, err := decodeSuccess[T](body, action)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	return data, json.RawMessage(body), nil
}

// doCreate is a convenience for POST endpoints that return a created resource.
func doCreate[T any](
	c *APIClient,
	ctx context.Context,
	path string,
	reqBody any,
	action string,
) (T, json.RawMessage, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost, path, reqBody)
	if err != nil {
		var zero T
		return zero, nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := c.doAndRead(req, action)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	data, err := decodeSuccess[T](body, action)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	return data, json.RawMessage(body), nil
}

// checkSuccess verifies the success field in a response body and extracts
// the message. Returns the message on success, or an error if success=false.
func checkSuccess(body []byte, action string) (string, error) {
	var envelope struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", fmt.Errorf("%s: failed to parse response: %w", action, err)
	}
	if !envelope.Success {
		if msg := ExtractServerMessage(body); msg != "" {
			return "", errors.New(msg)
		}
		return "", fmt.Errorf("%s returned success=false", action)
	}
	if msg := ExtractServerMessage(body); msg != "" {
		return msg, nil
	}
	return "", nil
}

// doDelete is a convenience for DELETE endpoints that return a message.
func (c *APIClient) doDelete(
	ctx context.Context,
	path, action string,
) (string, error) {
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	body, err := c.doAndRead(req, action)
	if err != nil {
		return "", err
	}

	return checkSuccess(body, action)
}

// doUpdate is a convenience for PATCH endpoints that return an updated resource.
func doUpdate[T any](
	c *APIClient,
	ctx context.Context,
	path string,
	reqBody any,
	action string,
) (T, json.RawMessage, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPatch, path, reqBody)
	if err != nil {
		var zero T
		return zero, nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := c.doAndRead(req, action)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	data, err := decodeSuccess[T](body, action)
	if err != nil {
		var zero T
		return zero, nil, err
	}

	return data, json.RawMessage(body), nil
}

func evalBasePath(orgID, projectID string) string {
	return fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
	)
}

// ---------------------------------------------------------------------------
// Datasets
// ---------------------------------------------------------------------------

type DatasetInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type DatasetListOptions struct {
	Page  int `url:"page,omitempty"`
	Limit int `url:"limit,omitempty"`
}

type DatasetCreateBody struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Metadata    any    `json:"metadata,omitempty"`
}

type datasetListData struct {
	Datasets []DatasetInfo `json:"datasets"`
	Meta     PageMeta      `json:"meta"`
}

type datasetWrapper struct {
	Dataset DatasetInfo `json:"dataset"`
}

func (c *APIClient) ListDatasets(
	ctx context.Context,
	orgID, projectID string,
	opts DatasetListOptions,
) ([]DatasetInfo, PageMeta, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/datasets"
	data, raw, err := doList[datasetListData](c, ctx, path, opts, "list datasets")
	if err != nil {
		return nil, PageMeta{}, nil, err
	}
	return data.Datasets, data.Meta, raw, nil
}

func (c *APIClient) GetDataset(
	ctx context.Context,
	orgID, projectID, name string,
) (*DatasetInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/datasets/" + url.PathEscape(name)
	data, raw, err := doGet[datasetWrapper](c, ctx, path, "get dataset")
	if err != nil {
		return nil, nil, err
	}
	return &data.Dataset, raw, nil
}

func (c *APIClient) CreateDataset(
	ctx context.Context,
	orgID, projectID string,
	body DatasetCreateBody,
) (*DatasetInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/datasets"
	data, raw, err := doCreate[datasetWrapper](c, ctx, path, body, "create dataset")
	if err != nil {
		return nil, nil, err
	}
	return &data.Dataset, raw, nil
}

// ---------------------------------------------------------------------------
// Dataset Items
// ---------------------------------------------------------------------------

type DatasetItemInfo struct {
	ID                  string          `json:"id"`
	Status              string          `json:"status"`
	DatasetID           string          `json:"dataset_id"`
	DatasetName         string          `json:"dataset_name"`
	Input               json.RawMessage `json:"input"`
	ExpectedOutput      json.RawMessage `json:"expected_output"`
	Metadata            json.RawMessage `json:"metadata"`
	SourceTraceID       string          `json:"source_trace_id"`
	SourceObservationID string          `json:"source_observation_id"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
}

type DatasetItemListOptions struct {
	DatasetName         string `url:"dataset_name,omitempty"`
	SourceTraceID       string `url:"source_trace_id,omitempty"`
	SourceObservationID string `url:"source_observation_id,omitempty"`
	Page                int    `url:"page,omitempty"`
	Limit               int    `url:"limit,omitempty"`
}

type DatasetItemCreateBody struct {
	ID                  string `json:"id,omitempty"`
	DatasetName         string `json:"dataset_name"`
	Input               any    `json:"input,omitempty"`
	ExpectedOutput      any    `json:"expected_output,omitempty"`
	Metadata            any    `json:"metadata,omitempty"`
	SourceTraceID       string `json:"source_trace_id,omitempty"`
	SourceObservationID string `json:"source_observation_id,omitempty"`
	Status              string `json:"status,omitempty"`
}

type datasetItemListData struct {
	Items []DatasetItemInfo `json:"items"`
	Meta  PageMeta          `json:"meta"`
}

type datasetItemWrapper struct {
	Item DatasetItemInfo `json:"item"`
}

func (c *APIClient) ListDatasetItems(
	ctx context.Context,
	orgID, projectID string,
	opts DatasetItemListOptions,
) ([]DatasetItemInfo, PageMeta, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/dataset-items"
	data, raw, err := doList[datasetItemListData](c, ctx, path, opts, "list dataset items")
	if err != nil {
		return nil, PageMeta{}, nil, err
	}
	return data.Items, data.Meta, raw, nil
}

func (c *APIClient) GetDatasetItem(
	ctx context.Context,
	orgID, projectID, itemID string,
) (*DatasetItemInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/dataset-items/" + url.PathEscape(itemID)
	data, raw, err := doGet[datasetItemWrapper](c, ctx, path, "get dataset item")
	if err != nil {
		return nil, nil, err
	}
	return &data.Item, raw, nil
}

func (c *APIClient) CreateDatasetItem(
	ctx context.Context,
	orgID, projectID string,
	body DatasetItemCreateBody,
) (*DatasetItemInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/dataset-items"
	data, raw, err := doCreate[datasetItemWrapper](c, ctx, path, body, "create dataset item")
	if err != nil {
		return nil, nil, err
	}
	return &data.Item, raw, nil
}

func (c *APIClient) DeleteDatasetItem(
	ctx context.Context,
	orgID, projectID, itemID string,
) (string, error) {
	path := evalBasePath(orgID, projectID) + "/dataset-items/" + url.PathEscape(itemID)
	return c.doDelete(ctx, path, "delete dataset item")
}

// ---------------------------------------------------------------------------
// Dataset Runs
// ---------------------------------------------------------------------------

type DatasetRunInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	DatasetID   string          `json:"dataset_id"`
	DatasetName string          `json:"dataset_name"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type DatasetRunListOptions struct {
	Page  int `url:"page,omitempty"`
	Limit int `url:"limit,omitempty"`
}

type datasetRunListData struct {
	Runs []DatasetRunInfo `json:"runs"`
	Meta PageMeta         `json:"meta"`
}

type datasetRunWrapper struct {
	Run DatasetRunInfo `json:"run"`
}

func (c *APIClient) ListDatasetRuns(
	ctx context.Context,
	orgID, projectID, datasetName string,
	opts DatasetRunListOptions,
) ([]DatasetRunInfo, PageMeta, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/datasets/" +
		url.PathEscape(datasetName) + "/runs"
	data, raw, err := doList[datasetRunListData](c, ctx, path, opts, "list dataset runs")
	if err != nil {
		return nil, PageMeta{}, nil, err
	}
	return data.Runs, data.Meta, raw, nil
}

func (c *APIClient) GetDatasetRun(
	ctx context.Context,
	orgID, projectID, datasetName, runName string,
) (*DatasetRunInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/datasets/" +
		url.PathEscape(datasetName) + "/runs/" + url.PathEscape(runName)
	data, raw, err := doGet[datasetRunWrapper](c, ctx, path, "get dataset run")
	if err != nil {
		return nil, nil, err
	}
	return &data.Run, raw, nil
}

func (c *APIClient) DeleteDatasetRun(
	ctx context.Context,
	orgID, projectID, datasetName, runName string,
) (string, error) {
	path := evalBasePath(orgID, projectID) + "/datasets/" +
		url.PathEscape(datasetName) + "/runs/" + url.PathEscape(runName)
	return c.doDelete(ctx, path, "delete dataset run")
}

// ---------------------------------------------------------------------------
// Dataset Run Items
// ---------------------------------------------------------------------------

type DatasetRunItemInfo struct {
	ID             string          `json:"id"`
	DatasetRunID   string          `json:"dataset_run_id"`
	DatasetRunName string          `json:"dataset_run_name"`
	DatasetItemID  string          `json:"dataset_item_id"`
	TraceID        string          `json:"trace_id"`
	ObservationID  string          `json:"observation_id"`
	Metadata       json.RawMessage `json:"metadata"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

type DatasetRunItemListOptions struct {
	RunName     string `url:"run_name,omitempty"`
	DatasetName string `url:"dataset_name,omitempty"`
	Page        int    `url:"page,omitempty"`
	Limit       int    `url:"limit,omitempty"`
}

type DatasetRunItemCreateBody struct {
	RunName        string `json:"run_name"`
	RunDescription string `json:"run_description,omitempty"`
	DatasetItemID  string `json:"dataset_item_id"`
	TraceID        string `json:"trace_id,omitempty"`
	ObservationID  string `json:"observation_id,omitempty"`
	Metadata       any    `json:"metadata,omitempty"`
}

type datasetRunItemListData struct {
	Items []DatasetRunItemInfo `json:"run_items"`
	Meta  PageMeta             `json:"meta"`
}

type datasetRunItemWrapper struct {
	RunItem DatasetRunItemInfo `json:"run_item"`
}

func (c *APIClient) ListDatasetRunItems(
	ctx context.Context,
	orgID, projectID string,
	opts DatasetRunItemListOptions,
) ([]DatasetRunItemInfo, PageMeta, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/dataset-run-items"
	data, raw, err := doList[datasetRunItemListData](
		c, ctx, path, opts, "list dataset run items",
	)
	if err != nil {
		return nil, PageMeta{}, nil, err
	}
	return data.Items, data.Meta, raw, nil
}

func (c *APIClient) CreateDatasetRunItem(
	ctx context.Context,
	orgID, projectID string,
	body DatasetRunItemCreateBody,
) (*DatasetRunItemInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/dataset-run-items"
	data, raw, err := doCreate[datasetRunItemWrapper](
		c, ctx, path, body, "create dataset run item",
	)
	if err != nil {
		return nil, nil, err
	}
	return &data.RunItem, raw, nil
}

// ---------------------------------------------------------------------------
// Annotation Queues
// ---------------------------------------------------------------------------

type AnnotationQueueInfo struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	ScoreConfigIDs []string `json:"score_config_ids"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
}

type AnnotationQueueListOptions struct {
	Page  int `url:"page,omitempty"`
	Limit int `url:"limit,omitempty"`
}

type AnnotationQueueCreateBody struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	ScoreConfigIDs []string `json:"score_config_ids,omitempty"`
}

type annotationQueueListData struct {
	Queues []AnnotationQueueInfo `json:"queues"`
	Meta   PageMeta              `json:"meta"`
}

type annotationQueueWrapper struct {
	Queue AnnotationQueueInfo `json:"queue"`
}

func (c *APIClient) ListAnnotationQueues(
	ctx context.Context,
	orgID, projectID string,
	opts AnnotationQueueListOptions,
) ([]AnnotationQueueInfo, PageMeta, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/annotation-queues"
	data, raw, err := doList[annotationQueueListData](
		c, ctx, path, opts, "list annotation queues",
	)
	if err != nil {
		return nil, PageMeta{}, nil, err
	}
	return data.Queues, data.Meta, raw, nil
}

func (c *APIClient) GetAnnotationQueue(
	ctx context.Context,
	orgID, projectID, queueID string,
) (*AnnotationQueueInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/annotation-queues/" +
		url.PathEscape(queueID)
	data, raw, err := doGet[annotationQueueWrapper](c, ctx, path, "get annotation queue")
	if err != nil {
		return nil, nil, err
	}
	return &data.Queue, raw, nil
}

func (c *APIClient) CreateAnnotationQueue(
	ctx context.Context,
	orgID, projectID string,
	body AnnotationQueueCreateBody,
) (*AnnotationQueueInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/annotation-queues"
	data, raw, err := doCreate[annotationQueueWrapper](
		c, ctx, path, body, "create annotation queue",
	)
	if err != nil {
		return nil, nil, err
	}
	return &data.Queue, raw, nil
}

// AssignQueue assigns a user to an annotation queue.
func (c *APIClient) AssignQueue(
	ctx context.Context,
	orgID, projectID, queueID, userID string,
) (string, error) {
	path := evalBasePath(orgID, projectID) + "/annotation-queues/" +
		url.PathEscape(queueID) + "/assignments"
	body := map[string]string{"user_id": userID}
	req, err := c.newJSONRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	respBody, err := c.doAndRead(req, "assign queue")
	if err != nil {
		return "", err
	}

	msg, err := checkSuccess(respBody, "assign queue")
	if err != nil {
		return "", err
	}
	if msg != "" {
		return msg, nil
	}

	return "User assigned to queue successfully.", nil
}

// UnassignQueue removes a user from an annotation queue.
func (c *APIClient) UnassignQueue(
	ctx context.Context,
	orgID, projectID, queueID, userID string,
) (string, error) {
	path := evalBasePath(orgID, projectID) + "/annotation-queues/" +
		url.PathEscape(queueID) + "/assignments"
	body := map[string]string{"user_id": userID}
	req, err := c.newJSONRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	respBody, err := c.doAndRead(req, "unassign queue")
	if err != nil {
		return "", err
	}

	msg, err := checkSuccess(respBody, "unassign queue")
	if err != nil {
		return "", err
	}
	if msg != "" {
		return msg, nil
	}

	return "User unassigned from queue successfully.", nil
}

// ---------------------------------------------------------------------------
// Queue Items
// ---------------------------------------------------------------------------

type QueueItemInfo struct {
	ID          string          `json:"id"`
	QueueID     string          `json:"queue_id"`
	ObjectID    string          `json:"object_id"`
	ObjectType  string          `json:"object_type"`
	Status      string          `json:"status"`
	Metadata    json.RawMessage `json:"metadata"`
	CompletedAt string          `json:"completed_at"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type QueueItemListOptions struct {
	Status string `url:"status,omitempty"`
	Page   int    `url:"page,omitempty"`
	Limit  int    `url:"limit,omitempty"`
}

type QueueItemCreateBody struct {
	ObjectID   string `json:"object_id"`
	ObjectType string `json:"object_type"`
	Status     string `json:"status,omitempty"`
}

type QueueItemUpdateBody struct {
	Status string `json:"status"`
}

type queueItemListData struct {
	Items []QueueItemInfo `json:"items"`
	Meta  PageMeta        `json:"meta"`
}

type queueItemWrapper struct {
	Item QueueItemInfo `json:"item"`
}

func queueItemsPath(orgID, projectID, queueID string) string {
	return evalBasePath(orgID, projectID) + "/annotation-queues/" +
		url.PathEscape(queueID) + "/items"
}

func (c *APIClient) ListQueueItems(
	ctx context.Context,
	orgID, projectID, queueID string,
	opts QueueItemListOptions,
) ([]QueueItemInfo, PageMeta, json.RawMessage, error) {
	path := queueItemsPath(orgID, projectID, queueID)
	data, raw, err := doList[queueItemListData](c, ctx, path, opts, "list queue items")
	if err != nil {
		return nil, PageMeta{}, nil, err
	}
	return data.Items, data.Meta, raw, nil
}

func (c *APIClient) GetQueueItem(
	ctx context.Context,
	orgID, projectID, queueID, itemID string,
) (*QueueItemInfo, json.RawMessage, error) {
	path := queueItemsPath(orgID, projectID, queueID) + "/" + url.PathEscape(itemID)
	data, raw, err := doGet[queueItemWrapper](c, ctx, path, "get queue item")
	if err != nil {
		return nil, nil, err
	}
	return &data.Item, raw, nil
}

func (c *APIClient) CreateQueueItem(
	ctx context.Context,
	orgID, projectID, queueID string,
	body QueueItemCreateBody,
) (*QueueItemInfo, json.RawMessage, error) {
	path := queueItemsPath(orgID, projectID, queueID)
	data, raw, err := doCreate[queueItemWrapper](c, ctx, path, body, "create queue item")
	if err != nil {
		return nil, nil, err
	}
	return &data.Item, raw, nil
}

func (c *APIClient) UpdateQueueItem(
	ctx context.Context,
	orgID, projectID, queueID, itemID string,
	body QueueItemUpdateBody,
) (*QueueItemInfo, json.RawMessage, error) {
	path := queueItemsPath(orgID, projectID, queueID) + "/" + url.PathEscape(itemID)
	data, raw, err := doUpdate[queueItemWrapper](c, ctx, path, body, "update queue item")
	if err != nil {
		return nil, nil, err
	}
	return &data.Item, raw, nil
}

func (c *APIClient) DeleteQueueItem(
	ctx context.Context,
	orgID, projectID, queueID, itemID string,
) (string, error) {
	path := queueItemsPath(orgID, projectID, queueID) + "/" + url.PathEscape(itemID)
	return c.doDelete(ctx, path, "delete queue item")
}

// ---------------------------------------------------------------------------
// Comments
// ---------------------------------------------------------------------------

type CommentInfo struct {
	ID           string          `json:"id"`
	ProjectID    string          `json:"project_id"`
	ObjectType   string          `json:"object_type"`
	ObjectID     string          `json:"object_id"`
	Content      string          `json:"content"`
	AuthorUserID string          `json:"author_user_id"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}

type CommentListOptions struct {
	ObjectType   string `url:"object_type,omitempty"`
	ObjectID     string `url:"object_id,omitempty"`
	AuthorUserID string `url:"author_user_id,omitempty"`
	Page         int    `url:"page,omitempty"`
	Limit        int    `url:"limit,omitempty"`
}

type CommentCreateBody struct {
	ObjectType   string `json:"object_type"`
	ObjectID     string `json:"object_id"`
	Content      string `json:"content"`
	AuthorUserID string `json:"author_user_id,omitempty"`
}

type commentListData struct {
	Comments []CommentInfo `json:"comments"`
	Meta     PageMeta      `json:"meta"`
}

type commentWrapper struct {
	Comment CommentInfo `json:"comment"`
}

func (c *APIClient) ListComments(
	ctx context.Context,
	orgID, projectID string,
	opts CommentListOptions,
) ([]CommentInfo, PageMeta, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/comments"
	data, raw, err := doList[commentListData](c, ctx, path, opts, "list comments")
	if err != nil {
		return nil, PageMeta{}, nil, err
	}
	return data.Comments, data.Meta, raw, nil
}

func (c *APIClient) GetComment(
	ctx context.Context,
	orgID, projectID, commentID string,
) (*CommentInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/comments/" + url.PathEscape(commentID)
	data, raw, err := doGet[commentWrapper](c, ctx, path, "get comment")
	if err != nil {
		return nil, nil, err
	}
	return &data.Comment, raw, nil
}

func (c *APIClient) CreateComment(
	ctx context.Context,
	orgID, projectID string,
	body CommentCreateBody,
) (*CommentInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/comments"
	data, raw, err := doCreate[commentWrapper](c, ctx, path, body, "create comment")
	if err != nil {
		return nil, nil, err
	}
	return &data.Comment, raw, nil
}

// ---------------------------------------------------------------------------
// Score Configs
// ---------------------------------------------------------------------------

type ScoreConfigInfo struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	DataType    string          `json:"data_type"`
	MinValue    *float64        `json:"min_value"`
	MaxValue    *float64        `json:"max_value"`
	Categories  json.RawMessage `json:"categories"`
	Description string          `json:"description"`
	IsArchived  bool            `json:"is_archived"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type ScoreConfigListOptions struct {
	Page  int `url:"page,omitempty"`
	Limit int `url:"limit,omitempty"`
}

type ScoreConfigCreateBody struct {
	Name        string          `json:"name"`
	DataType    string          `json:"data_type"`
	MinValue    *float64        `json:"min_value,omitempty"`
	MaxValue    *float64        `json:"max_value,omitempty"`
	Categories  json.RawMessage `json:"categories,omitempty"`
	Description string          `json:"description,omitempty"`
}

type ScoreConfigUpdateBody struct {
	Description *string         `json:"description,omitempty"`
	IsArchived  *bool           `json:"is_archived,omitempty"`
	MinValue    *float64        `json:"min_value,omitempty"`
	MaxValue    *float64        `json:"max_value,omitempty"`
	Categories  json.RawMessage `json:"categories,omitempty"`
}

type scoreConfigListData struct {
	Configs []ScoreConfigInfo `json:"configs"`
	Meta    PageMeta          `json:"meta"`
}

type scoreConfigWrapper struct {
	Config ScoreConfigInfo `json:"config"`
}

func (c *APIClient) ListScoreConfigs(
	ctx context.Context,
	orgID, projectID string,
	opts ScoreConfigListOptions,
) ([]ScoreConfigInfo, PageMeta, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/score-configs"
	data, raw, err := doList[scoreConfigListData](c, ctx, path, opts, "list score configs")
	if err != nil {
		return nil, PageMeta{}, nil, err
	}
	return data.Configs, data.Meta, raw, nil
}

func (c *APIClient) GetScoreConfig(
	ctx context.Context,
	orgID, projectID, configID string,
) (*ScoreConfigInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/score-configs/" +
		url.PathEscape(configID)
	data, raw, err := doGet[scoreConfigWrapper](c, ctx, path, "get score config")
	if err != nil {
		return nil, nil, err
	}
	return &data.Config, raw, nil
}

func (c *APIClient) CreateScoreConfig(
	ctx context.Context,
	orgID, projectID string,
	body ScoreConfigCreateBody,
) (*ScoreConfigInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/score-configs"
	data, raw, err := doCreate[scoreConfigWrapper](c, ctx, path, body, "create score config")
	if err != nil {
		return nil, nil, err
	}
	return &data.Config, raw, nil
}

func (c *APIClient) UpdateScoreConfig(
	ctx context.Context,
	orgID, projectID, configID string,
	body ScoreConfigUpdateBody,
) (*ScoreConfigInfo, json.RawMessage, error) {
	path := evalBasePath(orgID, projectID) + "/score-configs/" +
		url.PathEscape(configID)
	data, raw, err := doUpdate[scoreConfigWrapper](c, ctx, path, body, "update score config")
	if err != nil {
		return nil, nil, err
	}
	return &data.Config, raw, nil
}
