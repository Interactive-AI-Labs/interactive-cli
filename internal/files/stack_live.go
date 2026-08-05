package files

import (
	"context"
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"gopkg.in/yaml.v3"
)

func FetchLiveStack(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId, projectId, stackId string,
) (*StackConfig, error) {
	cfg := &StackConfig{
		StackId:   stackId,
		Services:  make(map[string]ServiceConfig),
		Agents:    make(map[string]AgentConfig),
		Databases: make(map[string]DatabaseConfig),
	}

	svcs, err := deployClient.ListServices(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	for _, svc := range svcs {
		desc, err := deployClient.DescribeService(ctx, orgId, projectId, svc.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe service %q: %w", svc.Name, err)
		}
		cfg.Services[svc.Name] = ServiceConfigFromDescribe(desc)
	}

	agents, err := deployClient.ListAgents(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	for _, a := range agents {
		desc, err := deployClient.DescribeAgent(ctx, orgId, projectId, a.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe agent %q: %w", a.Name, err)
		}
		cfg.Agents[a.Name] = AgentConfigFromDescribe(desc)
	}

	dbs, err := deployClient.ListDatabases(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	for _, db := range dbs {
		desc, err := deployClient.DescribeDatabase(ctx, orgId, projectId, db.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe database %q: %w", db.Name, err)
		}
		cfg.Databases[db.Name] = DatabaseConfigFromDescribe(desc)
	}

	return cfg, nil
}

func MarshalStackConfig(cfg *StackConfig) ([]byte, error) {
	return yaml.Marshal(cfg)
}
