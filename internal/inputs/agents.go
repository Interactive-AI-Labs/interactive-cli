package inputs

import (
	"fmt"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"gopkg.in/yaml.v3"
)

type AgentInput struct {
	Id       string
	Version  string
	FilePath string
	Endpoint bool

	EnvVars    []string
	SecretRefs []string

	ScheduleUptime   string
	ScheduleDowntime string
	ScheduleTimezone string

	StackId string

	// McpNames are attached as bare {mcp_id: <name>} references, appended to
	// whatever mcps entries the file already declares. The operator resolves
	// each reference from the mcp's own release at deploy time.
	McpNames []string

	// DetachMcpNames removes any existing mcps entry (bare ref or resolved)
	// matching one of these names, applied before McpNames are (re-)injected.
	DetachMcpNames []string
}

// InjectMcpRefs appends a bare-string reference for each name to the agent
// config's mcps list (creating the key if absent). Existing entries — whether
// bare refs or fully hand-written McpConfig blocks — are preserved.
func InjectMcpRefs(agentConfig any, mcpNames []string) (any, error) {
	names := make([]string, 0, len(mcpNames))
	for _, n := range mcpNames {
		if n = strings.TrimSpace(n); n != "" {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return agentConfig, nil
	}

	cfg, ok := agentConfig.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("agent config must be a YAML/JSON object to attach --mcp references")
	}
	mcps, _ := cfg["mcps"].([]any)

	// An mcp already present — as a bare-string ref or a resolved entry
	// (a map with an id matching the mcp's slug) — is skipped rather than
	// duplicated. Also covers passing the same --mcp name twice in one command.
	seen := make(map[string]bool, len(mcps)+len(names))
	for _, entry := range mcps {
		switch e := entry.(type) {
		case string:
			if e != "" {
				seen[e] = true
			}
		case map[string]any:
			if id, ok := e["id"].(string); ok && id != "" {
				seen[id] = true
			}
		}
	}
	for _, name := range names {
		if seen[name] {
			continue
		}
		mcps = append(mcps, name)
		seen[name] = true
	}
	cfg["mcps"] = mcps
	return cfg, nil
}

// DetachMcpRefs removes any mcps entry matching one of the given names — by
// bare-string ref or resolved entry (a map with a matching id) — from the agent
// config. Names not present are ignored, so it's safe to call speculatively.
func DetachMcpRefs(agentConfig any, mcpNames []string) (any, error) {
	names := make(map[string]bool, len(mcpNames))
	for _, n := range mcpNames {
		if n = strings.TrimSpace(n); n != "" {
			names[n] = true
		}
	}
	if len(names) == 0 {
		return agentConfig, nil
	}

	cfg, ok := agentConfig.(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"agent config must be a YAML/JSON object to detach --detach-mcp references",
		)
	}
	mcps, _ := cfg["mcps"].([]any)

	kept := make([]any, 0, len(mcps))
	for _, entry := range mcps {
		switch e := entry.(type) {
		case string:
			if names[e] {
				continue
			}
		case map[string]any:
			if id, _ := e["id"].(string); names[id] {
				continue
			}
		}
		kept = append(kept, entry)
	}
	cfg["mcps"] = kept
	return cfg, nil
}

func BuildAgentRequestBody(in AgentInput) (clients.CreateAgentBody, error) {
	if err := ValidateServiceEnvVars(in.EnvVars); err != nil {
		return clients.CreateAgentBody{}, err
	}
	var env []clients.EnvVar
	for _, e := range in.EnvVars {
		parts := strings.SplitN(e, "=", 2)
		env = append(env, clients.EnvVar{
			Name:  strings.TrimSpace(parts[0]),
			Value: parts[1],
		})
	}

	if err := ValidateServiceSecretRefs(in.SecretRefs); err != nil {
		return clients.CreateAgentBody{}, err
	}
	var secretRefs []clients.SecretRef
	for _, name := range in.SecretRefs {
		secretRefs = append(secretRefs, clients.SecretRef{
			SecretName: strings.TrimSpace(name),
		})
	}

	data, err := os.ReadFile(in.FilePath)
	if err != nil {
		return clients.CreateAgentBody{}, fmt.Errorf("failed to read file %q: %w", in.FilePath, err)
	}

	var agentConfig any
	if err := yaml.Unmarshal(data, &agentConfig); err != nil {
		return clients.CreateAgentBody{}, fmt.Errorf(
			"failed to parse YAML from %q: %w",
			in.FilePath,
			err,
		)
	}
	agentConfig, err = InjectMcpRefs(agentConfig, in.McpNames)
	if err != nil {
		return clients.CreateAgentBody{}, err
	}

	reqBody := clients.CreateAgentBody{
		Id:          in.Id,
		Version:     in.Version,
		AgentConfig: agentConfig,
		SecretRefs:  secretRefs,
		Endpoint:    in.Endpoint,
		Env:         env,
		StackId:     in.StackId,
	}

	if in.ScheduleUptime != "" || in.ScheduleDowntime != "" || in.ScheduleTimezone != "" {
		reqBody.Schedule = &clients.Schedule{
			Uptime:   in.ScheduleUptime,
			Downtime: in.ScheduleDowntime,
			Timezone: in.ScheduleTimezone,
		}
	}

	return reqBody, nil
}

