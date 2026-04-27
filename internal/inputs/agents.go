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

	reqBody := clients.CreateAgentBody{
		Id:          in.Id,
		Version:     in.Version,
		AgentConfig: agentConfig,
		SecretRefs:  secretRefs,
		Endpoint:    in.Endpoint,
		Env:         env,
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
}

// BuildAgentUpdatePatch produces a partial-update body containing only the
// fields whose flags the user explicitly set. `changed` reports whether a flag
// name was provided on the command line (typically cmd.Flags().Changed).
func BuildAgentUpdatePatch(
	in AgentInput,
	clearEnv, clearSecret, clearSchedule bool,
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

	return patch, nil
}
