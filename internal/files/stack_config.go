package files

import (
	"fmt"
	"os"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"gopkg.in/yaml.v3"
)

type StackConfig struct {
	Organization string                       `yaml:"organization"`
	Project      string                       `yaml:"project"`
	StackId      string                       `yaml:"stack-id"`
	Services     map[string]ServiceConfig     `yaml:"services"`
	VectorStores map[string]VectorStoreConfig `yaml:"vector-stores"`
	Agents       map[string]AgentConfig       `yaml:"agents"`
}

type VectorStoreConfig struct {
	Resources clients.VectorStoreResources `yaml:"resources"`
	Storage   clients.VectorStoreStorage   `yaml:"storage"`
	HA        bool                         `yaml:"ha"`
	Backups   bool                         `yaml:"backups"`
}

type ServiceConfig struct {
	Version     string               `yaml:"version,omitempty"`
	ServicePort int                  `yaml:"servicePort"`
	Image       clients.ImageSpec    `yaml:"image"`
	Resources   clients.Resources    `yaml:"resources"`
	Env         []clients.EnvVar     `yaml:"env,omitempty"`
	SecretRefs  []clients.SecretRef  `yaml:"secretRefs,omitempty"`
	Endpoint    bool                 `yaml:"endpoint,omitempty"`
	Replicas    int                  `yaml:"replicas,omitempty"`
	Autoscaling *clients.Autoscaling `yaml:"autoscaling,omitempty"`
	Healthcheck *clients.Healthcheck `yaml:"healthcheck,omitempty"`
	Schedule    *clients.Schedule    `yaml:"schedule,omitempty"`
}

type AgentConfig struct {
	Id          string              `yaml:"id"`
	Version     string              `yaml:"version"`
	AgentConfig any                 `yaml:"agentConfig"`
	SecretRefs  []clients.SecretRef `yaml:"secretRefs,omitempty"`
	Endpoint    bool                `yaml:"endpoint,omitempty"`
	Schedule    *clients.Schedule   `yaml:"schedule,omitempty"`
	Env         []clients.EnvVar    `yaml:"env,omitempty"`
}

func LoadStackConfig(path string) (*StackConfig, error) {
	if path == "" {
		return &StackConfig{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg StackConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if (len(cfg.Services) > 0 || len(cfg.VectorStores) > 0 || len(cfg.Agents) > 0) &&
		cfg.StackId == "" {
		return nil, fmt.Errorf(
			"stack-id is required when services, vector stores, or agents are defined in config file",
		)
	}

	if cfg.Services == nil {
		cfg.Services = make(map[string]ServiceConfig)
	}

	if cfg.VectorStores == nil {
		cfg.VectorStores = make(map[string]VectorStoreConfig)
	}

	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentConfig)
	}

	return &cfg, nil
}

func (v VectorStoreConfig) ToCreateRequest(stackId string) clients.CreateVectorStoreBody {
	return clients.CreateVectorStoreBody{
		Resources: v.Resources,
		Storage:   v.Storage,
		HA:        v.HA,
		Backups:   v.Backups,
		StackId:   stackId,
	}
}

func (a AgentConfig) ToCreateRequest(stackId string) clients.CreateAgentBody {
	return clients.CreateAgentBody{
		Id:          a.Id,
		Version:     a.Version,
		AgentConfig: a.AgentConfig,
		SecretRefs:  a.SecretRefs,
		Endpoint:    a.Endpoint,
		Schedule:    a.Schedule,
		Env:         a.Env,
		StackId:     stackId,
	}
}

func (s ServiceConfig) ToCreateRequest(stackId string) clients.CreateServiceBody {
	body := clients.CreateServiceBody{
		ServicePort: s.ServicePort,
		Image:       s.Image,
		Resources:   s.Resources,
		Env:         s.Env,
		SecretRefs:  s.SecretRefs,
		Endpoint:    s.Endpoint,
		StackId:     stackId,
	}

	body.Autoscaling = s.Autoscaling
	body.Replicas = s.Replicas

	if s.Healthcheck != nil {
		body.Healthcheck = s.Healthcheck
	}

	if s.Schedule != nil {
		body.Schedule = s.Schedule
	}

	return body
}