// AgentUpdateFlags is the set of cobra flag names BuildAgentUpdatePatch
// inspects via the `changed` predicate. Keep in sync with cmd/agents.go.
var AgentUpdateFlags = struct {
	Id               string
	Version          string
	File             string
	Endpoint         string
	Env              string
	Secret           string
	ScheduleUptime   string
	ScheduleDowntime string
	ScheduleTimezone string
	StackId          string
}{
	Id:               "id",
	Version:          "version",
	File:             "file",
	Endpoint:         "endpoint",
	Env:              "env",
	Secret:           "secret",
	ScheduleUptime:   "schedule-uptime",
	ScheduleDowntime: "schedule-downtime",
	ScheduleTimezone: "schedule-timezone",
	StackId:          "stack-id",
}

// BuildAgentUpdatePatch produces a partial-update body containing only the
// fields whose flags the user explicitly set. `changed` reports whether a flag
// name was provided on the command line (typically cmd.Flags().Changed).
func BuildAgentUpdatePatch(
	in AgentInput,
	clearEnv, clearSecret, clearSchedule, clearStackId bool,
	changed func(string) bool,
) (clients.UpdatePatch, error) {
	f := AgentUpdateFlags
	patch := clients.UpdatePatch{}

	if changed(f.Id) {
		if err := setJSON(patch, "id", in.Id); err != nil {
			return nil, err
		}
	}
	if changed(f.Version) {
		if err := setJSON(patch, "version", in.Version); err != nil {
			return nil, err
		}
	}

	if changed(f.File) {
		data, err := os.ReadFile(in.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", in.FilePath, err)
		}
		var agentConfig any
		if err := yaml.Unmarshal(data, &agentConfig); err != nil {
			return nil, fmt.Errorf(
				"failed to parse YAML from %q: %w", in.FilePath, err,
			)
		}
		agentConfig, err = DetachMcpRefs(agentConfig, in.DetachMcpNames)
		if err != nil {
			return nil, err
		}
		agentConfig, err = InjectMcpRefs(agentConfig, in.McpNames)
		if err != nil {
			return nil, err
		}
		if err := setJSON(patch, "agentConfig", agentConfig); err != nil {
			return nil, err
		}
	}

	if err := setEnvPatch(patch, in.EnvVars, changed(f.Env), clearEnv); err != nil {
		return nil, err
	}
	if err := setSecretRefsPatch(patch, in.SecretRefs, changed(f.Secret), clearSecret); err != nil {
		return nil, err
	}
	if err := setEndpointPatch(patch, in.Endpoint, changed(f.Endpoint)); err != nil {
		return nil, err
	}
	if err := setSchedulePatch(patch, ScheduleInput{
		Uptime:          in.ScheduleUptime,
		Downtime:        in.ScheduleDowntime,
		Timezone:        in.ScheduleTimezone,
		UptimeChanged:   changed(f.ScheduleUptime),
		DowntimeChanged: changed(f.ScheduleDowntime),
		TimezoneChanged: changed(f.ScheduleTimezone),
		Clear:           clearSchedule,
	}); err != nil {
		return nil, err
	}

	if err := setStackIdPatch(patch, in.StackId, changed(f.StackId), clearStackId); err != nil {
		return nil, err
	}

	return patch, nil
}
