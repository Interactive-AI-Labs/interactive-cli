package files

import (
	"context"
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// StackInfo holds summary info for a stack discovered from live resources.
type StackInfo struct {
	StackID        string `json:"stackId"`
	ServiceCount   int    `json:"serviceCount"`
	AgentCount     int    `json:"agentCount"`
	DatabaseCount  int    `json:"databaseCount"`
}

// ListStacks discovers stacks and their resource counts from live services,
// agents, and databases. Resources without a stackId are skipped.
func ListStacks(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId, projectId string,
) ([]StackInfo, error) {
	stacks := make(map[string]*StackInfo)

	svcs, err := deployClient.ListServices(ctx, orgId, projectId, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	for _, svc := range svcs {
		desc, err := deployClient.DescribeService(ctx, orgId, projectId, svc.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe service %q: %w", svc.Name, err)
		}
		if desc.StackId == "" {
			continue
		}
		s, ok := stacks[desc.StackId]
		if !ok {
			s = &StackInfo{StackID: desc.StackId}
			stacks[desc.StackId] = s
		}
		s.ServiceCount++
	}

	agents, err := deployClient.ListAgents(ctx, orgId, projectId, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	for _, a := range agents {
		desc, err := deployClient.DescribeAgent(ctx, orgId, projectId, a.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe agent %q: %w", a.Name, err)
		}
		if desc.StackId == "" {
			continue
		}
		s, ok := stacks[desc.StackId]
		if !ok {
			s = &StackInfo{StackID: desc.StackId}
			stacks[desc.StackId] = s
		}
		s.AgentCount++
	}

	dbs, err := deployClient.ListDatabases(ctx, orgId, projectId, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	for _, db := range dbs {
		desc, err := deployClient.DescribeDatabase(ctx, orgId, projectId, db.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe database %q: %w", db.Name, err)
		}
		if desc.StackId == "" {
			continue
		}
		s, ok := stacks[desc.StackId]
		if !ok {
			s = &StackInfo{StackID: desc.StackId}
			stacks[desc.StackId] = s
		}
		s.DatabaseCount++
	}

	result := make([]StackInfo, 0, len(stacks))
	for _, s := range stacks {
		result = append(result, *s)
	}
	return result, nil
}
