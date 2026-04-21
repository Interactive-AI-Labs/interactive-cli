package sync

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
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
	Skipped   []string
	Protected []string // would be deleted but deletion was not allowed
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

func HasVectorStores(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
) (bool, error) {
	existing, err := deployClient.ListVectorStores(ctx, orgId, projectId, stackId)
	if err != nil {
		return false, fmt.Errorf("failed to list vector stores: %w", err)
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
				result.Skipped,
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
		result.Skipped,
	)
	if len(result.Protected) > 0 {
		fmt.Fprintf(
			out,
			"\nProtected %s (not deleted): %s\n"+
				"Use --allow-delete=%s to delete them.\n",
			label,
			strings.Join(result.Protected, ", "),
			strings.ReplaceAll(label, " ", "-"),
		)
	}
	return nil
}

func Services(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]clients.CreateServiceBody,
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

	result := &Result{}

	desiredNames := make([]string, 0, len(desired))
	for name := range desired {
		desiredNames = append(desiredNames, name)
	}
	sort.Strings(desiredNames)

	for _, name := range desiredNames {
		body := desired[name]
		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateService(
				ctx, orgId, projectId, name, body,
			)
			if err != nil {
				return result, fmt.Errorf(
					"failed to create service %q: %w", name, err,
				)
			}
			result.Created = append(result.Created, name)
		} else {
			_, err := deployClient.UpdateService(
				ctx, orgId, projectId, name, body,
			)
			if err != nil {
				return result, fmt.Errorf(
					"failed to update service %q: %w", name, err,
				)
			}
			result.Updated = append(result.Updated, name)
		}
	}

	existingNames := make([]string, 0, len(existingByName))
	for name := range existingByName {
		existingNames = append(existingNames, name)
	}
	sort.Strings(existingNames)

	for _, name := range existingNames {
		if _, ok := desired[name]; !ok {
			_, err := deployClient.DeleteService(
				ctx, orgId, projectId, name,
			)
			if err != nil {
				return result, fmt.Errorf(
					"failed to delete service %q: %w", name, err,
				)
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}

// Agents syncs agents: creates new ones, updates existing ones, and deletes ones
// not present in the desired map.
func Agents(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]clients.CreateAgentBody,
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

	result := &Result{}

	desiredNames := make([]string, 0, len(desired))
	for name := range desired {
		desiredNames = append(desiredNames, name)
	}
	sort.Strings(desiredNames)

	for _, name := range desiredNames {
		body := desired[name]
		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateAgent(
				ctx, orgId, projectId, name, body,
			)
			if err != nil {
				return result, fmt.Errorf(
					"failed to create agent %q: %w", name, err,
				)
			}
			result.Created = append(result.Created, name)
		} else {
			_, err := deployClient.UpdateAgent(
				ctx, orgId, projectId, name, body,
			)
			if err != nil {
				return result, fmt.Errorf(
					"failed to update agent %q: %w", name, err,
				)
			}
			result.Updated = append(result.Updated, name)
		}
	}

	existingNames := make([]string, 0, len(existingByName))
	for name := range existingByName {
		existingNames = append(existingNames, name)
	}
	sort.Strings(existingNames)

	for _, name := range existingNames {
		if _, ok := desired[name]; !ok {
			_, err := deployClient.DeleteAgent(
				ctx, orgId, projectId, name,
			)
			if err != nil {
				return result, fmt.Errorf(
					"failed to delete agent %q: %w", name, err,
				)
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}

// VectorStores syncs vector stores. Existing stores are skipped (no update endpoint).
// When allowDelete is false, stores that would be deleted are collected in Protected instead.
func VectorStores(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]clients.CreateVectorStoreBody,
	allowDelete bool,
) (*Result, error) {
	existing, err := deployClient.ListVectorStores(
		ctx, orgId, projectId, stackId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list vector stores: %w", err)
	}

	existingByName := make(map[string]clients.VectorStoreInfo)
	for _, vs := range existing {
		existingByName[vs.VectorStoreName] = vs
	}

	result := &Result{}

	desiredNames := make([]string, 0, len(desired))
	for name := range desired {
		desiredNames = append(desiredNames, name)
	}
	sort.Strings(desiredNames)

	for _, name := range desiredNames {
		body := desired[name]
		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateVectorStore(
				ctx, orgId, projectId, name, body,
			)
			if err != nil {
				return result, fmt.Errorf(
					"failed to create vector store %q: %w",
					name, err,
				)
			}
			result.Created = append(result.Created, name)
		} else {
			result.Skipped = append(result.Skipped, name)
		}
	}

	existingNames := make([]string, 0, len(existingByName))
	for name := range existingByName {
		existingNames = append(existingNames, name)
	}
	sort.Strings(existingNames)

	for _, name := range existingNames {
		if _, ok := desired[name]; !ok {
			if !allowDelete {
				result.Protected = append(result.Protected, name)
				continue
			}
			_, err := deployClient.DeleteVectorStore(
				ctx, orgId, projectId, name,
			)
			if err != nil {
				return result, fmt.Errorf(
					"failed to delete vector store %q: %w",
					name, err,
				)
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}
