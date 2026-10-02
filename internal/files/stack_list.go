package files

import (
	"context"
	"fmt"
	"sort"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

// StackInfo holds summary info for a stack discovered from live resources.
type StackInfo struct {
	StackID       string `json:"stackId"`
	ServiceCount  int    `json:"serviceCount"`
	AgentCount    int    `json:"agentCount"`
	DatabaseCount int    `json:"databaseCount"`
	McpCount      int    `json:"mcpCount"`
	JobCount      int    `json:"jobCount"`
}

// ListStacks discovers stacks and their resource counts from live services,
// agents, databases, mcps, and jobs. Resources without a stackId are skipped.
func ListStacks(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	apiClient *platform.APIClient,
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

	mcps, err := deployClient.ListMcps(ctx, orgId, projectId, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list mcps: %w", err)
	}
	// MCPs expose stackId in the list response, so no describe call is needed.
	for _, mcp := range mcps {
		if mcp.StackId == "" {
			continue
		}
		s, ok := stacks[mcp.StackId]
		if !ok {
			s = &StackInfo{StackID: mcp.StackId}
			stacks[mcp.StackId] = s
		}
		s.McpCount++
	}
	remoteMcps, err := apiClient.ListRemoteMcps(ctx, orgId, projectId, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list remote mcps: %w", err)
	}
	// Counted even when a stack file cannot express them: list is a census of what carries the stack.
	for _, mcp := range remoteMcps {
		if mcp.StackId == nil || *mcp.StackId == "" {
			continue
		}
		s, ok := stacks[*mcp.StackId]
		if !ok {
			s = &StackInfo{StackID: *mcp.StackId}
			stacks[*mcp.StackId] = s
		}
		s.McpCount++
	}

	jobs, err := deployClient.ListJobs(ctx, orgId, projectId, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	for _, job := range jobs {
		desc, err := deployClient.DescribeJob(ctx, orgId, projectId, job.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe job %q: %w", job.Name, err)
		}
		if desc.StackId == "" {
			continue
		}
		s, ok := stacks[desc.StackId]
		if !ok {
			s = &StackInfo{StackID: desc.StackId}
			stacks[desc.StackId] = s
		}
		s.JobCount++
	}

	result := make([]StackInfo, 0, len(stacks))
	for _, s := range stacks {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].StackID < result[j].StackID
	})
	return result, nil
}
