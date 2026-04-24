package inputs

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
)

type ServiceInput struct {
	Port            int
	ImageType       string
	ImageRepository string
	ImageName       string
	ImageTag        string
	Memory          string
	CPU             string
	Endpoint        bool

	Replicas          int
	AutoscalingMin    int
	AutoscalingMax    int
	AutoscalingCPU    int
	AutoscalingMemory int

	EnvVars    []string
	SecretRefs []string

	HealthcheckPath         string
	HealthcheckInitialDelay int

	ScheduleUptime   string
	ScheduleDowntime string
	ScheduleTimezone string
}

func BuildServiceRequestBody(in ServiceInput) (clients.CreateServiceBody, error) {
	if err := ValidateServiceEnvVars(in.EnvVars); err != nil {
		return clients.CreateServiceBody{}, err
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
		return clients.CreateServiceBody{}, err
	}
	var secretRefs []clients.SecretRef
	for _, name := range in.SecretRefs {
		secretRefs = append(secretRefs, clients.SecretRef{
			SecretName: strings.TrimSpace(name),
		})
	}

	reqBody := clients.CreateServiceBody{
		ServicePort: in.Port,
		Image: clients.ImageSpec{
			Type:       in.ImageType,
			Repository: in.ImageRepository,
			Name:       in.ImageName,
			Tag:        in.ImageTag,
		},
		Resources: clients.Resources{
			Memory: in.Memory,
			CPU:    in.CPU,
		},
		Env:        env,
		SecretRefs: secretRefs,
		Endpoint:   in.Endpoint,
	}

	if in.AutoscalingMin > 0 || in.AutoscalingMax > 0 || in.AutoscalingCPU > 0 ||
		in.AutoscalingMemory > 0 {
		as := &clients.Autoscaling{
			MinReplicas: in.AutoscalingMin,
			MaxReplicas: in.AutoscalingMax,
		}
		if in.AutoscalingCPU > 0 {
			as.CPUPercentage = utils.ToPtr(in.AutoscalingCPU)
		}
		if in.AutoscalingMemory > 0 {
			as.MemoryPercentage = utils.ToPtr(in.AutoscalingMemory)
		}
		reqBody.Autoscaling = as
	}
	if in.Replicas > 0 {
		reqBody.Replicas = in.Replicas
	}

	if in.HealthcheckPath != "" || in.HealthcheckInitialDelay != 0 {
		hc := &clients.Healthcheck{
			Path: in.HealthcheckPath,
		}
		if in.HealthcheckInitialDelay > 0 {
			hc.InitialDelaySeconds = utils.ToPtr(in.HealthcheckInitialDelay)
		}
		reqBody.Healthcheck = hc
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

func ValidateServiceEnvVars(envVars []string) error {
	for _, e := range envVars {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return fmt.Errorf("invalid --env value %q; expected NAME=VALUE", e)
		}
	}
	return nil
}

func ValidateServiceSecretRefs(secretRefs []string) error {
	for _, name := range secretRefs {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return fmt.Errorf("invalid --secret value %q; name must not be empty", name)
		}
	}
	return nil
}
