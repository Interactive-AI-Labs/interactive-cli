package cmd

import (
	"context"
	"fmt"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	files "github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
)

type projectContext struct {
	apiClient *clients.APIClient
	projectId string
}

func resolveProjectContext(ctx context.Context, org, project string) (*projectContext, error) {
	cfg, err := files.LoadStackConfig(cfgFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, apiKey, cookies)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	sess := session.NewSession(cfgDirName)

	orgName, err := sess.ResolveOrganization(cfg.Organization, org)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve organization: %w", err)
	}

	projectName, err := sess.ResolveProject(cfg.Project, project)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project: %w", err)
	}

	_, projectId, err := apiClient.GetProjectId(ctx, orgName, projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project %q: %w", projectName, err)
	}

	return &projectContext{
		apiClient: apiClient,
		projectId: projectId,
	}, nil
}
