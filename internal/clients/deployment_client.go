package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type DeploymentClient struct {
	apiKey     string
	cookies    []*http.Cookie
	httpClient *http.Client
	hostname   string
}

type servicesResponse struct {
	Services []ServiceOutput `json:"services"`
}

type secretsResponse struct {
	Secrets []SecretInfo `json:"secrets"`
}

type imagesResponse struct {
	Images []ImageInfo `json:"images"`
}

type replicasResponse struct {
	Replicas []ReplicaInfo `json:"replicas"`
}

type vectorStoresResponse struct {
	VectorStores []VectorStoreInfo `json:"vectorStores"`
}

type SecretInfo struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	CreatedAt string            `json:"createdAt"`
	Keys      []string          `json:"keys"`
	Data      map[string]string `json:"data,omitempty"`
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

type ReplicaStatus struct {
	Name                 string                  `json:"name"`
	Ready                bool                    `json:"ready"`
	Status               string                  `json:"status"`
	StartTime            string                  `json:"startTime,omitempty"`
	RestartCount         int                     `json:"restartCount"`
	LastTerminationState *ReplicaLastTermination `json:"lastTerminationState,omitempty"`
	Resources            *ReplicaResources       `json:"resources,omitempty"`
	Healthcheck          *ReplicaHealthcheck     `json:"healthcheck,omitempty"`
	Events               []ReplicaEvent          `json:"events,omitempty"`
}

type ReplicaLastTermination struct {
	Reason     string `json:"reason"`
	ExitCode   int32  `json:"exitCode"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
}

type ReplicaResources struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type ReplicaHealthcheck struct {
	Path                string `json:"path"`
	InitialDelaySeconds int    `json:"initialDelaySeconds"`
}

type ReplicaEvent struct {
	Type           string `json:"type"`
	Reason         string `json:"reason"`
	Message        string `json:"message"`
	Count          int    `json:"count"`
	FirstTimestamp string `json:"firstTimestamp"`
	LastTimestamp  string `json:"lastTimestamp"`
}

func NewDeploymentClient(
	hostname string,
	timeout time.Duration,
	apiKey string,
	cookies []*http.Cookie,
) (*DeploymentClient, error) {
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
	resp, err := c.httpClient.Do(req)
	if err != nil && req.Context().Err() != nil {
		return nil, req.Context().Err()
	}
	return resp, err
}

func (c *DeploymentClient) newRequest(
	ctx context.Context,
	method, path string,
) (*http.Request, error) {
	u, err := url.Parse(c.hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment hostname: %w", err)
	}
	u.Path = path
	return http.NewRequestWithContext(ctx, method, u.String(), nil)
}

type CreateServiceBody struct {
	ServicePort int          `json:"servicePort"`
	Image       ImageSpec    `json:"image"`
	Resources   Resources    `json:"resources"`
	Env         []EnvVar     `json:"env,omitempty"`
	SecretRefs  []SecretRef  `json:"secretRefs,omitempty"`
	Endpoint    bool         `json:"endpoint,omitempty"`
	Replicas    int          `json:"replicas,omitempty"`
	Autoscaling *Autoscaling `json:"autoscaling,omitempty"`
	Healthcheck *Healthcheck `json:"healthcheck,omitempty"`
	Schedule    *Schedule    `json:"schedule,omitempty"`
	StackId     string       `json:"stackId,omitempty"`
}

type Resources struct {
	Memory string `json:"memory" yaml:"memory"`
	CPU    string `json:"cpu"    yaml:"cpu"`
}

type Autoscaling struct {
	Enabled          bool `json:"enabled"                    yaml:"enabled"`
	MinReplicas      int  `json:"minReplicas,omitempty"      yaml:"minReplicas,omitempty"`
	MaxReplicas      int  `json:"maxReplicas,omitempty"      yaml:"maxReplicas,omitempty"`
	CPUPercentage    int  `json:"cpuPercentage,omitempty"    yaml:"cpuPercentage,omitempty"`
	MemoryPercentage int  `json:"memoryPercentage,omitempty" yaml:"memoryPercentage,omitempty"`
}

type Healthcheck struct {
	Path                string `json:"path,omitempty"                yaml:"path,omitempty"`
	InitialDelaySeconds int    `json:"initialDelaySeconds,omitempty" yaml:"initialDelaySeconds,omitempty"`
}

type Schedule struct {
	Uptime   string `json:"uptime,omitempty"   yaml:"uptime,omitempty"`
	Downtime string `json:"downtime,omitempty" yaml:"downtime,omitempty"`
	Timezone string `json:"timezone,omitempty" yaml:"timezone,omitempty"`
}

type ImageSpec struct {
	Type       string `json:"type"                 yaml:"type"`
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty"`
	Name       string `json:"name"                 yaml:"name"`
	Tag        string `json:"tag"                  yaml:"tag"`
}

