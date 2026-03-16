package internal

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
	return c.httpClient.Do(req)
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
	ID          string   `json:"id"`
	Timestamp   string   `json:"timestamp"`
	Name        string   `json:"name"`
	SessionID   string   `json:"sessionId"`
	UserID      string   `json:"userId"`
	Release     string   `json:"release"`
	Version     string   `json:"version"`
	Public      bool     `json:"public"`
	Environment string   `json:"environment"`
	Tags        []string `json:"tags"`
	HtmlPath    string   `json:"htmlPath"`
	Latency     *float64 `json:"latency"`
	TotalCost   *float64 `json:"totalCost"`
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
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type traceListResponse struct {
	Data []TraceInfo `json:"data"`
	Meta TraceMeta   `json:"meta"`
}

type TraceListOptions struct {
	Page          int      `url:"page,omitempty"`
	Limit         int      `url:"limit,omitempty"`
	UserID        string   `url:"userId,omitempty"`
	Name          string   `url:"name,omitempty"`
	SessionID     string   `url:"sessionId,omitempty"`
	FromTimestamp string   `url:"fromTimestamp,omitempty"`
	ToTimestamp   string   `url:"toTimestamp,omitempty"`
	OrderBy       string   `url:"orderBy,omitempty"`
	Tags          []string `url:"tags,omitempty"`
	Version       string   `url:"version,omitempty"`
	Release       string   `url:"release,omitempty"`
	Environment   []string `url:"environment,omitempty"`
}

func (c *APIClient) ListTraces(
	ctx context.Context,
	opts TraceListOptions,
) ([]TraceInfo, TraceMeta, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/api/public/traces")
	if err != nil {
		return nil, TraceMeta{}, fmt.Errorf("failed to create request: %w", err)
	}

	q, _ := query.Values(opts)
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, TraceMeta{}, fmt.Errorf("traces list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, TraceMeta{}, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, TraceMeta{}, errors.New(msg)
		}
		return nil, TraceMeta{}, fmt.Errorf(
			"failed to list traces: server returned %s",
			resp.Status,
		)
	}

	var result traceListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, TraceMeta{}, fmt.Errorf("failed to decode traces response: %w", err)
	}

	return result.Data, result.Meta, nil
}

func (c *APIClient) GetTrace(ctx context.Context, traceID string) (*TraceDetail, error) {
	path := fmt.Sprintf("/api/public/traces/%s", url.PathEscape(traceID))
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("trace get request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, errors.New(msg)
		}
		return nil, fmt.Errorf("failed to get trace: server returned %s", resp.Status)
	}

	var trace TraceDetail
	if err := json.NewDecoder(resp.Body).Decode(&trace); err != nil {
		return nil, fmt.Errorf("failed to decode trace response: %w", err)
	}

	return &trace, nil
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
	Type          string   `json:"type"`
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
	Name   string          `json:"name"`
	Prompt json.RawMessage `json:"prompt"`
	Labels []string        `json:"labels,omitempty"`
	Tags   []string        `json:"tags,omitempty"`
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
	Folder string
	Type   string
	Page   int
	Limit  int
}

func promptBasePath(projectId, routeSegment string) string {
	base := fmt.Sprintf("/api/platform/v1/projects/%s/prompts", url.PathEscape(projectId))
	if routeSegment != "" {
		return base + "/" + routeSegment
	}
	return base
}

func (c *APIClient) CreatePrompt(
	ctx context.Context,
	projectId string,
	routeSegment string,
	body CreatePromptBody,
	skipSchema bool,
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

	if skipSchema {
		q := req.URL.Query()
		q.Set("skip_schema", "true")
		req.URL.RawQuery = q.Encode()
	}

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
	if opts.Folder != "" {
		q.Set("folder", opts.Folder)
	}
	if opts.Type != "" {
		q.Set("type", opts.Type)
	}
	if opts.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", opts.Page))
	}
	if opts.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", opts.Limit))
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
