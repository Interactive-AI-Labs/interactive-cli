package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// getOrgId looks up an organization by name using the
// /api/v1/session/organizations endpoint and returns its ID.
func GetOrgId(
	ctx context.Context,
	apiHostname string,
	cfgDirName string,
	sessionFileName string,
	orgName string,
	timeout time.Duration,
) (string, error) {
	orgName = strings.TrimSpace(orgName)
	if orgName == "" {
		return "", fmt.Errorf("organization name cannot be empty")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	cookies, err := LoadSessionCookies(cfgDirName, sessionFileName)
	if err != nil {
		return "", fmt.Errorf("failed to load session: %w", err)
	}
	if len(cookies) == 0 {
		return "", fmt.Errorf("no active session; please log in before selecting an organization")
	}

	base := strings.TrimRight(apiHostname, "/")
	url := base + "/api/v1/session/organizations"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create organizations request: %w", err)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("organizations request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(body); msg != "" {
			return "", fmt.Errorf("failed to list organizations: %s", msg)
		}
		return "", fmt.Errorf("failed to list organizations: server returned %s", resp.Status)
	}

	var payload struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("failed to decode organizations response: %w", err)
	}

	if len(payload.Organizations) == 0 {
		return "", fmt.Errorf("no organizations found in your session")
	}

	// Case-insensitive match on name.
	var matched []struct {
		ID   string
		Name string
	}
	for _, org := range payload.Organizations {
		if strings.EqualFold(org.Name, orgName) {
			matched = append(matched, struct {
				ID   string
				Name string
			}{
				ID:   org.ID,
				Name: org.Name,
			})
		}
	}

	if len(matched) == 0 {
		return "", fmt.Errorf("organization %q not found", orgName)
	}
	if len(matched) > 1 {
		return "", fmt.Errorf("organization name %q is ambiguous; please use a unique name", orgName)
	}

	return matched[0].ID, nil
}

// GetProjectId looks up a project by name within an organization,
// using the organizations and projects session APIs.
func GetProjectId(
	ctx context.Context,
	apiHostname string,
	cfgDirName string,
	sessionFileName string,
	orgName string,
	projectName string,
	timeout time.Duration,
) (string, string, error) {
	orgName = strings.TrimSpace(orgName)
	projectName = strings.TrimSpace(projectName)

	if orgName == "" {
		return "", "", fmt.Errorf("organization name cannot be empty")
	}
	if projectName == "" {
		return "", "", fmt.Errorf("project name cannot be empty")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	orgID, err := GetOrgId(ctx, apiHostname, cfgDirName, sessionFileName, orgName, timeout)
	if err != nil {
		return "", "", err
	}

	cookies, err := LoadSessionCookies(cfgDirName, sessionFileName)
	if err != nil {
		return "", "", fmt.Errorf("failed to load session: %w", err)
	}
	if len(cookies) == 0 {
		return "", "", fmt.Errorf("no active session; please log in before selecting a project")
	}

	base := strings.TrimRight(apiHostname, "/")
	url := fmt.Sprintf("%s/api/v1/session/organizations/%s/projects", base, orgID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create projects request: %w", err)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("projects request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if msg := ExtractServerMessage(body); msg != "" {
			return "", "", fmt.Errorf("failed to list projects: %s", msg)
		}
		return "", "", fmt.Errorf("failed to list projects: server returned %s", resp.Status)
	}

	var payload struct {
		Projects []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"projects"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", fmt.Errorf("failed to decode projects response: %w", err)
	}

	if len(payload.Projects) == 0 {
		return "", "", fmt.Errorf("no projects found in organization %q", orgName)
	}

	var matched []struct {
		ID   string
		Name string
	}
	for _, proj := range payload.Projects {
		if strings.EqualFold(proj.Name, projectName) {
			matched = append(matched, struct {
				ID   string
				Name string
			}{
				ID:   proj.ID,
				Name: proj.Name,
			})
		}
	}

	if len(matched) == 0 {
		return "", "", fmt.Errorf("project %q not found in organization %q", projectName, orgName)
	}
	if len(matched) > 1 {
		return "", "", fmt.Errorf("project name %q is ambiguous in organization %q; please use a unique name", projectName, orgName)
	}

	return orgID, matched[0].ID, nil
}