type EnvVar struct {
	Name  string `json:"name"  yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type SecretRef struct {
	SecretName string `json:"secretName" yaml:"secretName"`
}

type ServiceOutput struct {
	Name      string `json:"name"`
	ProjectId string `json:"projectId"`
	Revision  int    `json:"revision"`
	Status    string `json:"status"`
	Updated   string `json:"updated,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
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

	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(serviceName),
	)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

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

	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(serviceName),
	)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

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
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(serviceName),
	)
	reqHTTP, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(reqHTTP)
	if err != nil {
		return "", fmt.Errorf("service deletion request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("service deletion failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) RestartService(
	ctx context.Context,
	orgId,
	projectId string,
	serviceName string,
) (string, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services/%s/restart",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(serviceName),
	)
	reqHTTP, err := c.newRequest(ctx, http.MethodPost, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(reqHTTP)
	if err != nil {
		return "", fmt.Errorf("service restart request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return serverMessage, nil
	}

	if serverMessage != "" {
		return "", fmt.Errorf("%s", serverMessage)
	}
	return "", fmt.Errorf("service restart failed with status %s", resp.Status)
}

func (c *DeploymentClient) ListServices(
	ctx context.Context,
	orgId,
	projectId string,
	stackId string,
) ([]ServiceOutput, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
	)
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

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("service listing failed with status %s", resp.Status)
	}

	var result servicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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

	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/secrets/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(secretName),
	)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("secret creation failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) ReplaceSecret(
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

	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/secrets/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(secretName),
	)
	req, err := c.newRequest(ctx, http.MethodPut, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("secret replace request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("secret replace failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) DeleteSecret(
	ctx context.Context,
	orgId,
	projectId,
	secretName string,
) (string, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/secrets/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(secretName),
	)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("secret delete request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("secret delete failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) DeleteSecretKey(
	ctx context.Context,
	orgId,
	projectId,
	secretName,
	keyName string,
) (string, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/secrets/%s/keys/%s",
		url.PathEscape(
			orgId,
		),
		url.PathEscape(projectId),
		url.PathEscape(secretName),
		url.PathEscape(keyName),
	)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("secret key delete request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return serverMessage, nil
	}

	if serverMessage != "" {
		return "", fmt.Errorf("%s", serverMessage)
	}
	return "", fmt.Errorf("secret key delete failed with status %s", resp.Status)
}

func (c *DeploymentClient) UpdateSecretKey(
	ctx context.Context,
	orgId,
	projectId,
	secretName,
	keyName,
	value string,
) (string, error) {
	reqBody := struct {
		Value string `json:"value"`
	}{
		Value: value,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/secrets/%s/keys/%s",
		url.PathEscape(
			orgId,
		),
		url.PathEscape(projectId),
		url.PathEscape(secretName),
		url.PathEscape(keyName),
	)
	req, err := c.newRequest(ctx, http.MethodPut, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("secret key update request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("secret key update failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) GetSecret(
	ctx context.Context,
	orgId,
	projectId,
	secretName string,
) (*SecretInfo, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/secrets/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(secretName),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("secret request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to get secret: server returned %s", resp.Status)
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode secret response: %w", err)
	}

	var secret SecretInfo
	if val, ok := raw["secret"]; ok {
		if err := json.Unmarshal(val, &secret); err != nil {
			return nil, fmt.Errorf("failed to decode secret object: %w", err)
		}
	} else {
		rawBytes, _ := json.Marshal(raw)
		if err := json.Unmarshal(rawBytes, &secret); err != nil {
			return nil, fmt.Errorf("failed to decode secret object: %w", err)
		}
	}

	if len(secret.Keys) == 0 && len(secret.Data) > 0 {
		for k := range secret.Data {
			secret.Keys = append(secret.Keys, k)
		}
		sort.Strings(secret.Keys)
	}

	return &secret, nil
}

func (c *DeploymentClient) ListSecrets(
	ctx context.Context,
	orgId,
	projectId string,
) ([]SecretInfo, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/secrets",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("secrets request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to list secrets: server returned %s", resp.Status)
	}

	var result secretsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode secrets response: %w", err)
	}

	return result.Secrets, nil
}

func (c *DeploymentClient) ListImages(
	ctx context.Context,
	orgId,
	projectId string,
) ([]ImageInfo, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/images",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		if msg := ExtractServerMessage(body); msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to list images: server returned %s", resp.Status)
	}

	var result imagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
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
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services/%s/replicas",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(serviceName),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("replicas request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("replicas request failed with status %s", resp.Status)
	}

	var result replicasResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode replicas response: %w", err)
	}

	return result.Replicas, nil
}

