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
