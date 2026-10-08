package files

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"gopkg.in/yaml.v3"
)

// Auth types a stack file can hold for a remote MCP; oauth and client_credentials need 'iai mcps create'.
var stackRemoteAuthTypes = []string{
	string(platform.McpAuthNone),
	string(platform.McpAuthBearer),
	string(platform.McpAuthAPIKey),
	string(platform.McpAuthCustom),
}

type StackConfig struct {
	Organization string                    `yaml:"organization" json:"organization"`
	Project      string                    `yaml:"project"      json:"project"`
	StackId      string                    `yaml:"stack-id"     json:"stackId"`
	Services     map[string]ServiceConfig  `yaml:"services"     json:"services"`
	Agents       map[string]AgentConfig    `yaml:"agents"       json:"agents"`
	Databases    map[string]DatabaseConfig `yaml:"databases"    json:"databases"`
	Mcps         map[string]McpConfig      `yaml:"mcps"         json:"mcps"`
	Jobs         map[string]JobConfig      `yaml:"jobs"         json:"jobs"`
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
	Endpoint    bool                   `yaml:"endpoint,omitempty"    json:"endpoint,omitempty"`
	EndpointURL string                 `yaml:"endpointUrl,omitempty" json:"endpointUrl,omitempty"`
	CatalogID   string                 `yaml:"catalogId,omitempty"   json:"catalogId,omitempty"`
	Auth        deployment.McpAuthBody `yaml:"auth,omitempty"        json:"auth"`
	Headers     map[string]string      `yaml:"headers,omitempty"     json:"headers,omitempty"`
}

