package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/output"
)

type SyncResult struct {
	Created   []string
	Updated   []string
	Deleted   []string
	Skipped   []string
	Protected []string // would be deleted but deletion was not allowed
}

func printSyncOutcome(out io.Writer, label string, result *SyncResult, err error) error {
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
	return nil
}

func SyncServices(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]clients.CreateServiceBody,
) (*SyncResult, error) {
	existing, err := deployClient.ListServices(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	existingByName := make(map[string]clients.ServiceOutput)
	for _, svc := range existing {
		existingByName[svc.Name] = svc
	}

	result := &SyncResult{
		Created: []string{},
		Updated: []string{},
		Deleted: []string{},
	}

	desiredNames := make([]string, 0, len(desired))
	for name := range desired {
		desiredNames = append(desiredNames, name)
	}
	sort.Strings(desiredNames)

	for _, name := range desiredNames {
		body := desired[name]
		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateService(ctx, orgId, projectId, name, body)
			if err != nil {
				return result, fmt.Errorf("failed to create service %q: %w", name, err)
			}
			result.Created = append(result.Created, name)
		} else {
			_, err := deployClient.UpdateService(ctx, orgId, projectId, name, body)
			if err != nil {
				return result, fmt.Errorf("failed to update service %q: %w", name, err)
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
			_, err := deployClient.DeleteService(ctx, orgId, projectId, name)
			if err != nil {
				return result, fmt.Errorf("failed to delete service %q: %w", name, err)
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}

// SyncVectorStores syncs vector stores. Existing stores are skipped (no update endpoint).
// When allowDelete is false, stores that would be deleted are collected in Protected instead.
func SyncVectorStores(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId,
	projectId,
	stackId string,
	desired map[string]clients.CreateVectorStoreBody,
	allowDelete bool,
) (*SyncResult, error) {
	existing, err := deployClient.ListVectorStores(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list vector stores: %w", err)
	}

	existingByName := make(map[string]clients.VectorStoreInfo)
	for _, vs := range existing {
		existingByName[vs.VectorStoreName] = vs
	}

	result := &SyncResult{
		Created: []string{},
		Updated: []string{},
		Deleted: []string{},
	}

	desiredNames := make([]string, 0, len(desired))
	for name := range desired {
		desiredNames = append(desiredNames, name)
	}
	sort.Strings(desiredNames)

	for _, name := range desiredNames {
		body := desired[name]
		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateVectorStore(ctx, orgId, projectId, name, body)
			if err != nil {
				return result, fmt.Errorf("failed to create vector store %q: %w", name, err)
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
			_, err := deployClient.DeleteVectorStore(ctx, orgId, projectId, name)
			if err != nil {
				return result, fmt.Errorf("failed to delete vector store %q: %w", name, err)
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}
