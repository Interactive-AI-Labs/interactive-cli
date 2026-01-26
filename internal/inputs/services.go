package inputs

import (
	"fmt"
	"regexp"
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

func ValidateService(input ServiceInput) error {
	if input.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if input.Port <= 0 {
		return fmt.Errorf("service port must be greater than zero; please provide --port")
	}
	if input.ImageName == "" {
		return fmt.Errorf("image name is required; please provide --image-name")
	}
	if input.ImageTag == "" {
		return fmt.Errorf("image tag is required; please provide --image-tag")
	}
	if input.ImageType == "" {
		return fmt.Errorf("image type is required; please provide --image-type")
	}
	if input.ImageType != "internal" && input.ImageType != "external" {
		return fmt.Errorf("image type must be either 'internal' or 'external'; please provide --image-type")
	}
	if input.ImageType == "external" && input.ImageRepository == "" {
		return fmt.Errorf("image repository is required for external images; please provide --image-repository")
	}
	if err := ValidateMemory(input.Memory); err != nil {
		return fmt.Errorf("%w; please provide --memory", err)
	}
	if err := ValidateCPU(input.CPU); err != nil {
		return fmt.Errorf("%w; please provide --cpu", err)
	}

	hasReplicas := input.Replicas > 0
	hasAutoscaling := input.Autoscaling != nil && input.Autoscaling.Enabled

	if hasReplicas && hasAutoscaling {
		return fmt.Errorf("cannot specify both --replicas and --autoscaling-enabled; they are mutually exclusive")
	}

	if !hasReplicas && !hasAutoscaling {
		return fmt.Errorf("must specify either --replicas or --autoscaling-enabled")
	}

	if hasAutoscaling {
		if err := ValidateAutoscaling(*input.Autoscaling); err != nil {
			return err
		}
	}

	return nil
}

func ValidateAutoscaling(config AutoscalingInput) error {
	if config.MinReplicas <= 0 {
		return fmt.Errorf("--autoscaling-min-replicas must be greater than zero when autoscaling is enabled")
	}
	if config.MaxReplicas <= 0 {
		return fmt.Errorf("--autoscaling-max-replicas must be greater than zero when autoscaling is enabled")
	}
	if config.MinReplicas > config.MaxReplicas {
		return fmt.Errorf("--autoscaling-min-replicas cannot be greater than --autoscaling-max-replicas")
	}
	if config.CPUPercentage <= 0 && config.MemoryPercent <= 0 {
		return fmt.Errorf("at least one of --autoscaling-cpu-percentage or --autoscaling-memory-percentage must be set when autoscaling is enabled")
	}
	return nil
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

// cpuPattern matches whole numbers (e.g., "1", "2", "4") or millicores (e.g., "500m", "1000m")
var cpuPattern = regexp.MustCompile(`^(\d+|\d+m)$`)

// memoryPattern matches memory values with M or G units (e.g., "128M", "512M", "1G", "2G")
var memoryPattern = regexp.MustCompile(`^\d+(M|G)$`)

// ValidateCPU validates a CPU value as a whole number of cores or millicores.
// Valid formats: "1", "2", "4", "500m", "1000m"
func ValidateCPU(cpu string) error {
	if cpu == "" {
		return fmt.Errorf("cpu is required")
	}
	if !cpuPattern.MatchString(cpu) {
		return fmt.Errorf("invalid cpu value %q; expected a whole number of cores (e.g., '1', '2') or millicores (e.g., '500m', '1000m')", cpu)
	}
	return nil
}

// ValidateMemory validates a memory value with M or G units.
// Valid formats: "128M", "512M", "1G", "2G"
func ValidateMemory(memory string) error {
	if memory == "" {
		return fmt.Errorf("memory is required")
	}
	if !memoryPattern.MatchString(memory) {
		return fmt.Errorf("invalid memory value %q; expected a value with M or G unit (e.g., '128M', '512M', '1G')", memory)
	}
	return nil
}
