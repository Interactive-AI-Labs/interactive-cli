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

type DeploymentClient struct {
	apiKey     string
	cookies    []*http.Cookie
	httpClient *http.Client
	hostname   string
}

type SecretInfo struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	CreatedAt string   `json:"createdAt"`
	Keys      []string `json:"keys"`
}

type ImageInfo struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type ReplicaInfo struct {
	Name      string `json:"name"`
	Phase     string `json:"phase"`
	Status    string `json:"status"`
	Ready     bool   `json:"ready"`
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
	StartTime string `json:"startTime,omitempty"`
}

func NewDeploymentClient(hostname string, timeout time.Duration, apiKey string, cookies []*http.Cookie) (*DeploymentClient, error) {
	if apiKey == "" && len(cookies) == 0 {
		return nil, fmt.Errorf("no authentication method available: provide an API key or log in")
	}

	return &DeploymentClient{
		apiKey:     apiKey,
		cookies:    cookies,
		httpClient: &http.Client{Timeout: timeout},
		hostname:   hostname,
	}, nil
}

func (c *DeploymentClient) do(req *http.Request) (*http.Response, error) {
	if err := ApplyAuth(req, c.apiKey, c.cookies); err != nil {
		return nil, err
	}
	return c.httpClient.Do(req)
}

func (c *DeploymentClient) newRequest(ctx context.Context, method, path string) (*http.Request, error) {
	u, err := url.Parse(c.hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment hostname: %w", err)
	}
	u.Path = path
	return http.NewRequestWithContext(ctx, method, u.String(), nil)
}

func (c *DeploymentClient) CreateService(
	ctx context.Context,
	orgId,
	projectId string,
	serviceName string,
	req CreateServiceBody,
) (string, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/services/%s", orgId, projectId, serviceName)
	reqHTTP, err := c.newRequest(ctx, http.MethodPost, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	resp, err := c.do(reqHTTP)
	if err != nil {
		return "", fmt.Errorf("service creation request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("service creation failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) UpdateService(
	ctx context.Context,
	orgId,
	projectId string,
	serviceName string,
	req CreateServiceBody,
) (string, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/services/%s", orgId, projectId, serviceName)
	reqHTTP, err := c.newRequest(ctx, http.MethodPut, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	resp, err := c.do(reqHTTP)
	if err != nil {
		return "", fmt.Errorf("service update request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("service update failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) DeleteService(
	ctx context.Context,
	orgId,
	projectId string,
	serviceName string,
) (string, error) {
	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/services/%s", orgId, projectId, serviceName)
	reqHTTP, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(reqHTTP)
	if err != nil {
		return "", fmt.Errorf("service deletion request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("service deletion failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) ListServices(
	ctx context.Context,
	orgId,
	projectId string,
	stackId string,
) ([]ServiceOutput, error) {
	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/services", orgId, projectId)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if stackId != "" {
		q := req.URL.Query()
		q.Set("stackId", stackId)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("service list request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("service listing failed with status %s", resp.Status)
	}

	var result struct {
		Services []ServiceOutput `json:"services"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode services response: %w", err)
	}

	return result.Services, nil
}

func (c *DeploymentClient) CreateSecret(
	ctx context.Context,
	orgId,
	projectId,
	secretName string,
	data map[string]string,
) (string, error) {
	reqBody := struct {
		Data map[string]string `json:"data"`
	}{
		Data: data,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/secrets/%s", orgId, projectId, secretName)
	req, err := c.newRequest(ctx, http.MethodPost, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("secret creation request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("secret creation failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) UpdateSecret(
	ctx context.Context,
	orgId,
	projectId,
	secretName string,
	data map[string]string,
) (string, error) {
	reqBody := struct {
		Data map[string]string `json:"data"`
	}{
		Data: data,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/secrets/%s", orgId, projectId, secretName)
	req, err := c.newRequest(ctx, http.MethodPut, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("secret update request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("secret update failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) DeleteSecret(
	ctx context.Context,
	orgId,
	projectId,
	secretName string,
) (string, error) {
	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/secrets/%s", orgId, projectId, secretName)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("secret delete request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("secret delete failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) ListSecrets(
	ctx context.Context,
	orgId,
	projectId string,
) ([]SecretInfo, error) {
	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/secrets", orgId, projectId)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("secrets request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to list secrets: server returned %s", resp.Status)
	}

	var result struct {
		Secrets []SecretInfo `json:"secrets"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode secrets response: %w", err)
	}

	return result.Secrets, nil
}

func (c *DeploymentClient) ListImages(
	ctx context.Context,
	orgId string,
) ([]ImageInfo, error) {
	path := fmt.Sprintf("/v1/organizations/%s/images", orgId)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(body); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to list images: server returned %s", resp.Status)
	}

	var result struct {
		Images []ImageInfo `json:"images"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Images, nil
}

func (c *DeploymentClient) ListReplicas(
	ctx context.Context,
	orgId,
	projectId,
	serviceName string,
) ([]ReplicaInfo, error) {
	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/services/%s/replicas", orgId, projectId, serviceName)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("replicas request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("replicas request failed with status %s", resp.Status)
	}

	var result struct {
		Replicas []ReplicaInfo `json:"replicas"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode replicas response: %w", err)
	}

	return result.Replicas, nil
}

func (c *DeploymentClient) GetLogs(
	ctx context.Context,
	orgId,
	projectId,
	replicaName string,
	follow bool,
) (io.ReadCloser, error) {
	path := fmt.Sprintf("/v1/organizations/%s/projects/%s/services/replicas/%s/logs", orgId, projectId, replicaName)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if follow {
		q := req.URL.Query()
		q.Set("follow", "true")
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("logs request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("logs request failed with status %s", resp.Status)
	}

	return resp.Body, nil
}
