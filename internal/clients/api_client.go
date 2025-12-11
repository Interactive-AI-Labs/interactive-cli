package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func NewAPIClient(hostname string, timeout time.Duration, apiKey string, cookies []*http.Cookie) (*APIClient, error) {
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

func (c *APIClient) newRequest(ctx context.Context, method, path string) (*http.Request, error) {
	u, err := url.Parse(c.hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API hostname: %w", err)
	}
	u.Path = path
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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if msg := ExtractServerMessage(body); msg != "" {
			return fmt.Errorf("API key validation failed: %s", msg)
		}
		return fmt.Errorf("API key validation failed with status %s", resp.Status)
	}

	c.cachedOrgId = resp.Header.Get("x-org-id")
	c.cachedOrgName = resp.Header.Get("x-org-name")
	c.cachedProjectName = resp.Header.Get("x-project-name")
	c.cachedProjectId = resp.Header.Get("x-project-id")

	if c.cachedOrgId == "" || c.cachedOrgName == "" || c.cachedProjectId == "" || c.cachedProjectName == "" {
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

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(body); msg != "" {
			return nil, fmt.Errorf("failed to list organizations: %s", msg)
		}
		return nil, fmt.Errorf("failed to list organizations: server returned %s", resp.Status)
	}

	var payload struct {
		Organizations []Organization `json:"organizations"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
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
			return "", fmt.Errorf("organization %q not found; API key is scoped to organization %q", name, c.cachedOrgName)
		}
		if c.cachedOrgName != "" && strings.EqualFold(name, c.cachedOrgName) {
			return c.cachedOrgId, nil
		}
		if c.cachedOrgName == "" && strings.EqualFold(name, c.cachedOrgId) {
			return c.cachedOrgId, nil
		}
		return "", fmt.Errorf("organization %q not found; API key is scoped to organization ID %q", name, c.cachedOrgId)
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

	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v1/session/organizations/%s/projects", orgId))
	if err != nil {
		return nil, fmt.Errorf("failed to create projects request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("projects request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(body); msg != "" {
			return nil, fmt.Errorf("failed to list projects: %s", msg)
		}
		return nil, fmt.Errorf("failed to list projects: server returned %s", resp.Status)
	}

	var payload struct {
		Projects []Project `json:"projects"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to decode projects response: %w", err)
	}

	return payload.Projects, nil
}

func (c *APIClient) GetProjectByName(ctx context.Context, orgId, projectName string) (string, error) {
	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		return "", fmt.Errorf("project name cannot be empty")
	}

	if c.isApiKeyMode {
		if orgId != c.cachedOrgId {
			return "", fmt.Errorf("API key is scoped to organization ID %q, not %q", c.cachedOrgId, orgId)
		}
		if c.cachedProjectName != "" && !strings.EqualFold(projectName, c.cachedProjectName) {
			return "", fmt.Errorf("project %q not found; API key is scoped to project %q", projectName, c.cachedProjectName)
		}
		if c.cachedProjectName != "" && strings.EqualFold(projectName, c.cachedProjectName) {
			return c.cachedProjectId, nil
		}
		if c.cachedProjectName == "" && strings.EqualFold(projectName, c.cachedProjectId) {
			return c.cachedProjectId, nil
		}
		return "", fmt.Errorf("project %q not found; API key is scoped to project ID %q", projectName, c.cachedProjectId)
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

func (c *APIClient) GetProjectId(ctx context.Context, orgName, projectName string) (string, string, error) {
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
				return "", "", fmt.Errorf("organization %q not found; API key is scoped to organization %q", orgName, c.cachedOrgName)
			}
			return "", "", fmt.Errorf("organization %q not found; API key is scoped to organization ID %q", orgName, c.cachedOrgId)
		}

		projectMatch := false
		if c.cachedProjectName != "" {
			projectMatch = strings.EqualFold(projectName, c.cachedProjectName)
		} else {
			projectMatch = strings.EqualFold(projectName, c.cachedProjectId)
		}

		if !projectMatch {
			if c.cachedProjectName != "" {
				return "", "", fmt.Errorf("project %q not found; API key is scoped to project %q", projectName, c.cachedProjectName)
			}
			return "", "", fmt.Errorf("project %q not found; API key is scoped to project ID %q", projectName, c.cachedProjectId)
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
