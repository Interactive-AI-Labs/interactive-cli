package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

type APIClient struct {
	apiKey            string
	cookies           []*http.Cookie
	httpClient        *http.Client
	hostname          string
	isApiKeyMode      bool
	cachedOrgId       string
	cachedOrgName     string
	cachedProjectId   string
	cachedProjectName string
}

type Organization struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	ProjectCount int    `json:"project_count"`
	Role         string `json:"role"`
}

type Project struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func NewAPIClient(
	hostname string,
	timeout time.Duration,
	apiKey string,
	cookies []*http.Cookie,
) (*APIClient, error) {
	if apiKey == "" && len(cookies) == 0 {
		return nil, fmt.Errorf("no authentication method available: provide an API key or log in")
	}

	client := &APIClient{
		apiKey:       apiKey,
		cookies:      cookies,
		httpClient:   &http.Client{Timeout: timeout},
		hostname:     hostname,
		isApiKeyMode: apiKey != "",
	}

	if client.isApiKeyMode {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := client.validateApiKey(ctx); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *APIClient) do(req *http.Request) (*http.Response, error) {
	if err := ApplyAuth(req, c.apiKey, c.cookies); err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil && req.Context().Err() != nil {
		return nil, req.Context().Err()
	}
	return resp, err
}

func (c *APIClient) newRequest(ctx context.Context, method, rawPath string) (*http.Request, error) {
	u, err := url.Parse(c.hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API hostname: %w", err)
	}
	decodedPath, _ := url.PathUnescape(rawPath)
	u.Path = decodedPath
	u.RawPath = rawPath
	return http.NewRequestWithContext(ctx, method, u.String(), nil)
}

func (c *APIClient) validateApiKey(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/validate-api-key")
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return fmt.Errorf("API key validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		if msg := ExtractServerMessage(body); msg != "" {
			return fmt.Errorf("API key validation failed: %s", msg)
		}
		return fmt.Errorf("API key validation failed with status %s", resp.Status)
	}

	c.cachedOrgId = resp.Header.Get("x-org-id")
	c.cachedOrgName = resp.Header.Get("x-org-name")
	c.cachedProjectName = resp.Header.Get("x-project-name")
	c.cachedProjectId = resp.Header.Get("x-project-id")

	if c.cachedOrgId == "" || c.cachedOrgName == "" || c.cachedProjectId == "" ||
		c.cachedProjectName == "" {
		return fmt.Errorf("API key validation failed")
	}

	return nil
}

