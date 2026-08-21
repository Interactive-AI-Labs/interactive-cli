package sync

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/preflight"
)

func AllowDeleteResource(allowed []string, resource string) bool {
	for _, a := range allowed {
		if strings.EqualFold(a, resource) || strings.EqualFold(a, "all") {
			return true
		}
	}
	return false
}

type Result struct {
	Created   []string
	Updated   []string
	Deleted   []string
	Protected []string // would be deleted but deletion was not allowed
}

type Options struct {
	AllowDelete bool
	DryRun      bool
}

func HasServices(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
) (bool, error) {
	existing, err := deployClient.ListServices(ctx, orgId, projectId, stackId)
	if err != nil {
		return false, fmt.Errorf("failed to list services: %w", err)
	}

	return len(existing) > 0, nil
}

func HasAgents(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
) (bool, error) {
	existing, err := deployClient.ListAgents(ctx, orgId, projectId, stackId)
	if err != nil {
		return false, fmt.Errorf("failed to list agents: %w", err)
	}

	return len(existing) > 0, nil
}

func HasDatabases(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
) (bool, error) {
	existing, err := deployClient.ListDatabases(ctx, orgId, projectId, stackId)
	if err != nil {
		return false, fmt.Errorf("failed to list databases: %w", err)
	}

	return len(existing) > 0, nil
}

func HasMcps(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
) (bool, error) {
	existing, err := deployClient.ListMcps(ctx, orgId, projectId, stackId)
	if err != nil {
		return false, fmt.Errorf("failed to list mcps: %w", err)
	}

	return len(existing) > 0, nil
}

func PrintResult(
	out io.Writer,
	label string,
	result *Result,
	err error,
) error {
	if err != nil {
		if result != nil {
			output.PrintSyncResult(
				out,
				label+" (partial)",
				result.Created,
				result.Updated,
				result.Deleted,
			)
		}
		return err
	}
	output.PrintSyncResult(
		out,
		label,
		result.Created,
		result.Updated,
		result.Deleted,
	)
	if len(result.Protected) > 0 {
		fmt.Fprintf(
			out,
			"\nProtected %s (not deleted): %s\n"+
				"Use --allow-delete=%s to delete them.\n",
			label,
			strings.Join(result.Protected, ", "),
			label,
		)
	}
	return nil
}

func PrintPlan(out io.Writer, label string, result *Result) {
	if len(result.Created) > 0 {
		fmt.Fprintf(out, "Would create %s: %s\n", label, strings.Join(result.Created, ", "))
	}
	if len(result.Updated) > 0 {
		fmt.Fprintf(out, "Would update %s: %s\n", label, strings.Join(result.Updated, ", "))
	}
	if len(result.Deleted) > 0 {
		fmt.Fprintf(out, "Would delete %s: %s\n", label, strings.Join(result.Deleted, ", "))
	}
	if len(result.Protected) > 0 {
		fmt.Fprintf(
			out,
			"Would refuse to delete %s: %s (a config that omits a resource looks identical to a stale one — pass --allow-delete=%s to delete)\n",
			label,
			strings.Join(result.Protected, ", "),
			label,
		)
	}
	if len(result.Created) == 0 && len(result.Updated) == 0 &&
		len(result.Deleted) == 0 && len(result.Protected) == 0 {
		fmt.Fprintf(out, "No changes required; %s already match config.\n", label)
	}
}

