package cmd

import (
	"context"
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/session"
)

type projectContext struct {
	orgId       string
	orgName     string
	projectId   string
	projectName string
}

func resolveProject(
	ctx context.Context,
	org, project string,
) (*projectContext, *clients.APIClient, *clients.DeploymentClient, error) {
	cfg, err := files.LoadStackConfig(cfgFilePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load config file: %w", err)
	}

	cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load session: %w", err)
	}

	apiClient, err := clients.NewAPIClient(hostname, defaultHTTPTimeout, token, apiKey, cookies)
	if err != nil {
		return nil, nil, nil, err
	}

	deployClient, err := clients.NewDeploymentClient(
		deploymentHostname,
		defaultHTTPTimeout,
		token,
		apiKey,
		cookies,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	sess := session.NewSession(cfgDirName)

	orgName, err := sess.ResolveOrganization(cfg.Organization, org)
	if err != nil {
		return nil, nil, nil, err
	}

	projectName, err := sess.ResolveProject(cfg.Project, project)
	if err != nil {
		return nil, nil, nil, err
	}

	orgId, projectId, err := apiClient.GetProjectId(ctx, orgName, projectName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to resolve project %q: %w", projectName, err)
	}

	return &projectContext{
		orgId:       orgId,
		orgName:     orgName,
		projectId:   projectId,
		projectName: projectName,
	}, apiClient, deployClient, nil
}
