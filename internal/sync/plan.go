package sync

import (
	"context"
	"fmt"
	"io"
	"maps"
	"reflect"
	"slices"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

// PlanStack validates replacements without applying them; agents wait for changing MCP dependencies.
func PlanStack(
	ctx context.Context,
	warnW io.Writer,
	deployClient *deployment.DeploymentClient,
	apiClient *platform.APIClient,
	orgId, projectId, stackId string,
	local *files.StackConfig,
	allowMcpDelete bool,
) (*files.StackDiff, error) {
	live, err := files.FetchLiveStack(ctx, deployClient, apiClient, orgId, projectId, stackId)
	if err != nil {
		return nil, err
	}

	hosted, remote := files.McpRequests(local.Mcps, stackId)
	inventory, err := loadMcps(ctx, deployClient, apiClient, orgId, projectId)
	if err != nil {
		return nil, err
	}
	_, remoteByName, err := inventory.forStack(warnW, stackId, hosted, remote)
	if err != nil {
		return nil, err
	}

	plan := &files.StackPlan{StackId: stackId}
	plan.Services, err = planResources(local.Services, live.Services,
		func(name string, cfg files.ServiceConfig) (files.ServiceConfig, bool, error) {
			p, err := deployClient.PlanService(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return files.ServiceConfig{}, false, fmt.Errorf(
					"failed to plan service %q: %w",
					name,
					err,
				)
			}
			return files.ServiceConfigFromDescribe(&p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}
	plan.Databases, err = planResources(local.Databases, live.Databases,
		func(name string, cfg files.DatabaseConfig) (files.DatabaseConfig, bool, error) {
			p, err := deployClient.PlanDatabase(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return files.DatabaseConfig{}, false, fmt.Errorf(
					"failed to plan database %q: %w",
					name,
					err,
				)
			}
			return files.DatabaseConfigFromDescribe(&p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}
	plan.Mcps, err = planResources(local.Mcps, live.Mcps,
		func(name string, cfg files.McpConfig) (files.McpConfig, bool, error) {
			if cfg.Type == deployment.McpTypeRemote {
				return planRemoteMcp(name, remoteByName[name], cfg, stackId)
			}
			p, err := deployClient.PlanMcp(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return files.McpConfig{}, false, fmt.Errorf("failed to plan mcp %q: %w", name, err)
			}
			return files.McpConfigFromDescribe(&p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}

	changingMcps := pendingMcpNames(plan.Mcps, live.Mcps, allowMcpDelete)
	deferred := make(map[string][]string)
	plan.Agents, err = planResources(local.Agents, live.Agents,
		func(name string, cfg files.AgentConfig) (files.AgentConfig, bool, error) {
			if pending := inputs.PendingMcpRefs(cfg.AgentConfig, changingMcps); len(pending) > 0 {
				deferred[name] = pending
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
				return files.AgentConfig{}, false, fmt.Errorf(
					"failed to plan agent %q: %w",
					name,
					err,
				)
			}
			return files.AgentConfigFromDescribe(&p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}
	plan.Agents.Deferred = deferred
	plan.Jobs, err = planResources(local.Jobs, live.Jobs,
		func(name string, cfg files.JobConfig) (files.JobConfig, bool, error) {
			p, err := deployClient.PlanJob(
				ctx,
				orgId,
				projectId,
				name,
				cfg.ToCreateRequest(stackId),
			)
			if err != nil {
				return files.JobConfig{}, false, fmt.Errorf("failed to plan job %q: %w", name, err)
			}
			return files.JobConfigFromRequest(p.Config), p.Changed, nil
		})
	if err != nil {
		return nil, err
	}
	return files.DiffStackConfigs(plan, live), nil
}

// PrintStackPlan summarizes a preview, respecting the sync command's deletion permissions.
func PrintStackPlan(out io.Writer, plan *files.StackDiff, allowDelete []string) {
	if !plan.HasChanges() {
		fmt.Fprintln(out, "No changes required.")
		return
	}
	for _, section := range []struct {
		name string
		diff files.ResourceTypeDiff
	}{
		{"services", plan.Services},
		{"databases", plan.Databases},
		{"mcps", plan.Mcps},
		{"agents", plan.Agents},
		{"jobs", plan.Jobs},
	} {
		result := Result{Created: section.diff.Created, Deferred: section.diff.Deferred}
		for _, change := range section.diff.Updated {
			result.Updated = append(result.Updated, change.Name)
		}
		if AllowDeleteResource(allowDelete, section.name) {
			result.Deleted = section.diff.Deleted
		} else {
			result.Protected = section.diff.Deleted
		}
		PrintPlan(out, section.name, &result)
	}
}

// planRemoteMcp validates the patch without exposing its write-only credential in the plan.
func planRemoteMcp(
	name string,
	live platform.McpSchema,
	cfg files.McpConfig,
	stackId string,
) (files.McpConfig, bool, error) {
	patch, err := remoteMcpPatch(name, live, cfg.ToPlatformCreateRequest(name, stackId))
	cfg.Auth.Credential = ""
	return cfg, len(patch) > 0, err
}

func pendingMcpNames(
	plan files.ResourcePlan[files.McpConfig],
	live map[string]files.McpConfig,
	allowDelete bool,
) []string {
	var names []string
	for name := range plan.Changed {
		// Credential-only remote changes leave the redacted configuration unchanged.
		if plan.Desired[name].Type == deployment.McpTypeRemote &&
			reflect.DeepEqual(plan.Desired[name], live[name]) {
			continue
		}
		names = append(names, name)
	}
	for name := range plan.Desired {
		if _, exists := live[name]; !exists {
			names = append(names, name)
		}
	}
	if allowDelete {
		for name := range live {
			if _, exists := plan.Desired[name]; !exists {
				names = append(names, name)
			}
		}
	}
	slices.Sort(names)
	return names
}

func planResources[T any](
	local, live map[string]T,
	plan func(string, T) (T, bool, error),
) (files.ResourcePlan[T], error) {
	out := files.ResourcePlan[T]{
		Desired: make(map[string]T, len(local)),
		Changed: make(map[string]bool),
	}
	for _, name := range slices.Sorted(maps.Keys(local)) {
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