func (c *APIClient) ListOrganizations(ctx context.Context) ([]Organization, error) {
	if c.isApiKeyMode {
		return []Organization{{
			Id:   c.cachedOrgId,
			Name: c.cachedOrgName,
		}}, nil
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/api/v1/session/organizations")
	if err != nil {
		return nil, fmt.Errorf("failed to create organizations request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("organizations request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		if msg := ExtractServerMessage(body); msg != "" {
			return nil, fmt.Errorf("failed to list organizations: %s", msg)
		}
		return nil, fmt.Errorf("failed to list organizations: server returned %s", resp.Status)
	}

	var payload struct {
		Organizations []Organization `json:"organizations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode organizations response: %w", err)
	}

	return payload.Organizations, nil
}

func (c *APIClient) GetOrgIdByName(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("organization name cannot be empty")
	}

	if c.isApiKeyMode {
		if c.cachedOrgName != "" && !strings.EqualFold(name, c.cachedOrgName) {
			return "", fmt.Errorf(
				"organization %q not found; API key is scoped to organization %q",
				name,
				c.cachedOrgName,
			)
		}
		if c.cachedOrgName != "" && strings.EqualFold(name, c.cachedOrgName) {
			return c.cachedOrgId, nil
		}
		if c.cachedOrgName == "" && strings.EqualFold(name, c.cachedOrgId) {
			return c.cachedOrgId, nil
		}
		return "", fmt.Errorf(
			"organization %q not found; API key is scoped to organization ID %q",
			name,
			c.cachedOrgId,
		)
	}

	orgs, err := c.ListOrganizations(ctx)
	if err != nil {
		return "", err
	}

	if len(orgs) == 0 {
		return "", fmt.Errorf("no organizations found in your session")
	}

	var matched []Organization
	for _, org := range orgs {
		if strings.EqualFold(org.Name, name) {
			matched = append(matched, org)
		}
	}

	if len(matched) == 0 {
		return "", fmt.Errorf("organization %q not found", name)
	}
	if len(matched) > 1 {
		return "", fmt.Errorf("organization name %q is ambiguous; please use a unique name", name)
	}

	return matched[0].Id, nil
}

func (c *APIClient) ListProjects(ctx context.Context, orgId string) ([]Project, error) {
	if c.isApiKeyMode {
		return []Project{{
			Id:   c.cachedProjectId,
			Name: c.cachedProjectName,
		}}, nil
	}

	req, err := c.newRequest(
		ctx,
		http.MethodGet,
		fmt.Sprintf("/api/v1/session/organizations/%s/projects", orgId),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create projects request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("projects request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		if msg := ExtractServerMessage(body); msg != "" {
			return nil, fmt.Errorf("failed to list projects: %s", msg)
		}
		return nil, fmt.Errorf("failed to list projects: server returned %s", resp.Status)
	}

	var payload struct {
		Projects []Project `json:"projects"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode projects response: %w", err)
	}

	return payload.Projects, nil
}

func (c *APIClient) GetProjectByName(
	ctx context.Context,
	orgId, projectName string,
) (string, error) {
	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		return "", fmt.Errorf("project name cannot be empty")
	}

	if c.isApiKeyMode {
		if orgId != c.cachedOrgId {
			return "", fmt.Errorf(
				"API key is scoped to organization ID %q, not %q",
				c.cachedOrgId,
				orgId,
			)
		}
		if c.cachedProjectName != "" && !strings.EqualFold(projectName, c.cachedProjectName) {
			return "", fmt.Errorf(
				"project %q not found; API key is scoped to project %q",
				projectName,
				c.cachedProjectName,
			)
		}
		if c.cachedProjectName != "" && strings.EqualFold(projectName, c.cachedProjectName) {
			return c.cachedProjectId, nil
		}
		if c.cachedProjectName == "" && strings.EqualFold(projectName, c.cachedProjectId) {
			return c.cachedProjectId, nil
		}
		return "", fmt.Errorf(
			"project %q not found; API key is scoped to project ID %q",
			projectName,
			c.cachedProjectId,
		)
	}

	projects, err := c.ListProjects(ctx, orgId)
	if err != nil {
		return "", err
	}

	if len(projects) == 0 {
		return "", fmt.Errorf("no projects found in organization")
	}

	var matched []Project
	for _, proj := range projects {
		if strings.EqualFold(proj.Name, projectName) {
			matched = append(matched, proj)
		}
	}

	if len(matched) == 0 {
		return "", fmt.Errorf("project %q not found", projectName)
	}
	if len(matched) > 1 {
		return "", fmt.Errorf("project name %q is ambiguous; please use a unique name", projectName)
	}

	return matched[0].Id, nil
}

type TraceInfo struct {
	ID               string   `json:"id"`
	Timestamp        string   `json:"timestamp"`
	Name             string   `json:"name"`
	SessionID        string   `json:"session_id"`
	UserID           string   `json:"user_id"`
	Release          string   `json:"release"`
	Version          string   `json:"version"`
	Public           bool     `json:"public"`
	Environment      string   `json:"environment"`
	Tags             []string `json:"tags"`
	HtmlPath         string   `json:"html_path"`
	LatencyMs        *float64 `json:"latency_ms"`
	TotalCost        *float64 `json:"total_cost"`
	ObservationCount *int     `json:"observation_count"`
	InputTokens      *int     `json:"input_tokens"`
	OutputTokens     *int     `json:"output_tokens"`
	TotalTokens      *int     `json:"total_tokens"`
	Level            string   `json:"level"`
}

type TraceDetail struct {
	TraceInfo
	Input    json.RawMessage `json:"input"`
	Output   json.RawMessage `json:"output"`
	Metadata json.RawMessage `json:"metadata"`
}

type TraceMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type traceListResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Traces []TraceInfo `json:"traces"`
		Meta   TraceMeta   `json:"meta"`
	} `json:"data"`
}

type traceGetResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Trace TraceDetail `json:"trace"`
	} `json:"data"`
}

type TraceListOptions struct {
	Page          int      `url:"page,omitempty"`
	Limit         int      `url:"limit,omitempty"`
	UserID        string   `url:"user_id,omitempty"`
	Name          string   `url:"name,omitempty"`
	SessionID     string   `url:"session_id,omitempty"`
	FromTimestamp string   `url:"from_timestamp,omitempty"`
	ToTimestamp   string   `url:"to_timestamp,omitempty"`
	OrderBy       string   `url:"order_by,omitempty"`
	Order         string   `url:"order,omitempty"`
	Tags          []string `url:"tags,omitempty"`
	Version       string   `url:"version,omitempty"`
	Release       string   `url:"release,omitempty"`
	Environment   []string `url:"environment,omitempty"`
	MinCost       *float64 `url:"min_cost,omitempty"`
	MaxCost       *float64 `url:"max_cost,omitempty"`
	MinLatency    *float64 `url:"min_latency,omitempty"`
	MaxLatency    *float64 `url:"max_latency,omitempty"`
	MinTokens     *int     `url:"min_tokens,omitempty"`
	MaxTokens     *int     `url:"max_tokens,omitempty"`
	Model         string   `url:"model,omitempty"`
	HasError      *bool    `url:"has_error,omitempty"`
	Level         string   `url:"level,omitempty"`
	Search        string   `url:"search,omitempty"`
	Fields        string   `url:"fields,omitempty"`
}

