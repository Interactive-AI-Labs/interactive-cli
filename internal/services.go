package internal

import (
	"context"
)

type ResourceRequirements struct {
	Memory string `json:"memory" yaml:"memory"`
	CPU    string `json:"cpu" yaml:"cpu"`
}

type Resources struct {
	Requests ResourceRequirements `json:"requests" yaml:"requests"`
	Limits   ResourceRequirements `json:"limits" yaml:"limits"`
}

type ImageSpec struct {
	Type       string `json:"type" yaml:"type"`
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty"`
	Name       string `json:"name" yaml:"name"`
	Tag        string `json:"tag" yaml:"tag"`
}

type EnvVar struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type SecretRef struct {
	SecretName string `json:"secretName" yaml:"secretName"`
}

type CreateServiceBody struct {
	ServicePort int         `json:"servicePort"`
	Image       ImageSpec   `json:"image"`
	Resources   Resources   `json:"resources"`
	Env         []EnvVar    `json:"env,omitempty"`
	SecretRefs  []SecretRef `json:"secretRefs,omitempty"`
	Endpoint    bool        `json:"endpoint,omitempty"`
	Hostname    string      `json:"hostname,omitempty"`
	Replicas    int         `json:"replicas"`
	StackId     string      `json:"stackId,omitempty"`
}

type ServiceOutput struct {
	Name      string `json:"name"`
	ProjectId string `json:"projectId"`
	Revision  int    `json:"revision"`
	Status    string `json:"status"`
	Updated   string `json:"updated,omitempty"`
	Endpoint  string `json:"endpoint,omitempty"`
}

type SyncResult struct {
	Created []string
	Updated []string
	Deleted []string
}

func SyncServices(
	ctx context.Context,
	apiClient *APIClient,
	deployClient *DeploymentClient,
	orgName,
	projectName string,
	cfg *StackConfig,
) (*SyncResult, error) {
	orgId, projectId, err := apiClient.GetProjectId(ctx, orgName, projectName)
	if err != nil {
		return nil, err
	}

	existing, err := deployClient.ListServices(ctx, orgId, projectId, cfg.StackId)
	if err != nil {
		return nil, err
	}

	existingByName := make(map[string]ServiceOutput)
	for _, svc := range existing {
		existingByName[svc.Name] = svc
	}

	result := &SyncResult{
		Created: []string{},
		Updated: []string{},
		Deleted: []string{},
	}

	for name, svcCfg := range cfg.Services {
		req := svcCfg.ToCreateRequest(cfg.StackId)

		if _, exists := existingByName[name]; !exists {
			_, err := deployClient.CreateService(ctx, orgId, projectId, name, req)
			if err != nil {
				return nil, err
			}
			result.Created = append(result.Created, name)
		} else {
			_, err := deployClient.UpdateService(ctx, orgId, projectId, name, req)
			if err != nil {
				return nil, err
			}
			result.Updated = append(result.Updated, name)
		}
	}

	for name := range existingByName {
		if _, desired := cfg.Services[name]; !desired {
			_, err := deployClient.DeleteService(ctx, orgId, projectId, name)
			if err != nil {
				return nil, err
			}
			result.Deleted = append(result.Deleted, name)
		}
	}

	return result, nil
}
