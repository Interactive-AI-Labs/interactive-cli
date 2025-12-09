package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type ResourceRequirements struct {
	Memory string `json:"memory" yaml:"memory"`
	CPU    string `json:"cpu" yaml:"cpu"`
}

type Resources struct {
	Requests ResourceRequirements `json:"requests" yaml:"requests"`
	Limits   ResourceRequirements `json:"limits" yaml:"limits"`
}

type ImageSpec struct {
	Type       string `json:"type" yaml:"type"`
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty"`
	Name       string `json:"name" yaml:"name"`
	Tag        string `json:"tag" yaml:"tag"`
}

type EnvVar struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type SecretRef struct {
	SecretName string `json:"secretName" yaml:"secretName"`
}

type CreateServiceBody struct {
	ServicePort int               `json:"servicePort"`
	Image       ImageSpec         `json:"image"`
	Resources   Resources         `json:"resources"`
	Env         []EnvVar          `json:"env,omitempty"`
	SecretRefs  []SecretRef       `json:"secretRefs,omitempty"`
	Endpoint    bool              `json:"endpoint,omitempty"`
	Hostname    string            `json:"hostname,omitempty"`
	Replicas    int               `json:"replicas"`
	Labels      map[string]string `json:"labels,omitempty"`
}

func CreateService(
	ctx context.Context,
	hostname string,
	timeout time.Duration,
	cookies []*http.Cookie,
	orgID,
	projectID string,
	serviceName string,
	req CreateServiceBody,
) (string, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	u, err := url.Parse(hostname)
	if err != nil {
		return "", fmt.Errorf("failed to parse deployment service URL: %w", err)
	}
	u.Path = fmt.Sprintf("/v1/organizations/%s/projects/%s/services/%s", orgID, projectID, serviceName)

	reqHTTP, err := NewRequestWCookies(ctx, http.MethodPost, u.String(), bodyBytes, cookies)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(reqHTTP)
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

func UpdateService(
	ctx context.Context,
	hostname string,
	timeout time.Duration,
	cookies []*http.Cookie,
	orgID,
	projectID string,
	serviceName string,
	req CreateServiceBody,
) (string, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to encode request body: %w", err)
	}

	u, err := url.Parse(hostname)
	if err != nil {
		return "", fmt.Errorf("failed to parse deployment service URL: %w", err)
	}
	u.Path = fmt.Sprintf("/v1/organizations/%s/projects/%s/services/%s", orgID, projectID, serviceName)

	reqHTTP, err := NewRequestWCookies(ctx, http.MethodPut, u.String(), bodyBytes, cookies)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(reqHTTP)
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

func DeleteService(
	ctx context.Context,
	hostname string,
	timeout time.Duration,
	cookies []*http.Cookie,
	orgID,
	projectID string,
	serviceName string,
) (string, error) {
	u, err := url.Parse(hostname)
	if err != nil {
		return "", fmt.Errorf("failed to parse deployment service URL: %w", err)
	}
	u.Path = fmt.Sprintf("/v1/organizations/%s/projects/%s/services/%s", orgID, projectID, serviceName)

	reqHTTP, err := NewRequestWCookies(ctx, http.MethodDelete, u.String(), nil, cookies)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(reqHTTP)
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

type ListServicesResponse struct {
	Services []ServiceOutput `json:"services"`
}

type ServiceOutput struct {
	Name      string `json:"name"`
	ProjectId string `json:"projectId"`
	Revision  int    `json:"revision"`
	Status    string `json:"status"`
	Updated   string `json:"updated,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
}

func ListServices(
	ctx context.Context,
	hostname string,
	timeout time.Duration,
	cookies []*http.Cookie,
	orgID,
	projectID string,
	stackId string,
) ([]ServiceOutput, error) {
	u, err := url.Parse(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment service URL: %w", err)
	}
	u.Path = fmt.Sprintf("/v1/organizations/%s/projects/%s/services", orgID, projectID)

	if stackId != "" {
		q := u.Query()
		q.Set("stackId", stackId)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for _, cookie := range cookies {
		if cookie != nil {
			req.AddCookie(cookie)
		}
	}

	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(req)
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

	var result ListServicesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode services response: %w", err)
	}

	return result.Services, nil
}

type SyncResult struct {
	Created []string
	Updated []string
	Deleted []string
}

func SyncServices(
	ctx context.Context,
	hostname string,
	timeout time.Duration,
	cookies []*http.Cookie,
	orgID,
	projectID string,
	cfg *StackConfig,
) (*SyncResult, error) {
	existing, err := ListServices(ctx, hostname, timeout, cookies, orgID, projectID, cfg.StackID)
	if err != nil {
		return nil, err
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

	for name, svcCfg := range cfg.Services {
		req := svcCfg.ToCreateRequest(cfg.StackID)

		if _, exists := existingByName[name]; !exists {
			_, err := CreateService(ctx, hostname, timeout, cookies, orgID, projectID, name, req)
			if err != nil {
				return nil, err
			}
			result.Created = append(result.Created, name)
		} else {
			_, err := UpdateService(ctx, hostname, timeout, cookies, orgID, projectID, name, req)
			if err != nil {
				return nil, err
			}
			result.Updated = append(result.Updated, name)
		}
	}

	for name := range existingByName {
		if _, desired := cfg.Services[name]; !desired {
			_, err := DeleteService(ctx, hostname, timeout, cookies, orgID, projectID, name)
			if err != nil {
				return nil, err
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}