// ObservationInfo represents an observation returned by the list endpoint.
type ObservationInfo struct {
	ID                  string          `json:"id"`
	TraceID             string          `json:"trace_id"`
	Type                string          `json:"type"`
	Name                string          `json:"name"`
	StartTime           string          `json:"start_time"`
	EndTime             string          `json:"end_time"`
	ParentObservationID string          `json:"parent_observation_id"`
	Level               string          `json:"level"`
	StatusMessage       string          `json:"status_message"`
	Model               string          `json:"model"`
	InputTokens         *int            `json:"input_tokens"`
	OutputTokens        *int            `json:"output_tokens"`
	TotalTokens         *int            `json:"total_tokens"`
	TotalCost           *float64        `json:"total_cost"`
	LatencyMs           *float64        `json:"latency_ms"`
	Input               json.RawMessage `json:"input,omitempty"`
	Output              json.RawMessage `json:"output,omitempty"`
	Metadata            json.RawMessage `json:"metadata,omitempty"`
}

// ObservationDetail represents a single observation from the get endpoint.
type ObservationDetail struct {
	ObservationInfo
	ModelParameters json.RawMessage `json:"model_parameters,omitempty"`
	PromptName      string          `json:"prompt_name"`
	PromptVersion   *int            `json:"prompt_version"`
}

type observationListResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Observations []ObservationInfo `json:"observations"`
	} `json:"data"`
}

type observationGetResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Observation ObservationDetail `json:"observation"`
	} `json:"data"`
}

// ListTraces calls the platform trace exploration API.
// orgID and projectID are required for the new endpoint path.
func (c *APIClient) ListTraces(
	ctx context.Context,
	orgID, projectID string,
	opts TraceListOptions,
) ([]TraceInfo, TraceMeta, json.RawMessage, error) {
	path := fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s/traces",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, TraceMeta{}, nil, fmt.Errorf("failed to create request: %w", err)
	}

	q, _ := query.Values(opts)
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, TraceMeta{}, nil, fmt.Errorf("traces list request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, TraceMeta{}, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, TraceMeta{}, nil, errors.New(msg)
		}
		return nil, TraceMeta{}, nil, fmt.Errorf(
			"failed to list traces: server returned %s",
			resp.Status,
		)
	}

	var result traceListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, TraceMeta{}, nil, fmt.Errorf("failed to decode traces response: %w", err)
	}

	return result.Data.Traces, result.Data.Meta, respBody, nil
}

// GetTrace retrieves a single trace from the platform API.
func (c *APIClient) GetTrace(
	ctx context.Context,
	orgID, projectID, traceID, fields string,
) (*TraceDetail, json.RawMessage, error) {
	path := fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s/traces/%s",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
		url.PathEscape(traceID),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	if fields != "" {
		q := req.URL.Query()
		q.Set("fields", fields)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("trace get request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, nil, errors.New(msg)
		}
		return nil, nil, fmt.Errorf("failed to get trace: server returned %s", resp.Status)
	}

	var result traceGetResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to decode trace response: %w", err)
	}

	return &result.Data.Trace, respBody, nil
}

// ListObservations retrieves observations for a trace.
func (c *APIClient) ListObservations(
	ctx context.Context,
	orgID, projectID, traceID string,
	includeIO bool,
) ([]ObservationInfo, json.RawMessage, error) {
	path := fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s/traces/%s/observations",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
		url.PathEscape(traceID),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	if includeIO {
		q := req.URL.Query()
		q.Set("include_io", "true")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("observations list request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, nil, errors.New(msg)
		}
		return nil, nil, fmt.Errorf(
			"failed to list observations: server returned %s",
			resp.Status,
		)
	}

	var result observationListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to decode observations response: %w", err)
	}

	return result.Data.Observations, respBody, nil
}

