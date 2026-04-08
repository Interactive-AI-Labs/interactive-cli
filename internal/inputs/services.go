package inputs

import (
	"fmt"
	"strings"
)

type ServiceInput struct {
	Name            string
	Port            int
	ImageType       string
	ImageRepository string
	ImageName       string
	ImageTag        string
	Memory          string
	CPU             string
	Replicas        int
	Autoscaling     *AutoscalingInput
}

type AutoscalingInput struct {
	Enabled       bool
	MinReplicas   int
	MaxReplicas   int
	CPUPercentage int
	MemoryPercent int
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
