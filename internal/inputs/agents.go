package inputs

import (
	"fmt"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
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

	McpRefs        []McpRef // the one attachment --mcp asks for, resolved from the mcp's release at deploy time
	DetachMcpNames []string // removed before McpRefs are (re-)injected
}

// McpRef is one attachment: the mcp, and the prefix this agent calls its tools by.
type McpRef struct {
	Name string
	Id   string // empty means the mcp's name
	// SetId tells --mcp-id naming the mcp itself, which clears a prefix, from no
	// --mcp-id at all, which leaves whatever is attached alone.
	SetId bool
}

// McpRefsFor builds the attachment --mcp asks for, with the prefix --mcp-id gives it.
// One attachment per command: a prefix describes exactly one, and so does a detach.
func McpRefsFor(names []string, id string, idGiven bool) ([]McpRef, error) {
	if len(names) > 1 {
		return nil, fmt.Errorf(
			"--mcp attaches one mcp (got %d); attach further mcps in their own commands",
			len(names),
		)
	}
	id = strings.TrimSpace(id)
	if idGiven {
		if id == "" {
			return nil, fmt.Errorf(
				"--mcp-id must name the prefix the agent calls the mcp's tools by",
			)
		}
		if len(names) == 0 {
			return nil, fmt.Errorf(
				"--mcp-id sets the prefix for the mcp --mcp attaches; pass --mcp <name> too",
			)
		}
	}
	refs := make([]McpRef, 0, len(names))
	for _, name := range names {
		if name = strings.TrimSpace(name); name == "" {
			return nil, fmt.Errorf("--mcp must name an mcp")
		}
		prefix := id
		if prefix == name {
			prefix = "" // the default, so the config keeps its short form
		}
		refs = append(refs, McpRef{Name: name, Id: prefix, SetId: idGiven})
	}
	return refs, nil
}

// mcpEntryName is the mcp an existing entry attaches: "id" is only the tool prefix once "ref" is set.
func mcpEntryName(entry any) string {
	switch e := entry.(type) {
	case string:
		return e
	case map[string]any:
		if ref, ok := e["ref"].(string); ok && ref != "" {
			return ref
		}
		id, _ := e["id"].(string)
		return id
	}
	return ""
}

func mcpEntry(ref McpRef) any {
	if ref.Id == "" {
		return ref.Name
	}
	return map[string]any{"ref": ref.Name, "id": ref.Id}
}

// InjectMcpRefs attaches each ref to the agent config's mcps list, preserving existing entries.
// One already attached keeps its prefix unless this call gives one, so no unrelated update resets it.
func InjectMcpRefs(agentConfig any, refs []McpRef) (any, error) {
	if len(refs) == 0 {
		return agentConfig, nil
	}

	cfg, ok := agentConfig.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("agent config must be a YAML/JSON object to attach --mcp references")
	}
	mcps, _ := cfg["mcps"].([]any)

	at := make(map[string]int, len(mcps))
	for i, entry := range mcps {
		if name := mcpEntryName(entry); name != "" {
			at[name] = i
		}
	}
	for _, ref := range refs {
		i, attached := at[ref.Name]
		if !attached {
			at[ref.Name] = len(mcps)
			mcps = append(mcps, mcpEntry(ref))
			continue
		}
		if !ref.SetId {
			continue
		}
		if existing, isMap := mcps[i].(map[string]any); isMap {
			if attaches, _ := existing["ref"].(string); attaches == "" {
				return nil, fmt.Errorf(
					"mcp %q is configured in full in this agent's config; set its prefix there, not with --mcp",
					ref.Name,
				)
			}
		}
		mcps[i] = mcpEntry(ref)
	}
	cfg["mcps"] = mcps
	return cfg, nil
}

// DetachMcpRefs removes any mcps entry matching one of the given names; names not present are ignored.
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
		if names[mcpEntryName(entry)] {
			continue
		}
		kept = append(kept, entry)
	}
	cfg["mcps"] = kept
	return cfg, nil
}

func BuildAgentRequestBody(in AgentInput) (deployment.CreateAgentBody, error) {
	if err := ValidateServiceEnvVars(in.EnvVars); err != nil {
		return deployment.CreateAgentBody{}, err
	}
	var env []deployment.EnvVar
	for _, e := range in.EnvVars {
		parts := strings.SplitN(e, "=", 2)
		env = append(env, deployment.EnvVar{
			Name:  strings.TrimSpace(parts[0]),
			Value: parts[1],
		})
	}

	if err := ValidateServiceSecretRefs(in.SecretRefs); err != nil {
		return deployment.CreateAgentBody{}, err
	}
	var secretRefs []deployment.SecretRef
	for _, name := range in.SecretRefs {
		secretRefs = append(secretRefs, deployment.SecretRef{
			SecretName: strings.TrimSpace(name),
		})
	}

	data, err := os.ReadFile(in.FilePath)
	if err != nil {
		return deployment.CreateAgentBody{}, fmt.Errorf(
			"failed to read file %q: %w",
			in.FilePath,
			err,
		)
	}

	var agentConfig any
	if err := yaml.Unmarshal(data, &agentConfig); err != nil {
		return deployment.CreateAgentBody{}, fmt.Errorf(
			"failed to parse YAML from %q: %w",
			in.FilePath,
			err,
		)
	}
	agentConfig, err = InjectMcpRefs(agentConfig, in.McpRefs)
	if err != nil {
		return deployment.CreateAgentBody{}, err
	}

	reqBody := deployment.CreateAgentBody{
		Id:          in.Id,
		Version:     in.Version,
		AgentConfig: agentConfig,
		SecretRefs:  secretRefs,
		Endpoint:    in.Endpoint,
		Env:         env,
		StackId:     in.StackId,
	}

	if in.ScheduleUptime != "" || in.ScheduleDowntime != "" || in.ScheduleTimezone != "" {
		reqBody.Schedule = &deployment.Schedule{
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
) (deployment.UpdatePatch, error) {
	f := AgentUpdateFlags
	patch := deployment.UpdatePatch{}

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
		agentConfig, err = InjectMcpRefs(agentConfig, in.McpRefs)
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
