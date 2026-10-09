package files

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

// StackPlan holds server predictions and agent plans deferred by changing MCPs.
type StackPlan struct {
	StackId   string
	Services  ResourcePlan[ServiceConfig]
	Agents    ResourcePlan[AgentConfig]
	Databases ResourcePlan[DatabaseConfig]
	Mcps      ResourcePlan[McpConfig]
	Jobs      ResourcePlan[JobConfig]
}

// ResourcePlan holds planned configurations, changed names, and dependencies that defer a plan.
type ResourcePlan[T any] struct {
	Desired  map[string]T
	Changed  map[string]bool
	Deferred map[string][]string // Deferred resources keep their local config, not a server prediction.
}

// PlanStack asks the server what replacing each live resource with the local config would do.
func PlanStack(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	orgId, projectId, stackId string,
	local, live *StackConfig,
) (*StackPlan, error) {
	plan := &StackPlan{StackId: stackId}
	var err error
	plan.Services, err = planResources(local.Services, live.Services,
		func(name string, cfg ServiceConfig) (ServiceConfig, bool, error) {
			p, err := deployClient.PlanService(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return ServiceConfig{}, false, fmt.Errorf(
					"failed to plan service %q: %w",
					name,
					err,
				)
			}
			return ServiceConfigFromDescribe(&p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}

	plan.Databases, err = planResources(local.Databases, live.Databases,
		func(name string, cfg DatabaseConfig) (DatabaseConfig, bool, error) {
			p, err := deployClient.PlanDatabase(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return DatabaseConfig{}, false, fmt.Errorf(
					"failed to plan database %q: %w",
					name,
					err,
				)
			}
			return DatabaseConfigFromDescribe(&p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}

	plan.Mcps, err = planResources(local.Mcps, live.Mcps,
		func(name string, cfg McpConfig) (McpConfig, bool, error) {
			liveMcp := live.Mcps[name]
			// The platform owns remote mcps and has no dry run: the file is compared with what it
			// reports, which never includes the credential.
			if cfg.Type == deployment.McpTypeRemote || liveMcp.Type == deployment.McpTypeRemote {
				cfg.Auth.Credential = ""
				return cfg, len(diffFields(liveMcp, cfg)) > 0, nil
			}
			p, err := deployClient.PlanMcp(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return McpConfig{}, false, fmt.Errorf("failed to plan mcp %q: %w", name, err)
			}
			return McpConfigFromDescribe(&p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}

	changingMcps := slices.Collect(maps.Keys(plan.Mcps.Changed))
	for name := range local.Mcps {
		if _, exists := live.Mcps[name]; !exists {
			changingMcps = append(changingMcps, name)
		}
	}
	for name := range live.Mcps {
		if _, exists := local.Mcps[name]; !exists {
			changingMcps = append(changingMcps, name)
		}
	}

	deferredAgents := make(map[string][]string)
	plan.Agents, err = planResources(local.Agents, live.Agents,
		func(name string, cfg AgentConfig) (AgentConfig, bool, error) {
			if pending := inputs.PendingMcpRefs(cfg.AgentConfig, changingMcps); len(pending) > 0 {
				deferredAgents[name] = pending
				return cfg, false, nil
			}

			p, err := deployClient.PlanAgent(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return AgentConfig{}, false, fmt.Errorf("failed to plan agent %q: %w", name, err)
			}
			return AgentConfigFromDescribe(&p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}
	plan.Agents.Deferred = deferredAgents

	plan.Jobs, err = planResources(local.Jobs, live.Jobs,
		func(name string, cfg JobConfig) (JobConfig, bool, error) {
			p, err := deployClient.PlanJob(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return JobConfig{}, false, fmt.Errorf("failed to plan job %q: %w", name, err)
			}
			return JobConfigFromRequest(p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}
	return plan, nil
}

// planResources plans each local resource that exists live and keeps the others as written.
func planResources[T any](
	local, live map[string]T,
	plan func(name string, cfg T) (T, bool, error),
) (ResourcePlan[T], error) {
	out := ResourcePlan[T]{Desired: make(map[string]T, len(local)), Changed: make(map[string]bool)}
	names := make([]string, 0, len(local))
	for name := range local {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, exists := live[name]; !exists {
			out.Desired[name] = local[name]
			continue
		}
		desired, changed, err := plan(name, local[name])
		if err != nil {
			return out, err
		}
		out.Desired[name] = desired
		if changed {
			out.Changed[name] = true
		}
	}
	return out, nil
}
