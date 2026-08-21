package files

import (
	"context"
	"fmt"
	"os"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"gopkg.in/yaml.v3"
)

type StackConfig struct {
	Organization string                    `yaml:"organization" json:"organization"`
	Project      string                    `yaml:"project"      json:"project"`
	StackId      string                    `yaml:"stack-id"     json:"stackId"`
	Services     map[string]ServiceConfig  `yaml:"services"     json:"services"`
	Agents       map[string]AgentConfig    `yaml:"agents"       json:"agents"`
	Databases    map[string]DatabaseConfig `yaml:"databases"    json:"databases"`
	Mcps         map[string]McpConfig      `yaml:"mcps"         json:"mcps"`
}

type ServiceConfig struct {
	Version     string                  `yaml:"version,omitempty"     json:"version,omitempty"`
	ServicePort int                     `yaml:"servicePort"           json:"servicePort"`
	Image       deployment.ImageSpec    `yaml:"image"                 json:"image"`
	Resources   deployment.Resources    `yaml:"resources"             json:"resources"`
	Env         []deployment.EnvVar     `yaml:"env,omitempty"         json:"env,omitempty"`
	SecretRefs  []deployment.SecretRef  `yaml:"secretRefs,omitempty"  json:"secretRefs,omitempty"`
	Endpoint    bool                    `yaml:"endpoint,omitempty"    json:"endpoint,omitempty"`
	Replicas    int                     `yaml:"replicas,omitempty"    json:"replicas,omitempty"`
	Autoscaling *deployment.Autoscaling `yaml:"autoscaling,omitempty" json:"autoscaling,omitempty"`
	Healthcheck *deployment.Healthcheck `yaml:"healthcheck,omitempty" json:"healthcheck,omitempty"`
	Schedule    *deployment.Schedule    `yaml:"schedule,omitempty"    json:"schedule,omitempty"`
}

type DatabaseConfig struct {
	Instances       int                              `yaml:"instances"                 json:"instances"`
	PostgresVersion string                           `yaml:"postgresVersion,omitempty" json:"postgresVersion,omitempty"`
	Resources       deployment.Resources             `yaml:"resources"                 json:"resources"`
	Storage         deployment.DatabaseStorageConfig `yaml:"storage"                   json:"storage"`
	Extensions      []string                         `yaml:"extensions,omitempty"      json:"extensions,omitempty"`
	Backup          *deployment.DatabaseBackupConfig `yaml:"backup,omitempty"          json:"backup,omitempty"`
}

type AgentConfig struct {
	Id          string                 `yaml:"id"                   json:"id"`
	Version     string                 `yaml:"version"              json:"version"`
	AgentConfig any                    `yaml:"agentConfig"          json:"agentConfig"`
	SecretRefs  []deployment.SecretRef `yaml:"secretRefs,omitempty" json:"secretRefs,omitempty"`
	Endpoint    bool                   `yaml:"endpoint,omitempty"   json:"endpoint,omitempty"`
	Schedule    *deployment.Schedule   `yaml:"schedule,omitempty"   json:"schedule,omitempty"`
	Env         []deployment.EnvVar    `yaml:"env,omitempty"        json:"env,omitempty"`
}