func (c *DeploymentClient) DescribeReplica(
	ctx context.Context,
	orgId,
	projectId,
	replicaName string,
) (*ReplicaStatus, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services/replicas/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(replicaName),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("describe replica request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("describe replica request failed with status %s", resp.Status)
	}

	var result ReplicaStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode replica status response: %w", err)
	}

	return &result, nil
}

type VectorStoreInfo struct {
	VectorStoreName string `json:"instanceName"`
	Status          string `json:"status"`
	SecretName      string `json:"secretName,omitempty"`
}

type CreateVectorStoreBody struct {
	Resources VectorStoreResources `json:"resources"`
	Storage   VectorStoreStorage   `json:"storage"`
	HA        bool                 `json:"ha"`
	Backups   bool                 `json:"backups"`
	StackId   string               `json:"stackId,omitempty"`
}

type VectorStoreResources struct {
	CPU    int     `json:"cpu"    yaml:"cpu"`
	Memory float64 `json:"memory" yaml:"memory"`
}

type VectorStoreStorage struct {
	Size            int  `json:"size"            yaml:"size"`
	AutoResize      bool `json:"autoResize"      yaml:"autoResize"`
	AutoResizeLimit int  `json:"autoResizeLimit" yaml:"autoResizeLimit"`
}