// GetObservation retrieves a single observation by ID.
func (c *APIClient) GetObservation(
	ctx context.Context,
	orgID, projectID, observationID string,
) (*ObservationDetail, json.RawMessage, error) {
	path := fmt.Sprintf(
		"/api/platform/v1/organizations/%s/projects/%s/observations/%s",
		url.PathEscape(orgID),
		url.PathEscape(projectID),
		url.PathEscape(observationID),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("observation get request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, nil, errors.New(msg)
		}
		return nil, nil, fmt.Errorf(
			"failed to get observation: server returned %s",
			resp.Status,
		)
	}

	var result observationGetResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to decode observation response: %w", err)
	}

	return &result.Data.Observation, respBody, nil
}

func (c *APIClient) GetProjectId(
	ctx context.Context,
	orgName, projectName string,
) (string, string, error) {
	orgName = strings.TrimSpace(orgName)
	projectName = strings.TrimSpace(projectName)

	if orgName == "" {
		return "", "", fmt.Errorf("organization name cannot be empty")
	}
	if projectName == "" {
		return "", "", fmt.Errorf("project name cannot be empty")
	}

	if c.isApiKeyMode {
		orgMatch := false
		if c.cachedOrgName != "" {
			orgMatch = strings.EqualFold(orgName, c.cachedOrgName)
		} else {
			orgMatch = strings.EqualFold(orgName, c.cachedOrgId)
		}

		if !orgMatch {
			if c.cachedOrgName != "" {
				return "", "", fmt.Errorf(
					"organization %q not found; API key is scoped to organization %q",
					orgName,
					c.cachedOrgName,
				)
			}
			return "", "", fmt.Errorf(
				"organization %q not found; API key is scoped to organization ID %q",
				orgName,
				c.cachedOrgId,
			)
		}

		projectMatch := false
		if c.cachedProjectName != "" {
			projectMatch = strings.EqualFold(projectName, c.cachedProjectName)
		} else {
			projectMatch = strings.EqualFold(projectName, c.cachedProjectId)
		}

		if !projectMatch {
			if c.cachedProjectName != "" {
				return "", "", fmt.Errorf(
					"project %q not found; API key is scoped to project %q",
					projectName,
					c.cachedProjectName,
				)
			}
			return "", "", fmt.Errorf(
				"project %q not found; API key is scoped to project ID %q",
				projectName,
				c.cachedProjectId,
			)
		}

		return c.cachedOrgId, c.cachedProjectId, nil
	}

	orgId, err := c.GetOrgIdByName(ctx, orgName)
	if err != nil {
		return "", "", err
	}

	projectId, err := c.GetProjectByName(ctx, orgId, projectName)
	if err != nil {
		return "", "", err
	}

	return orgId, projectId, nil
}

type PromptInfo struct {
	Name          string   `json:"name"`
	Versions      []int    `json:"versions"`
	Labels        []string `json:"labels"`
	Tags          []string `json:"tags"`
	LastUpdatedAt string   `json:"lastUpdatedAt"`
}

type PromptDetail struct {
	Id             string          `json:"id"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	Version        int             `json:"version"`
	ProjectId      string          `json:"projectId"`
	Prompt         json.RawMessage `json:"prompt"`
	Config         json.RawMessage `json:"config"`
	Labels         []string        `json:"labels"`
	Tags           []string        `json:"tags"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
	ExpectedFormat string          `json:"expectedFormat"`
}

