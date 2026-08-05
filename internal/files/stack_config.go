package files

import (
	"context"
	"fmt"
	"os"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"gopkg.in/yaml.v3"
)

type StackConfig struct {
	Organization string                    `yaml:"organization" json:"organization"`
	Project      string                    `yaml:"project"      json:"project"`
	StackId      string                    `yaml:"stack-id"     json:"stackId"`
	Services     map[string]ServiceConfig  `yaml:"services"     json:"services"`
	Agents       map[string]AgentConfig    `yaml:"agents"       json:"agents"`
	Databases    map[string]DatabaseConfig `yaml:"databases"    json:"databases"`
}

type ServiceConfig struct {
	Version     string               `yaml:"version,omitempty"     json:"version,omitempty"`
	ServicePort int                  `yaml:"servicePort"           json:"servicePort"`
	Image       clients.ImageSpec    `yaml:"image"                 json:"image"`
	Resources   clients.Resources    `yaml:"resources"             json:"resources"`
	Env         []clients.EnvVar     `yaml:"env,omitempty"         json:"env,omitempty"`
	SecretRefs  []clients.SecretRef  `yaml:"secretRefs,omitempty"  json:"secretRefs,omitempty"`
	Endpoint    bool                 `yaml:"endpoint,omitempty"    json:"endpoint,omitempty"`
	Replicas    int                  `yaml:"replicas,omitempty"    json:"replicas,omitempty"`
	Autoscaling *clients.Autoscaling `yaml:"autoscaling,omitempty" json:"autoscaling,omitempty"`
	Healthcheck *clients.Healthcheck `yaml:"healthcheck,omitempty" json:"healthcheck,omitempty"`
	Schedule    *clients.Schedule    `yaml:"schedule,omitempty"    json:"schedule,omitempty"`
}

type DatabaseConfig struct {
	Instances       int                           `yaml:"instances"                 json:"instances"`
	PostgresVersion string                        `yaml:"postgresVersion,omitempty" json:"postgresVersion,omitempty"`
	Resources       clients.Resources             `yaml:"resources"                 json:"resources"`
	Storage         clients.DatabaseStorageConfig `yaml:"storage"                   json:"storage"`
	Extensions      []string                      `yaml:"extensions,omitempty"      json:"extensions,omitempty"`
	Backup          *clients.DatabaseBackupConfig `yaml:"backup,omitempty"          json:"backup,omitempty"`
}

type AgentConfig struct {
	Id          string              `yaml:"id"                   json:"id"`
	Version     string              `yaml:"version"              json:"version"`
	AgentConfig any                 `yaml:"agentConfig"          json:"agentConfig"`
	SecretRefs  []clients.SecretRef `yaml:"secretRefs,omitempty" json:"secretRefs,omitempty"`
	Endpoint    bool                `yaml:"endpoint,omitempty"   json:"endpoint,omitempty"`
	Schedule    *clients.Schedule   `yaml:"schedule,omitempty"   json:"schedule,omitempty"`
	Env         []clients.EnvVar    `yaml:"env,omitempty"        json:"env,omitempty"`
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

	if (len(cfg.Services) > 0 || len(cfg.Agents) > 0 || len(cfg.Databases) > 0) &&
		cfg.StackId == "" {
		return nil, fmt.Errorf(
			"stack-id is required when services, agents, or databases are defined in config file",
		)
	}

	if cfg.Services == nil {
		cfg.Services = make(map[string]ServiceConfig)
	}

	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentConfig)
	}

	if cfg.Databases == nil {
		cfg.Databases = make(map[string]DatabaseConfig)
	}

	return &cfg, nil
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

func (d DatabaseConfig) ToCreateRequest(stackId string) clients.CreateDatabaseBody {
	return clients.CreateDatabaseBody{
		Instances:       d.Instances,
		PostgresVersion: d.PostgresVersion,
		Resources:       d.Resources,
		Storage:         d.Storage,
		Extensions:      d.Extensions,
		Backup:          d.Backup,
		StackId:         stackId,
	}
}

// ServiceConfigFromDescribe converts a DescribeServiceResponse to a ServiceConfig.
func ServiceConfigFromDescribe(svc *clients.DescribeServiceResponse) ServiceConfig {
	return ServiceConfig{
		ServicePort: svc.ServicePort,
		Image:       svc.Image,
		Resources:   svc.Resources,
		Env:         svc.Env,
		SecretRefs:  svc.SecretRefs,
		Endpoint:    svc.Endpoint != "",
		Replicas:    svc.Replicas,
		Autoscaling: svc.Autoscaling,
		Healthcheck: svc.Healthcheck,
		Schedule:    svc.Schedule,
	}
}

// AgentConfigFromDescribe converts a DescribeAgentResponse to an AgentConfig.
func AgentConfigFromDescribe(agent *clients.DescribeAgentResponse) AgentConfig {
	return AgentConfig{
		Id:          agent.Id,
		Version:     agent.Version,
		AgentConfig: agent.AgentConfig,
		SecretRefs:  agent.SecretRefs,
		Endpoint:    agent.Endpoint != "",
		Schedule:    agent.Schedule,
		Env:         agent.Env,
	}
}

// DatabaseConfigFromDescribe converts a DescribeDatabaseResponse to a DatabaseConfig.
func DatabaseConfigFromDescribe(db *clients.DescribeDatabaseResponse) DatabaseConfig {
	var backup *clients.DatabaseBackupConfig
	if db.Backup.Enabled {
		backup = &clients.DatabaseBackupConfig{
			Schedule:        db.Backup.Schedule,
			RetentionPolicy: db.Backup.RetentionPolicy,
		}
	}
	return DatabaseConfig{
		Instances:       db.Replicas,
		PostgresVersion: db.PostgresVersion,
		Resources:       db.Resources,
		Storage:         db.Storage,
		Extensions:      db.Extensions,
		Backup:          backup,
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

func FetchLiveStack(
	ctx context.Context,
	deployClient *clients.DeploymentClient,
	orgId, projectId, stackId string,
) (*StackConfig, error) {
	cfg := &StackConfig{
		StackId:   stackId,
		Services:  make(map[string]ServiceConfig),
		Agents:    make(map[string]AgentConfig),
		Databases: make(map[string]DatabaseConfig),
	}

	svcs, err := deployClient.ListServices(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	for _, svc := range svcs {
		desc, err := deployClient.DescribeService(ctx, orgId, projectId, svc.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe service %q: %w", svc.Name, err)
		}
		cfg.Services[svc.Name] = ServiceConfigFromDescribe(desc)
	}

	agents, err := deployClient.ListAgents(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	for _, a := range agents {
		desc, err := deployClient.DescribeAgent(ctx, orgId, projectId, a.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe agent %q: %w", a.Name, err)
		}
		cfg.Agents[a.Name] = AgentConfigFromDescribe(desc)
	}

	dbs, err := deployClient.ListDatabases(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	for _, db := range dbs {
		desc, err := deployClient.DescribeDatabase(ctx, orgId, projectId, db.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe database %q: %w", db.Name, err)
		}
		cfg.Databases[db.Name] = DatabaseConfigFromDescribe(desc)
	}

	return cfg, nil
}

func MarshalStackConfig(cfg *StackConfig) ([]byte, error) {
	return yaml.Marshal(cfg)
}
