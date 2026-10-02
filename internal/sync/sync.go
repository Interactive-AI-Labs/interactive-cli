package sync

import (
	"cmp"
	"context"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
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
	apiClient *platform.APIClient,
	orgId,
	projectId,
	stackId string,
) (bool, error) {
	existing, err := deployClient.ListMcps(ctx, orgId, projectId, stackId)
	if err != nil {
		return false, fmt.Errorf("failed to list mcps: %w", err)
	}
	if len(existing) > 0 {
		return true, nil
	}
	remote, err := apiClient.ListRemoteMcps(ctx, orgId, projectId, stackId)
	if err != nil {
		return false, fmt.Errorf("failed to list remote mcps: %w", err)
	}

	return len(remote) > 0, nil
}

func HasJobs(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
) (bool, error) {
	existing, err := deployClient.ListJobs(ctx, orgId, projectId, stackId)
	if err != nil {
		return false, fmt.Errorf("failed to list jobs: %w", err)
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

// Mcps syncs self-hosted MCPs on the deployment operator and remote MCPs on the platform, which owns them.
func Mcps(
	ctx context.Context,
	warnW io.Writer,
	deployClient *deployment.DeploymentClient,
	apiClient *platform.APIClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]deployment.CreateMcpBody,
	remote map[string]platform.McpCreateRequest,
	opts Options,
) (*Result, error) {
	// Listed project-wide on both sides: names are unique per project, so a clash must fail before any write.
	operatorList, err := deployClient.ListMcps(ctx, orgId, projectId, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list mcps: %w", err)
	}
	operatorByName := make(map[string]deployment.McpOutput)
	existingByName := make(map[string]deployment.McpOutput)
	for _, mcp := range operatorList {
		operatorByName[mcp.Name] = mcp
		if mcp.StackId == stackId {
			existingByName[mcp.Name] = mcp
		}
	}

	platformList, _, err := apiClient.ListMcps(ctx, orgId, projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to list platform mcps: %w", err)
	}
	platformByName := make(map[string]platform.McpSchema)
	for _, mcp := range platformList.Mcps {
		platformByName[mcp.Name] = mcp
	}
	remoteByName := make(map[string]platform.McpSchema)
	var skipped []string
	for _, mcp := range platform.RemoteStackMcps(platformList.Mcps, stackId) {
		if files.StackManagedRemote(mcp) {
			remoteByName[mcp.Name] = mcp
			continue
		}
		// A name the config claims is reported as a conflict below, not as skipped.
		if _, claimed := remote[mcp.Name]; !claimed {
			skipped = append(skipped, mcp.Name)
		}
	}
	if len(skipped) > 0 {
		sort.Strings(skipped)
		fmt.Fprintf(
			warnW,
			"Remote mcps %s in stack %q use an auth type a stack file cannot hold; "+
				"they are managed with 'iai mcps' only and left as they are.\n",
			strings.Join(skipped, ", "), stackId,
		)
	}

	typeChanged, unmanaged := mcpConflicts(
		operatorByName,
		platformByName,
		stackId,
		desired,
		remote,
	)
	if len(typeChanged) > 0 {
		return nil, fmt.Errorf(
			"mcps %s exist live with another type or as legacy remote mcps on the deployment operator; "+
				"omit them from the config, sync with --allow-delete mcps, then add them back",
			strings.Join(typeChanged, ", "),
		)
	}
	if len(unmanaged) > 0 {
		return nil, fmt.Errorf(
			"mcps %s exist in the project but stack %q cannot manage them "+
				"(another stack, or an auth type set up with 'iai mcps'); "+
				"delete them with 'iai mcps delete', or with 'iai stacks sync --allow-delete mcps' "+
				"on the stack that owns them, or rename them in the config",
			strings.Join(unmanaged, ", "), stackId,
		)
	}

	result, err := syncResources(
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
				if err := requireMcpCredential(name, authType, body.Auth.Credential); err != nil {
					return err
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
	if err != nil {
		return result, err
	}

	remoteResult, err := syncResources(
		warnW,
		remoteByName,
		remote,
		opts,
		resourceOps[platform.McpSchema, platform.McpCreateRequest]{
			resource:  "mcp",
			allowFlag: "mcps",
			create: func(name string, body platform.McpCreateRequest) error {
				created, _, err := apiClient.CreateMcp(ctx, orgId, projectId, body)
				if err != nil {
					return err
				}
				kept := func(m *platform.McpSchema) bool {
					return m.StackId != nil && *m.StackId == stackId
				}
				if kept(created) {
					return nil
				}
				// The read path decides: a platform may store the stack yet omit it from the create response.
				live, _, descErr := apiClient.DescribeMcp(ctx, orgId, projectId, name)
				if descErr == nil && kept(live) {
					return nil
				}
				// An older platform drops the stack; delete the record this call just made rather than leave an mcp no stack command can find.
				if _, _, delErr := apiClient.DeleteMcp(ctx, orgId, projectId, name); delErr != nil {
					return fmt.Errorf(
						"the platform did not keep stack %q for mcp %q and removing it failed: %w",
						stackId, name, delErr,
					)
				}
				return fmt.Errorf(
					"the platform did not keep stack %q for mcp %q; upgrade the platform before syncing remote mcps",
					stackId,
					name,
				)
			},
			update: func(name string, body platform.McpCreateRequest) error {
				patch, err := remoteMcpPatch(name, remoteByName[name], body)
				if err != nil || len(patch) == 0 {
					return err
				}
				_, _, err = apiClient.UpdateMcp(ctx, orgId, projectId, name, patch)
				return err
			},
			delete: func(name string) error {
				_, _, err := apiClient.DeleteMcp(ctx, orgId, projectId, name)
				return err
			},
		},
	)
	result.Created = append(result.Created, remoteResult.Created...)
	result.Updated = append(result.Updated, remoteResult.Updated...)
	result.Deleted = append(result.Deleted, remoteResult.Deleted...)
	result.Protected = append(result.Protected, remoteResult.Protected...)
	// A name live on both sides would otherwise be listed twice.
	for _, names := range []*[]string{&result.Created, &result.Updated, &result.Deleted, &result.Protected} {
		sort.Strings(*names)
		*names = slices.Compact(*names)
	}
	// The merged partial result lets PrintResult show what landed before a remote error.
	return result, err
}

// mcpConflicts lists config MCPs the sync cannot update in place: a live type that differs, or a name taken by an mcp this stack cannot manage.
func mcpConflicts(
	operatorMcps map[string]deployment.McpOutput,
	platformMcps map[string]platform.McpSchema,
	stackId string,
	selfHosted map[string]deployment.CreateMcpBody,
	remote map[string]platform.McpCreateRequest,
) (typeChanged, unmanaged []string) {
	// The sync never deletes an mcp outside the stack or one a stack file cannot hold, so a type change cannot free its name.
	classify := func(name string, backend platform.McpBackend) {
		// The platform does not list a legacy operator-side remote mcp, so the operator row decides first.
		if op, ok := operatorMcps[name]; ok {
			switch {
			case op.StackId != stackId:
				unmanaged = append(unmanaged, name)
			case backend == platform.McpBackendExternal ||
				deployment.McpTypeName(op.Type) == deployment.McpTypeRemote:
				typeChanged = append(typeChanged, name)
			}
			return
		}
		live, ok := platformMcps[name]
		if !ok {
			return
		}
		inStack := live.StackId != nil && *live.StackId == stackId
		switch {
		case !inStack,
			live.Backend == platform.McpBackendExternal && !files.StackManagedRemote(live):
			unmanaged = append(unmanaged, name)
		case live.Backend != backend:
			typeChanged = append(typeChanged, name)
		}
	}
	for name := range selfHosted {
		classify(name, platform.McpBackendInternal)
	}
	for name := range remote {
		classify(name, platform.McpBackendExternal)
	}
	sort.Strings(typeChanged)
	sort.Strings(unmanaged)
	return typeChanged, unmanaged
}

// requireMcpCredential refuses an update that would drop a credential: stack get never exports them.
func requireMcpCredential(name, authType, credential string) error {
	if credential == "" && authType != "" && !strings.EqualFold(authType, "none") {
		return fmt.Errorf(
			"auth.credential is required to update mcp %q; stack get never exports credentials",
			name,
		)
	}
	return nil
}

// remoteMcpPatch builds the platform's partial update for a remote MCP, or refuses what it cannot change: the catalog entry, and a credential the config dropped.
func remoteMcpPatch(
	name string,
	live platform.McpSchema,
	body platform.McpCreateRequest,
) (platform.McpUpdateRequest, error) {
	str := func(p *string) string {
		if p == nil {
			return ""
		}
		return *p
	}
	if str(live.CatalogID) != str(body.CatalogID) {
		return nil, fmt.Errorf(
			"mcp %q changed its catalog entry; omit it from the config, sync with --allow-delete mcps, then add it back",
			name,
		)
	}
	patch := platform.McpUpdateRequest{}
	if body.EndpointURL != nil {
		patch["endpoint_url"] = *body.EndpointURL
	}
	authType, credential := string(body.Auth.Type), str(body.Auth.Credential)
	if credential == "" && authType != string(platform.McpAuthNone) {
		// A partial update keeps the live credential, so a bearer or api_key config without one leaves auth alone.
		if live.HasCredential && authType == str(live.AuthType) &&
			authType != string(platform.McpAuthCustom) {
			return patch, nil
		}
		return nil, requireMcpCredential(name, authType, credential)
	}
	auth := map[string]any{"type": authType}
	if credential != "" {
		auth["credential"] = credential
	}
	// An endpoint-backed mcp always sends its header routing, as null when unset, or a dropped override would stay live; a catalog entry owns it.
	if body.CatalogID == nil {
		auth["header_name"], auth["header_prefix"] = nil, nil
		if body.Auth.HeaderName != nil {
			auth["header_name"] = *body.Auth.HeaderName
		}
		if body.Auth.HeaderPrefix != nil {
			auth["header_prefix"] = *body.Auth.HeaderPrefix
		}
	}
	patch["auth"] = auth
	return patch, nil
}

const mcpWaitInterval = 5 * time.Second

// WaitForMcps polls each mcp until its verify status is ok, failing once timeout passes.
func WaitForMcps(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId string,
	names []string,
	timeout time.Duration,
) error {
	fetch := func(ctx context.Context, name string) (deployment.McpVerifyState, error) {
		mcp, err := deployClient.DescribeMcp(ctx, orgId, projectId, name)
		if err != nil {
			return deployment.McpVerifyState{}, err
		}
		return mcp.Verify, nil
	}
	return waitForMcps(ctx, fetch, names, timeout, mcpWaitInterval)
}

func waitForMcps(
	ctx context.Context,
	fetch func(context.Context, string) (deployment.McpVerifyState, error),
	names []string,
	timeout,
	interval time.Duration,
) error {
	timedOut := time.After(timeout)
	for {
		var notReady []string
		for _, name := range names {
			state, err := fetch(ctx, name)
			if err != nil {
				state = deployment.McpVerifyState{Status: "check failed", Error: err.Error()}
			}
			if state.Status == "ok" {
				continue
			}
			reason := cmp.Or(state.Status, "unknown")
			if state.Error != "" {
				reason += ": " + state.Error
			}
			notReady = append(notReady, fmt.Sprintf("%s (%s)", name, reason))
		}
		if len(notReady) == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timedOut:
			return fmt.Errorf(
				"mcps not ready after %s: %s",
				timeout,
				strings.Join(notReady, ", "),
			)
		case <-time.After(interval):
		}
	}
}

func Jobs(
	ctx context.Context,
	warnW io.Writer,
	deployClient *deployment.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]deployment.CreateJobBody,
	opts Options,
) (*Result, error) {
	existing, err := deployClient.ListJobs(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	existingByName := make(map[string]deployment.JobOutput)
	for _, job := range existing {
		existingByName[job.Name] = job
	}

	return syncResources(
		warnW,
		existingByName,
		desired,
		opts,
		resourceOps[deployment.JobOutput, deployment.CreateJobBody]{
			resource:  "job",
			allowFlag: "jobs",
			create: func(name string, body deployment.CreateJobBody) error {
				_, err := deployClient.CreateJob(ctx, orgId, projectId, name, body)
				return err
			},
			update: func(name string, body deployment.CreateJobBody) error {
				_, err := deployClient.PutJob(ctx, orgId, projectId, name, body)
				return err
			},
			delete: func(name string) error {
				_, err := deployClient.DeleteJob(ctx, orgId, projectId, name)
				return err
			},
			banner: func(w io.Writer, job deployment.JobOutput) {
				preflight.PrintUpdateBanner(w, "job "+job.Name, job.Revision, job.Updated)
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