type McpConfig struct {
	Type        string                 `yaml:"type,omitempty"        json:"type,omitempty"`
	Port        int                    `yaml:"port,omitempty"        json:"port,omitempty"`
	Path        string                 `yaml:"path,omitempty"        json:"path,omitempty"`
	Image       deployment.ImageSpec   `yaml:"image,omitempty"       json:"image,omitempty"`
	Resources   deployment.Resources   `yaml:"resources,omitempty"   json:"resources,omitempty"`
	Env         []deployment.EnvVar    `yaml:"env,omitempty"         json:"env,omitempty"`
	SecretRefs  []deployment.SecretRef `yaml:"secretRefs,omitempty"  json:"secretRefs,omitempty"`
	EndpointURL string                 `yaml:"endpointUrl,omitempty" json:"endpointUrl,omitempty"`
	CatalogID   string                 `yaml:"catalogId,omitempty"   json:"catalogId,omitempty"`
	Auth        deployment.McpAuthBody `yaml:"auth,omitempty"        json:"auth"`
	Headers     map[string]string      `yaml:"headers,omitempty"     json:"headers,omitempty"`
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

	if (len(cfg.Services) > 0 || len(cfg.Agents) > 0 || len(cfg.Databases) > 0 || len(cfg.Mcps) > 0) &&
		cfg.StackId == "" {
		return nil, fmt.Errorf(
			"stack-id is required when services, agents, databases, or mcps are defined in config file",
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

	if cfg.Mcps == nil {
		cfg.Mcps = make(map[string]McpConfig)
	}

	return &cfg, nil
}

func (a AgentConfig) ToCreateRequest(stackId string) deployment.CreateAgentBody {
	return deployment.CreateAgentBody{
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

func (d DatabaseConfig) ToCreateRequest(stackId string) deployment.CreateDatabaseBody {
	return deployment.CreateDatabaseBody{
		Instances:       d.Instances,
		PostgresVersion: d.PostgresVersion,
		Resources:       d.Resources,
		Storage:         d.Storage,
		Extensions:      d.Extensions,
		Backup:          d.Backup,
		StackId:         stackId,
	}
}

func (m McpConfig) ToCreateRequest(stackId string) deployment.CreateMcpBody {
	return deployment.CreateMcpBody{
		Type:        m.Type,
		Port:        m.Port,
		Path:        m.Path,
		Image:       m.Image,
		Resources:   m.Resources,
		Env:         m.Env,
		SecretRefs:  m.SecretRefs,
		EndpointURL: m.EndpointURL,
		CatalogID:   m.CatalogID,
		Auth:        m.Auth,
		Headers:     m.Headers,
		StackId:     stackId,
	}
}

func ServiceConfigFromDescribe(svc *deployment.DescribeServiceResponse) ServiceConfig {
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

func AgentConfigFromDescribe(agent *deployment.DescribeAgentResponse) AgentConfig {
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

func DatabaseConfigFromDescribe(db *deployment.DescribeDatabaseResponse) DatabaseConfig {
	var backup *deployment.DatabaseBackupConfig
	if db.Backup.Enabled {
		backup = &deployment.DatabaseBackupConfig{
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

func McpConfigFromDescribe(mcp *deployment.DescribeMcpResponse) McpConfig {
	endpointURL := ""
	if mcp.Type == "external" {
		endpointURL = mcp.EndpointURL
	}
	return McpConfig{
		Type:        mcp.Type,
		Port:        mcp.Port,
		Path:        mcp.Path,
		Image:       mcp.Image,
		Resources:   mcp.Resources,
		Env:         mcp.Env,
		SecretRefs:  mcp.SecretRefs,
		EndpointURL: endpointURL,
		CatalogID:   mcp.CatalogID,
		Auth: deployment.McpAuthBody{
			Type:         mcp.Auth.Type,
			Header:       mcp.Auth.Header,
			HeaderPrefix: mcp.Auth.HeaderPrefix,
		},
		Headers: mcp.Headers,
	}
}

func (s ServiceConfig) ToCreateRequest(stackId string) deployment.CreateServiceBody {
	body := deployment.CreateServiceBody{
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
	deployClient *deployment.DeploymentClient,
	orgId, projectId, stackId string,
) (*StackConfig, error) {
	cfg := &StackConfig{
		StackId:   stackId,
		Services:  make(map[string]ServiceConfig),
		Agents:    make(map[string]AgentConfig),
		Databases: make(map[string]DatabaseConfig),
		Mcps:      make(map[string]McpConfig),
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

	mcps, err := deployClient.ListMcps(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list mcps: %w", err)
	}
	for _, mcp := range mcps {
		desc, err := deployClient.DescribeMcp(ctx, orgId, projectId, mcp.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe mcp %q: %w", mcp.Name, err)
		}
		cfg.Mcps[mcp.Name] = McpConfigFromDescribe(desc)
	}

	return cfg, nil
}

func MarshalStackConfig(cfg *StackConfig) ([]byte, error) {
	return yaml.Marshal(cfg)
}