// JobConfig references script job files by path; loading reads them into Script and Pyproject.
type JobConfig struct {
	Type          string                   `yaml:"type"                    json:"type"`
	Image         *deployment.ImageSpec    `yaml:"image,omitempty"         json:"image,omitempty"`
	Command       []string                 `yaml:"command,omitempty"       json:"command,omitempty"`
	Args          []string                 `yaml:"args,omitempty"          json:"args,omitempty"`
	ScriptFile    string                   `yaml:"scriptFile,omitempty"    json:"scriptFile,omitempty"`
	PyprojectFile string                   `yaml:"pyprojectFile,omitempty" json:"pyprojectFile,omitempty"`
	Script        string                   `yaml:"-"                       json:"-"`
	Pyproject     string                   `yaml:"-"                       json:"-"`
	Resources     deployment.Resources     `yaml:"resources"               json:"resources"`
	Env           []deployment.EnvVar      `yaml:"env,omitempty"           json:"env,omitempty"`
	SecretRefs    []deployment.SecretRef   `yaml:"secretRefs,omitempty"    json:"secretRefs,omitempty"`
	Cron          *deployment.JobCron      `yaml:"cron,omitempty"          json:"cron,omitempty"`
	Timeout       *int64                   `yaml:"timeout,omitempty"       json:"timeout,omitempty"`
	Retries       *int32                   `yaml:"retries,omitempty"       json:"retries,omitempty"`
	Retention     *deployment.JobRetention `yaml:"retention,omitempty"     json:"retention,omitempty"`
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

	if (len(cfg.Services) > 0 || len(cfg.Agents) > 0 || len(cfg.Databases) > 0 || len(cfg.Mcps) > 0 ||
		len(cfg.Jobs) > 0) &&
		cfg.StackId == "" {
		return nil, fmt.Errorf(
			"stack-id is required when services, agents, databases, mcps, or jobs are defined in config file",
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
	// Older internal/external names load as self-hosted/remote so diff matches live.
	for name, mcp := range cfg.Mcps {
		mcp.Type = deployment.McpTypeName(mcp.Type)
		if mcp.Type == deployment.McpTypeRemote {
			if err := validateRemoteMcp(name, mcp); err != nil {
				return nil, err
			}
		}
		cfg.Mcps[name] = mcp
	}

	if cfg.Jobs == nil {
		cfg.Jobs = make(map[string]JobConfig)
	}
	for name, job := range cfg.Jobs {
		job, err := loadJobFiles(job, filepath.Dir(path))
		if err != nil {
			return nil, fmt.Errorf("job %q: %w", name, err)
		}
		cfg.Jobs[name] = job
	}

	return &cfg, nil
}

// loadJobFiles resolves file paths against the config file's directory.
func loadJobFiles(job JobConfig, dir string) (JobConfig, error) {
	if job.Type != "script" {
		if job.ScriptFile != "" || job.PyprojectFile != "" {
			return JobConfig{}, fmt.Errorf(
				"scriptFile and pyprojectFile are only available for script jobs",
			)
		}
		return job, nil
	}
	if job.ScriptFile == "" || job.PyprojectFile == "" {
		return JobConfig{}, fmt.Errorf("scriptFile and pyprojectFile are required for script jobs")
	}
	var err error
	if job.Script, err = inputs.ReadJobFile(resolvePath(dir, job.ScriptFile)); err != nil {
		return JobConfig{}, err
	}
	if job.Pyproject, err = inputs.ReadJobFile(resolvePath(dir, job.PyprojectFile)); err != nil {
		return JobConfig{}, err
	}
	return job, nil
}

func resolvePath(dir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(dir, path)
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
		Type:        deployment.McpTypeValue(m.Type),
		Port:        m.Port,
		Path:        m.Path,
		Image:       m.Image,
		Resources:   m.Resources,
		Env:         m.Env,
		SecretRefs:  m.SecretRefs,
		Endpoint:    m.Endpoint,
		EndpointURL: m.EndpointURL,
		CatalogID:   m.CatalogID,
		Auth:        m.Auth,
		Headers:     m.Headers,
		StackId:     stackId,
	}
}

// validateRemoteMcp rejects what the platform's remote MCP contract cannot carry.
func validateRemoteMcp(name string, mcp McpConfig) error {
	if (mcp.CatalogID == "") == (mcp.EndpointURL == "") {
		return fmt.Errorf("mcp %q: remote mcps need exactly one of catalogId or endpointUrl", name)
	}
	if mcp.CatalogID != "" && (mcp.Auth.Header != "" || mcp.Auth.HeaderPrefix != "") {
		return fmt.Errorf(
			"mcp %q: auth.header and auth.headerPrefix come from the catalog entry", name,
		)
	}
	var selfHostedOnly []string
	for _, field := range []struct {
		key string
		set bool
	}{
		{"port", mcp.Port != 0},
		{"path", mcp.Path != ""},
		{"image", mcp.Image != (deployment.ImageSpec{})},
		{"resources", mcp.Resources != (deployment.Resources{})},
		{"env", len(mcp.Env) > 0},
		{"secretRefs", len(mcp.SecretRefs) > 0},
		{"endpoint", mcp.Endpoint},
	} {
		if field.set {
			selfHostedOnly = append(selfHostedOnly, field.key)
		}
	}
	if len(selfHostedOnly) > 0 {
		return fmt.Errorf(
			"mcp %q: %s not allowed for remote mcps",
			name, strings.Join(selfHostedOnly, ", "),
		)
	}
	if len(mcp.Headers) > 0 {
		return fmt.Errorf("mcp %q: headers are not supported for remote mcps", name)
	}
	if mcp.Auth.Type == "" {
		return fmt.Errorf(
			"mcp %q: auth.type is required for remote mcps (%s)",
			name, strings.Join(stackRemoteAuthTypes, ", "),
		)
	}
	if !slices.Contains(stackRemoteAuthTypes, mcp.Auth.Type) {
		return fmt.Errorf(
			"mcp %q: auth.type %q is not supported in stack files; use 'iai mcps create'",
			name, mcp.Auth.Type,
		)
	}
	custom := mcp.Auth.Type == string(platform.McpAuthCustom)
	if custom && mcp.Auth.Header == "" {
		return fmt.Errorf("mcp %q: auth.header is required for auth.type custom", name)
	}
	if !custom && (mcp.Auth.Header != "" || mcp.Auth.HeaderPrefix != "") {
		return fmt.Errorf(
			"mcp %q: auth.header and auth.headerPrefix require auth.type custom", name,
		)
	}
	return nil
}

// ToPlatformCreateRequest maps a remote MCP to the platform's create body; the platform owns remote MCPs.
func (m McpConfig) ToPlatformCreateRequest(name, stackId string) platform.McpCreateRequest {
	return platform.McpCreateRequest{
		Name:        name,
		Backend:     platform.McpBackendExternal,
		CatalogID:   utils.NilIfZero(m.CatalogID),
		EndpointURL: utils.NilIfZero(m.EndpointURL),
		Transport:   "streamable_http",
		Auth: platform.McpAuth{
			Type:         platform.McpAuthType(m.Auth.Type),
			Credential:   utils.NilIfZero(m.Auth.Credential),
			HeaderName:   utils.NilIfZero(m.Auth.Header),
			HeaderPrefix: utils.NilIfZero(m.Auth.HeaderPrefix),
		},
		StackID: utils.NilIfZero(stackId),
	}
}

func (j JobConfig) ToCreateRequest(stackId string) deployment.CreateJobBody {
	return deployment.CreateJobBody{
		Type:       j.Type,
		Image:      j.Image,
		Command:    j.Command,
		Args:       j.Args,
		Script:     j.Script,
		Pyproject:  j.Pyproject,
		Resources:  j.Resources,
		Env:        j.Env,
		SecretRefs: j.SecretRefs,
		StackId:    stackId,
		Cron:       j.Cron,
		Timeout:    j.Timeout,
		Retries:    j.Retries,
		Retention:  j.Retention,
	}
}

func JobConfigFromRequest(job deployment.CreateJobBody) JobConfig {
	return JobConfig{
		Type:       job.Type,
		Image:      job.Image,
		Command:    job.Command,
		Args:       job.Args,
		Script:     job.Script,
		Pyproject:  job.Pyproject,
		Resources:  job.Resources,
		Env:        job.Env,
		SecretRefs: job.SecretRefs,
		Cron:       job.Cron,
		Timeout:    job.Timeout,
		Retries:    job.Retries,
		Retention:  job.Retention,
	}
}

func ServiceConfigFromDescribe(svc *deployment.DescribeServiceResponse) ServiceConfig {
	return ServiceConfig{
		ServicePort: svc.ServicePort,
		Image:       svc.Image,
		Resources:   svc.Resources,
		Env:         svc.Env,
		SecretRefs:  svc.SecretRefs,
		Endpoint:    svc.Endpoint != nil && svc.Endpoint.Public != "",
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
		Endpoint:    agent.Endpoint != nil && agent.Endpoint.Public != "",
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
	mcpType := deployment.McpTypeName(mcp.Type)
	auth := deployment.McpAuthBody{
		Type:         mcp.Auth.Type,
		Header:       mcp.Auth.Header,
		HeaderPrefix: mcp.Auth.HeaderPrefix,
	}
	if auth.Type != string(platform.McpAuthCustom) {
		auth.Header, auth.HeaderPrefix = "", ""
	}
	if mcpType == deployment.McpTypeRemote {
		// Only what a remote mcp may declare, so the export loads back: a missing auth type reads as none and header routing belongs to a custom credential.
		auth.Type = cmp.Or(mcp.Auth.Type, string(platform.McpAuthNone))
		if mcp.CatalogID != "" {
			auth.Header, auth.HeaderPrefix = "", ""
		}
		if mcp.CatalogID != "" {
			return McpConfig{Type: mcpType, CatalogID: mcp.CatalogID, Auth: auth}
		}
		return McpConfig{Type: mcpType, EndpointURL: mcp.EndpointURL, Auth: auth}
	}
	return McpConfig{
		Type:       mcpType,
		Port:       mcp.Port,
		Path:       mcp.Path,
		Image:      mcp.Image,
		Resources:  mcp.Resources,
		Env:        mcp.Env,
		SecretRefs: mcp.SecretRefs,
		Endpoint:   mcp.Endpoint != nil && mcp.Endpoint.Public != "",
		Auth:       auth,
		Headers:    mcp.Headers,
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

// StackManagedRemote reports whether a stack file can express a remote MCP; oauth and client_credentials ones are managed with 'iai mcps' only.
func StackManagedRemote(mcp platform.McpSchema) bool {
	// A missing auth type reads as none, which is what a stack file sends for it.
	return mcp.AuthType == nil || slices.Contains(stackRemoteAuthTypes, *mcp.AuthType)
}

// McpConfigFromPlatform maps a remote MCP the platform owns; a catalog entry supplies its own endpoint and headers, so only its id is kept.
func McpConfigFromPlatform(mcp platform.McpSchema) McpConfig {
	// A missing auth type reads as none, the default StackManagedRemote applies.
	cfg := McpConfig{
		Type: deployment.McpTypeRemote,
		Auth: deployment.McpAuthBody{Type: string(platform.McpAuthNone)},
	}
	if mcp.AuthType != nil && *mcp.AuthType != "" {
		cfg.Auth.Type = *mcp.AuthType
	}
	if mcp.CatalogID != nil && *mcp.CatalogID != "" {
		cfg.CatalogID = *mcp.CatalogID
		return cfg
	}
	if mcp.EndpointURL != nil {
		cfg.EndpointURL = *mcp.EndpointURL
	}
	// Header routing belongs to a custom credential; the other types imply their own.
	if cfg.Auth.Type == string(platform.McpAuthCustom) {
		if mcp.AuthHeaderName != nil {
			cfg.Auth.Header = *mcp.AuthHeaderName
		}
		if mcp.AuthHeaderPrefix != nil {
			cfg.Auth.HeaderPrefix = *mcp.AuthHeaderPrefix
		}
	}
	return cfg
}

func FetchLiveStack(
	ctx context.Context,
	deployClient *deployment.DeploymentClient,
	apiClient *platform.APIClient,
	orgId, projectId, stackId string,
) (*StackConfig, error) {
	cfg := &StackConfig{
		StackId:   stackId,
		Services:  make(map[string]ServiceConfig),
		Agents:    make(map[string]AgentConfig),
		Databases: make(map[string]DatabaseConfig),
		Mcps:      make(map[string]McpConfig),
		Jobs:      make(map[string]JobConfig),
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
	remoteMcps, err := apiClient.ListRemoteMcps(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote mcps: %w", err)
	}
	for _, mcp := range remoteMcps {
		if StackManagedRemote(mcp) {
			cfg.Mcps[mcp.Name] = McpConfigFromPlatform(mcp)
		}
	}

	jobs, err := deployClient.ListJobs(ctx, orgId, projectId, stackId)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	for _, job := range jobs {
		desc, err := deployClient.DescribeJob(ctx, orgId, projectId, job.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to describe job %q: %w", job.Name, err)
		}
		cfg.Jobs[job.Name] = JobConfigFromRequest(desc.CreateJobBody)
	}

	return cfg, nil
}

// WriteJobFiles writes script job files under dir/jobs/<name>/ and points the config at them.
func WriteJobFiles(cfg *StackConfig, dir string) error {
	for name, job := range cfg.Jobs {
		if job.Type != "script" {
			continue
		}
		jobDir := filepath.Join("jobs", name)
		if err := os.MkdirAll(filepath.Join(dir, jobDir), 0o755); err != nil {
			return fmt.Errorf("failed to create %s: %w", jobDir, err)
		}
		job.ScriptFile = filepath.Join(jobDir, "main.py")
		job.PyprojectFile = filepath.Join(jobDir, "pyproject.toml")
		for path, contents := range map[string]string{
			job.ScriptFile:    job.Script,
			job.PyprojectFile: job.Pyproject,
		} {
			if err := os.WriteFile(filepath.Join(dir, path), []byte(contents), 0o644); err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}
		}
		cfg.Jobs[name] = job
	}
	return nil
}

// ScriptJobNames lists script jobs, whose files only WriteJobFiles exports.
func ScriptJobNames(cfg *StackConfig) []string {
	var names []string
	for name, job := range cfg.Jobs {
		if job.Type == "script" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func MarshalStackConfig(cfg *StackConfig) ([]byte, error) {
	return yaml.Marshal(cfg)
}
