package sync

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
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

// Options controls how a sync applies its computed diff.
type Options struct {
	// AllowDelete permits deleting resources the config file no longer
	// mentions. Off by default: a stale config is indistinguishable from a
	// decommissioning one, so omitted resources are reported as Protected
	// instead of deleted.
	AllowDelete bool
	// DryRun computes the full plan (creates, updates, deletes, protected)
	// without writing anything.
	DryRun bool
}

func HasServices(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
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
	deployClient *clients.DeploymentClient,
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
	deployClient *clients.DeploymentClient,
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

// PrintPlan renders a dry-run Result as the changes a real run would make,
// without implying anything was written. label is the plural noun, which is
// also the --allow-delete value ("services", "agents", "databases").
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

// Services syncs services: creates new ones, updates existing ones, and —
// only with opts.AllowDelete — deletes ones not present in the desired map;
// otherwise those are refused and reported as Protected. Updates go through
// PUT and replace the whole live spec, so the live revision each one
// overwrites is announced on warnW (deploy awareness), and the deletion
// decision is announced there before any service write happens.
func Services(
	ctx context.Context,
	warnW io.Writer,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]clients.CreateServiceBody,
	opts Options,
) (*Result, error) {
	existing, err := deployClient.ListServices(
		ctx, orgId, projectId, stackId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	existingByName := make(map[string]clients.ServiceOutput)
	for _, svc := range existing {
		existingByName[svc.Name] = svc
	}

	return syncResources(
		ctx,
		warnW,
		existingByName,
		desired,
		opts,
		resourceOps[clients.ServiceOutput, clients.CreateServiceBody]{
			resource:  "service",
			allowFlag: "services",
			create: func(name string, body clients.CreateServiceBody) error {
				_, err := deployClient.CreateService(ctx, orgId, projectId, name, body)
				return err
			},
			update: func(name string, body clients.CreateServiceBody) error {
				_, err := deployClient.PutService(ctx, orgId, projectId, name, body)
				return err
			},
			delete: func(name string) error {
				_, err := deployClient.DeleteService(ctx, orgId, projectId, name)
				return err
			},
			banner: func(w io.Writer, svc clients.ServiceOutput) {
				preflight.PrintUpdateBanner(w, "service "+svc.Name, svc.Revision, svc.Updated)
			},
		},
	)
}

// Agents syncs agents: creates new ones, updates existing ones, and — only
// with opts.AllowDelete — deletes ones not present in the desired map;
// otherwise those are refused and reported as Protected. Updates go through
// PUT and replace the whole live spec, so the live revision each one
// overwrites is announced on warnW (deploy awareness), and the deletion
// decision is announced there before any agent write happens.
func Agents(
	ctx context.Context,
	warnW io.Writer,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]clients.CreateAgentBody,
	opts Options,
) (*Result, error) {
	existing, err := deployClient.ListAgents(
		ctx, orgId, projectId, stackId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	existingByName := make(map[string]clients.AgentOutput)
	for _, a := range existing {
		existingByName[a.Name] = a
	}

	return syncResources(
		ctx,
		warnW,
		existingByName,
		desired,
		opts,
		resourceOps[clients.AgentOutput, clients.CreateAgentBody]{
			resource:  "agent",
			allowFlag: "agents",
			create: func(name string, body clients.CreateAgentBody) error {
				_, err := deployClient.CreateAgent(ctx, orgId, projectId, name, body)
				return err
			},
			update: func(name string, body clients.CreateAgentBody) error {
				_, err := deployClient.PutAgent(ctx, orgId, projectId, name, body)
				return err
			},
			delete: func(name string) error {
				_, err := deployClient.DeleteAgent(ctx, orgId, projectId, name)
				return err
			},
			banner: func(w io.Writer, a clients.AgentOutput) {
				preflight.PrintUpdateBanner(w, "agent "+a.Name, a.Revision, a.Updated)
			},
		},
	)
}

// Databases syncs databases: creates new ones, updates existing ones, and —
// only with opts.AllowDelete — deletes ones not present in the desired map;
// otherwise those are refused and reported as Protected. The deletion
// decision is announced on warnW (deploy awareness) before any database
// write happens.
func Databases(
	ctx context.Context,
	warnW io.Writer,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]clients.CreateDatabaseBody,
	opts Options,
) (*Result, error) {
	existing, err := deployClient.ListDatabases(
		ctx, orgId, projectId, stackId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	existingByName := make(map[string]clients.DatabaseOutput)
	for _, db := range existing {
		existingByName[db.Name] = db
	}

	return syncResources(
		ctx,
		warnW,
		existingByName,
		desired,
		opts,
		resourceOps[clients.DatabaseOutput, clients.CreateDatabaseBody]{
			resource:  "database",
			allowFlag: "databases",
			create: func(name string, body clients.CreateDatabaseBody) error {
				_, err := deployClient.CreateDatabase(ctx, orgId, projectId, name, body)
				return err
			},
			update: func(name string, body clients.CreateDatabaseBody) error {
				_, err := deployClient.PutDatabase(ctx, orgId, projectId, name, body)
				return err
			},
			delete: func(name string) error {
				_, err := deployClient.DeleteDatabase(ctx, orgId, projectId, name)
				return err
			},
			// Databases carry no revision to announce, so no banner.
		},
	)
}

// resourceOps binds one resource type's API calls and wording for
// syncResources. banner, when set, announces the live state an update
// replaces; leave it nil for resources with nothing to announce.
type resourceOps[E, B any] struct {
	resource  string // singular noun, e.g. "service"
	allowFlag string // --allow-delete value, e.g. "services"
	create    func(name string, body B) error
	update    func(name string, body B) error
	delete    func(name string) error
	banner    func(w io.Writer, existing E)
}

// syncResources applies the shared sync contract for one resource type:
// announce the deletion decision on warnW before any write, refuse deletions
// without opts.AllowDelete (reporting them as Protected), create/update the
// desired resources in name order, delete last, and — with opts.DryRun —
// accumulate the same plan into the Result without calling the API at all.
func syncResources[E, B any](
	ctx context.Context,
	warnW io.Writer,
	existingByName map[string]E,
	desired map[string]B,
	opts Options,
	ops resourceOps[E, B],
) (*Result, error) {
	result := &Result{}

	toDelete := deletionList(existingByName, desired)
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

// deletionList returns, sorted, the existing resource names a sync will
// delete because the desired config no longer mentions them.
func deletionList[E, D any](existing map[string]E, desired map[string]D) []string {
	var names []string
	for name := range existing {
		if _, ok := desired[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