type VectorStoreBackupConfig struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"startTime"`
}

type DescribeVectorStoreResponse struct {
	VectorStoreName string                  `json:"vectorStoreName"`
	Status          string                  `json:"status"`
	EngineVersion   string                  `json:"engineVersion"`
	CreatedAt       string                  `json:"createdAt"`
	Resources       VectorStoreResources    `json:"resources"`
	Storage         VectorStoreStorage      `json:"storage"`
	HA              bool                    `json:"ha"`
	Backups         VectorStoreBackupConfig `json:"backups"`
	SecretName      string                  `json:"secretName,omitempty"`
}

func (c *DeploymentClient) CreateVectorStore(
	ctx context.Context,
	orgId,
	projectId,
	vectorStoreName string,
	req CreateVectorStoreBody,
) (string, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/vector-stores/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(vectorStoreName),
	)
	reqHTTP, err := c.newRequest(ctx, http.MethodPost, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	resp, err := c.do(reqHTTP)
	if err != nil {
		return "", fmt.Errorf("vector store creation request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("vector store creation failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

func (c *DeploymentClient) ListVectorStores(
	ctx context.Context,
	orgId,
	projectId string,
	stackId string,
) ([]VectorStoreInfo, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/vector-stores",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
	)
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
		return nil, fmt.Errorf("vector store list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to list vector stores: server returned %s", resp.Status)
	}

	var result vectorStoresResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode vector stores response: %w", err)
	}

	return result.VectorStores, nil
}

func (c *DeploymentClient) DescribeVectorStore(
	ctx context.Context,
	orgId,
	projectId,
	vectorStoreName string,
) (*DescribeVectorStoreResponse, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/vector-stores/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(vectorStoreName),
	)
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("describe vector store request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("failed to describe vector store: server returned %s", resp.Status)
	}

	var result DescribeVectorStoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode vector store response: %w", err)
	}

	return &result, nil
}

func (c *DeploymentClient) DeleteVectorStore(
	ctx context.Context,
	orgId,
	projectId,
	vectorStoreName string,
) (string, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/vector-stores/%s",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(vectorStoreName),
	)
	req, err := c.newRequest(ctx, http.MethodDelete, path)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("vector store deletion request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	serverMessage := ExtractServerMessage(respBody)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if serverMessage != "" {
			return "", fmt.Errorf("%s", serverMessage)
		}
		return "", fmt.Errorf("vector store deletion failed with status %s", resp.Status)
	}

	return serverMessage, nil
}

type LogsOptions struct {
	Follow    bool
	Since     string
	StartTime string
	EndTime   string
}

// LogsResponse wraps the log body stream together with metadata returned by the server.
type LogsResponse struct {
	Body      io.ReadCloser
	Since     string // effective start timestamp (from X-Log-Since header)
	Truncated bool   // true when the server hit the entry limit (X-Log-Truncated header)
	Empty     bool   // true when there are no logs (X-Log-Empty header)
}

func (c *DeploymentClient) GetReplicaLogs(
	ctx context.Context,
	orgId,
	projectId,
	replicaName string,
	opts LogsOptions,
) (*LogsResponse, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services/replicas/%s/logs",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(replicaName),
	)
	return c.fetchLogs(ctx, path, opts)
}

func (c *DeploymentClient) GetServiceLogs(
	ctx context.Context,
	orgId,
	projectId,
	serviceName string,
	opts LogsOptions,
) (*LogsResponse, error) {
	path := fmt.Sprintf(
		"/v1/organizations/%s/projects/%s/services/%s/logs",
		url.PathEscape(orgId),
		url.PathEscape(projectId),
		url.PathEscape(serviceName),
	)
	return c.fetchLogs(ctx, path, opts)
}

func (c *DeploymentClient) fetchLogs(
	ctx context.Context,
	path string,
	opts LogsOptions,
) (*LogsResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	if opts.Follow {
		q.Set("follow", "true")
	}
	if opts.Since != "" {
		q.Set("since", opts.Since)
	}
	if opts.StartTime != "" {
		q.Set("start-time", opts.StartTime)
	}
	if opts.EndTime != "" {
		q.Set("end-time", opts.EndTime)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("logs request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		msg := ExtractServerMessage(respBody)
		if msg != "" {
			return nil, fmt.Errorf("%s", msg)
		}
		return nil, fmt.Errorf("logs request failed with status %s", resp.Status)
	}

	return &LogsResponse{
		Body:      resp.Body,
		Since:     resp.Header.Get("X-Log-Since"),
		Truncated: resp.Header.Get("X-Log-Truncated") == "true",
		Empty:     resp.Header.Get("X-Log-Empty") == "true",
	}, nil
}

// SyncResult holds the outcome of a sync operation.
type SyncResult struct {
	Created []string
	Updated []string
	Deleted []string
}

// SyncServices creates, updates, and deletes services to match the desired state.
// It is scoped to the given stackId.
func (c *DeploymentClient) SyncServices(
	ctx context.Context,
	orgId,
	projectId,
	stackId string,
	desired map[string]CreateServiceBody,
) (*SyncResult, error) {
	existing, err := c.ListServices(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	existingByName := make(map[string]ServiceOutput)
	for _, svc := range existing {
		existingByName[svc.Name] = svc
	}

	result := &SyncResult{
		Created: []string{},
		Updated: []string{},
		Deleted: []string{},
	}

	desiredNames := make([]string, 0, len(desired))
	for name := range desired {
		desiredNames = append(desiredNames, name)
	}
	sort.Strings(desiredNames)

	for _, name := range desiredNames {
		body := desired[name]
		if _, exists := existingByName[name]; !exists {
			_, err := c.CreateService(ctx, orgId, projectId, name, body)
			if err != nil {
				return result, fmt.Errorf("failed to create service %q: %w", name, err)
			}
			result.Created = append(result.Created, name)
		} else {
			_, err := c.UpdateService(ctx, orgId, projectId, name, body)
			if err != nil {
				return result, fmt.Errorf("failed to update service %q: %w", name, err)
			}
			result.Updated = append(result.Updated, name)
		}
	}

	existingNames := make([]string, 0, len(existingByName))
	for name := range existingByName {
		existingNames = append(existingNames, name)
	}
	sort.Strings(existingNames)

	for _, name := range existingNames {
		if _, ok := desired[name]; !ok {
			_, err := c.DeleteService(ctx, orgId, projectId, name)
			if err != nil {
				return result, fmt.Errorf("failed to delete service %q: %w", name, err)
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}

// SyncVectorStores creates and deletes vector stores to match the desired state.
// Updates are not supported (no update endpoint). Scoped to the given stackId.
func (c *DeploymentClient) SyncVectorStores(
	ctx context.Context,
	orgId,
	projectId,
	stackId string,
	desired map[string]CreateVectorStoreBody,
) (*SyncResult, error) {
	existing, err := c.ListVectorStores(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list vector stores: %w", err)
	}

	existingByName := make(map[string]VectorStoreInfo)
	for _, vs := range existing {
		existingByName[vs.VectorStoreName] = vs
	}

	result := &SyncResult{
		Created: []string{},
		Updated: []string{},
		Deleted: []string{},
	}

	desiredNames := make([]string, 0, len(desired))
	for name := range desired {
		desiredNames = append(desiredNames, name)
	}
	sort.Strings(desiredNames)

	for _, name := range desiredNames {
		body := desired[name]
		if _, exists := existingByName[name]; !exists {
			_, err := c.CreateVectorStore(ctx, orgId, projectId, name, body)
			if err != nil {
				return result, fmt.Errorf("failed to create vector store %q: %w", name, err)
			}
			result.Created = append(result.Created, name)
		}
	}

	existingNames := make([]string, 0, len(existingByName))
	for name := range existingByName {
		existingNames = append(existingNames, name)
	}
	sort.Strings(existingNames)

	for _, name := range existingNames {
		if _, ok := desired[name]; !ok {
			_, err := c.DeleteVectorStore(ctx, orgId, projectId, name)
			if err != nil {
				return result, fmt.Errorf("failed to delete vector store %q: %w", name, err)
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}