func Services(
	ctx context.Context,
	warnW io.Writer,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]deployment.CreateServiceBody,
	opts Options,
) (*Result, error) {
	existing, err := deployClient.ListServices(
		ctx, orgId, projectId, stackId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	existingByName := make(map[string]deployment.ServiceOutput)
	for _, svc := range existing {
		existingByName[svc.Name] = svc
	}

	return syncResources(
		warnW,
		existingByName,
		desired,
		opts,
		resourceOps[deployment.ServiceOutput, deployment.CreateServiceBody]{
			resource:  "service",
			allowFlag: "services",
			create: func(name string, body deployment.CreateServiceBody) error {
				_, err := deployClient.CreateService(ctx, orgId, projectId, name, body)
				return err
			},
			update: func(name string, body deployment.CreateServiceBody) error {
				_, err := deployClient.PutService(ctx, orgId, projectId, name, body)
				return err
			},
			delete: func(name string) error {
				_, err := deployClient.DeleteService(ctx, orgId, projectId, name)
				return err
			},
			banner: func(w io.Writer, svc deployment.ServiceOutput) {
				preflight.PrintUpdateBanner(w, "service "+svc.Name, svc.Revision, svc.Updated)
			},
		},
	)
}

func Agents(
	ctx context.Context,
	warnW io.Writer,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]deployment.CreateAgentBody,
	opts Options,
) (*Result, error) {
	existing, err := deployClient.ListAgents(
		ctx, orgId, projectId, stackId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	existingByName := make(map[string]deployment.AgentOutput)
	for _, a := range existing {
		existingByName[a.Name] = a
	}

	return syncResources(
		warnW,
		existingByName,
		desired,
		opts,
		resourceOps[deployment.AgentOutput, deployment.CreateAgentBody]{
			resource:  "agent",
			allowFlag: "agents",
			create: func(name string, body deployment.CreateAgentBody) error {
				_, err := deployClient.CreateAgent(ctx, orgId, projectId, name, body)
				return err
			},
			update: func(name string, body deployment.CreateAgentBody) error {
				_, err := deployClient.PutAgent(ctx, orgId, projectId, name, body)
				return err
			},
			delete: func(name string) error {
				_, err := deployClient.DeleteAgent(ctx, orgId, projectId, name)
				return err
			},
			banner: func(w io.Writer, a deployment.AgentOutput) {
				preflight.PrintUpdateBanner(w, "agent "+a.Name, a.Revision, a.Updated)
			},
		},
	)
}

func Databases(
	ctx context.Context,
	warnW io.Writer,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]deployment.CreateDatabaseBody,
	opts Options,
) (*Result, error) {
	existing, err := deployClient.ListDatabases(
		ctx, orgId, projectId, stackId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	existingByName := make(map[string]deployment.DatabaseOutput)
	for _, db := range existing {
		existingByName[db.Name] = db
	}

	return syncResources(
		warnW,
		existingByName,
		desired,
		opts,
		resourceOps[deployment.DatabaseOutput, deployment.CreateDatabaseBody]{
			resource:  "database",
			allowFlag: "databases",
			create: func(name string, body deployment.CreateDatabaseBody) error {
				_, err := deployClient.CreateDatabase(ctx, orgId, projectId, name, body)
				return err
			},
			update: func(name string, body deployment.CreateDatabaseBody) error {
				_, err := deployClient.PutDatabase(ctx, orgId, projectId, name, body)
				return err
			},
			delete: func(name string) error {
				_, err := deployClient.DeleteDatabase(ctx, orgId, projectId, name)
				return err
			},
		},
	)
}

func Mcps(
	ctx context.Context,
	warnW io.Writer,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]deployment.CreateMcpBody,
	opts Options,
) (*Result, error) {
	existing, err := deployClient.ListMcps(
		ctx, orgId, projectId, stackId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list mcps: %w", err)
	}

	existingByName := make(map[string]deployment.McpOutput)
	for _, mcp := range existing {
		existingByName[mcp.Name] = mcp
	}

	return syncResources(
		warnW,
		existingByName,
		desired,
		opts,
		resourceOps[deployment.McpOutput, deployment.CreateMcpBody]{
			resource:  "mcp",
			allowFlag: "mcps",
			create: func(name string, body deployment.CreateMcpBody) error {
				_, err := deployClient.CreateMcp(ctx, orgId, projectId, name, body)
				return err
			},
			update: func(name string, body deployment.CreateMcpBody) error {
				authType := body.Auth.Type
				if authType == "" {
					authType = existingByName[name].Auth.Type
				}
				if body.Auth.Credential == "" && authType != "" &&
					!strings.EqualFold(authType, "none") {
					return fmt.Errorf(
						"auth.credential is required to update mcp %q; stack get never exports credentials",
						name,
					)
				}
				_, err := deployClient.PutMcp(ctx, orgId, projectId, name, body)
				return err
			},
			delete: func(name string) error {
				_, err := deployClient.DeleteMcp(ctx, orgId, projectId, name, false)
				if err != nil {
					return fmt.Errorf("detach agents before deleting mcp %q: %w", name, err)
				}
				return nil
			},
			banner: func(w io.Writer, mcp deployment.McpOutput) {
				preflight.PrintUpdateBanner(w, "mcp "+mcp.Name, mcp.Revision, mcp.Updated)
			},
		},
	)
}

type resourceOps[E, B any] struct {
	resource  string
	allowFlag string
	create    func(name string, body B) error
	update    func(name string, body B) error
	delete    func(name string) error
	banner    func(w io.Writer, existing E)
}

func syncResources[E, B any](
	warnW io.Writer,
	existingByName map[string]E,
	desired map[string]B,
	opts Options,
	ops resourceOps[E, B],
) (*Result, error) {
	result := &Result{}

	var toDelete []string
	for name := range existingByName {
		if _, ok := desired[name]; !ok {
			toDelete = append(toDelete, name)
		}
	}
	sort.Strings(toDelete)
	if !opts.DryRun {
		preflight.PrintSyncDeletions(warnW, ops.resource, ops.allowFlag, toDelete, opts.AllowDelete)
	}
	if !opts.AllowDelete {
		result.Protected = toDelete
		toDelete = nil
	}

	desiredNames := make([]string, 0, len(desired))
	for name := range desired {
		desiredNames = append(desiredNames, name)
	}
	sort.Strings(desiredNames)

	for _, name := range desiredNames {
		body := desired[name]
		if existing, exists := existingByName[name]; !exists {
			if !opts.DryRun {
				if err := ops.create(name, body); err != nil {
					return result, fmt.Errorf(
						"failed to create %s %q: %w", ops.resource, name, err,
					)
				}
			}
			result.Created = append(result.Created, name)
		} else {
			if !opts.DryRun {
				if ops.banner != nil {
					ops.banner(warnW, existing)
				}
				if err := ops.update(name, body); err != nil {
					return result, fmt.Errorf(
						"failed to update %s %q: %w", ops.resource, name, err,
					)
				}
			}
			result.Updated = append(result.Updated, name)
		}
	}

	for _, name := range toDelete {
		if !opts.DryRun {
			if err := ops.delete(name); err != nil {
				return result, fmt.Errorf(
					"failed to delete %s %q: %w", ops.resource, name, err,
				)
			}
		}
		result.Deleted = append(result.Deleted, name)
	}

	return result, nil
}