type CreatePromptBody struct {
	Name       string   `json:"name"`
	Prompt     string   `json:"prompt"`
	Labels     []string `json:"labels,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	PromptType string   `json:"promptType,omitempty"`
}

type promptAPIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

type promptListData struct {
	Prompts    []PromptInfo `json:"prompts"`
	TotalCount int          `json:"totalCount"`
}

type PromptListResponse struct {
	Prompts    []PromptInfo
	TotalCount int
}

type PromptListOptions struct {
	Page   int
	Limit  int
	Folder string
}

func promptBasePath(projectId, routeSegment string) string {
	base := fmt.Sprintf("/api/platform/v1/projects/%s/prompts", url.PathEscape(projectId))
	if routeSegment != "" {
		return base + "/" + url.PathEscape(routeSegment)
	}
	return base
}

func (c *APIClient) CreatePrompt(
	ctx context.Context,
	projectId string,
	routeSegment string,
	body CreatePromptBody,
) (*PromptDetail, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	path := promptBasePath(projectId, routeSegment)
	req, err := c.newRequest(ctx, http.MethodPost, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("prompt creation request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("prompt creation failed with status %s", resp.Status)
	}

	var envelope promptAPIResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode prompt response: %w", err)
	}

	var result PromptDetail
	if err := json.Unmarshal(envelope.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to decode prompt data: %w", err)
	}

	return &result, nil
}

func (c *APIClient) ListPrompts(
	ctx context.Context,
	projectId string,
	routeSegment string,
	opts PromptListOptions,
) (*PromptListResponse, error) {
	path := promptBasePath(projectId, routeSegment)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()

	// The generic /prompts endpoint (empty routeSegment) expects all query
	// parameters encoded as a JSON object in a single "input" query param.
	// Typed endpoints use flat query parameters.
	if routeSegment == "" {
		inputMap := map[string]interface{}{
			"filter":  []interface{}{},
			"orderBy": map[string]interface{}{},
		}
		if opts.Folder != "" {
			inputMap["folder"] = opts.Folder
		}
		if opts.Page > 0 {
			inputMap["page"] = opts.Page
		}
		if opts.Limit > 0 {
			inputMap["limit"] = opts.Limit
		}
		inputBytes, err := json.Marshal(inputMap)
		if err != nil {
			return nil, fmt.Errorf("failed to encode list parameters: %w", err)
		}
		q.Set("input", string(inputBytes))
	} else {
		if opts.Page > 0 {
			q.Set("page", fmt.Sprintf("%d", opts.Page))
		}
		if opts.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("prompt list request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to list prompts: server returned %s", resp.Status)
	}

	var envelope promptAPIResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode prompts response: %w", err)
	}

	var listData promptListData
	if err := json.Unmarshal(envelope.Data, &listData); err != nil {
		return nil, fmt.Errorf("failed to decode prompts data: %w", err)
	}

	return &PromptListResponse{
		Prompts:    listData.Prompts,
		TotalCount: listData.TotalCount,
	}, nil
}

func (c *APIClient) GetPrompt(
	ctx context.Context,
	projectId string,
	routeSegment string,
	name string,
	version int,
	label string,
) (*PromptDetail, error) {
	path := promptBasePath(projectId, routeSegment) + "/" + url.PathEscape(name)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if version > 0 {
		q.Set("version", fmt.Sprintf("%d", version))
	}
	if label != "" {
		q.Set("label", label)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("prompt get request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to get prompt: server returned %s", resp.Status)
	}

	var envelope promptAPIResponse
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode prompt response: %w", err)
	}

	var result PromptDetail
	if err := json.Unmarshal(envelope.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to decode prompt data: %w", err)
	}

	return &result, nil
}

func (c *APIClient) DeletePrompt(
	ctx context.Context,
	projectId string,
	routeSegment string,
	name string,
	version int,
	label string,
) error {
	path := promptBasePath(projectId, routeSegment) + "/" + url.PathEscape(name)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if version > 0 {
		q.Set("version", fmt.Sprintf("%d", version))
	}
	if label != "" {
		q.Set("label", label)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return fmt.Errorf("prompt deletion request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return fmt.Errorf("prompt deletion failed with status %s", resp.Status)
	}

	return nil
}

type SchemaResponse struct {
	Schema        json.RawMessage `json:"schema"`
	SchemaVersion string          `json:"schemaVersion"`
}

// GetPromptSchema fetches the JSON Schema for a prompt type from the public
// schemas endpoint. No authentication is required.
func GetPromptSchema(
	ctx context.Context,
	hostname string,
	timeout time.Duration,
	typeName string,
) (*SchemaResponse, error) {
	u, err := url.Parse(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hostname: %w", err)
	}
	rawPath := fmt.Sprintf("/api/platform/v1/prompts/schemas/%s", url.PathEscape(typeName))
	decodedPath, _ := url.PathUnescape(rawPath)
	u.Path = decodedPath
	u.RawPath = rawPath

	httpClient := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Could not fetch schema. Ensure --hostname is correct: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(body); msg != "" {
			return nil, fmt.Errorf("failed to fetch schema: %s", msg)
		}
		return nil, fmt.Errorf("failed to fetch schema: server returned %s", resp.Status)
	}

	var envelope struct {
		Success bool           `json:"success"`
		Data    SchemaResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode schema response: %w", err)
	}

	if !envelope.Success {
		return nil, fmt.Errorf("schema endpoint returned success=false")
	}

	return &envelope.Data, nil
}

func (c *APIClient) DeletePromptByName(
	ctx context.Context,
	projectId string,
	routeSegment string,
	name string,
) error {
	path := promptBasePath(projectId, routeSegment) + "/by-name/" + url.PathEscape(name)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return fmt.Errorf("prompt deletion request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(respBody); msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return fmt.Errorf("prompt deletion failed with status %s", resp.Status)
	}

	return nil
}
