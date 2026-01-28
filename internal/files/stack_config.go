package files

import (
	"fmt"
	"os"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
	"gopkg.in/yaml.v3"
)

type StackConfig struct {
	Organization string                   `yaml:"organization"`
	Project      string                   `yaml:"project"`
	StackId      string                   `yaml:"stack-id"`
	Services     map[string]ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	Version     string               `yaml:"version,omitempty"`
	ServicePort int                  `yaml:"servicePort"`
	Image       clients.ImageSpec    `yaml:"image"`
	Resources   clients.Resources    `yaml:"resources"`
	Env         []clients.EnvVar     `yaml:"env,omitempty"`
	SecretRefs  []clients.SecretRef  `yaml:"secretRefs,omitempty"`
	Endpoint    bool                 `yaml:"endpoint,omitempty"`
	Replicas    int                  `yaml:"replicas,omitempty"`
	Autoscaling *clients.Autoscaling `yaml:"autoscaling,omitempty"`
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

	if len(cfg.Services) > 0 && cfg.StackId == "" {
		return nil, fmt.Errorf("stack-id is required when services are defined in config file")
	}

	if cfg.Services == nil {
		cfg.Services = make(map[string]ServiceConfig)
	}

	for name, svc := range cfg.Services {
		if name == "" {
			return nil, fmt.Errorf("service name cannot be empty")
		}
		if svc.ServicePort <= 0 {
			return nil, fmt.Errorf("service %q: servicePort must be greater than zero", name)
		}
		if svc.Image.Type == "" {
			return nil, fmt.Errorf("service %q: image.type is required", name)
		}
		if svc.Image.Type != "internal" && svc.Image.Type != "external" {
			return nil, fmt.Errorf("service %q: image.type must be 'internal' or 'external'", name)
		}
		if svc.Image.Type == "external" && svc.Image.Repository == "" {
			return nil, fmt.Errorf("service %q: image.repository is required for external images", name)
		}
		if svc.Image.Name == "" {
			return nil, fmt.Errorf("service %q: image.name is required", name)
		}
		if svc.Image.Tag == "" {
			return nil, fmt.Errorf("service %q: image.tag is required", name)
		}

		if err := inputs.ValidateMemory(svc.Resources.Memory); err != nil {
			return nil, fmt.Errorf("service %q: resources.memory: %w", name, err)
		}
		if err := inputs.ValidateCPU(svc.Resources.CPU); err != nil {
			return nil, fmt.Errorf("service %q: resources.cpu: %w", name, err)
		}

		hasReplicas := svc.Replicas > 0
		hasAutoscaling := svc.Autoscaling != nil && svc.Autoscaling.Enabled

		if hasReplicas && hasAutoscaling {
			return nil, fmt.Errorf("service %q: cannot set both replicas and autoscaling.enabled; only one scaling method can be configured", name)
		}

		if !hasReplicas && !hasAutoscaling {
			return nil, fmt.Errorf("service %q: must specify either replicas or autoscaling", name)
		}

		if svc.Autoscaling != nil && svc.Autoscaling.Enabled {
			if svc.Autoscaling.MinReplicas <= 0 {
				return nil, fmt.Errorf("service %q: autoscaling.minReplicas must be greater than zero when autoscaling is enabled", name)
			}
			if svc.Autoscaling.MaxReplicas <= 0 {
				return nil, fmt.Errorf("service %q: autoscaling.maxReplicas must be greater than zero when autoscaling is enabled", name)
			}
			if svc.Autoscaling.MinReplicas > svc.Autoscaling.MaxReplicas {
				return nil, fmt.Errorf("service %q: autoscaling.minReplicas cannot be greater than autoscaling.maxReplicas", name)
			}
			if svc.Autoscaling.CPUPercentage <= 0 && svc.Autoscaling.MemoryPercentage <= 0 {
				return nil, fmt.Errorf("service %q: at least one of autoscaling.cpuPercentage or autoscaling.memoryPercentage must be set when autoscaling is enabled", name)
			}
		}

		for j, env := range svc.Env {
			if env.Name == "" {
				return nil, fmt.Errorf("service %q: env[%d].name is required", name, j)
			}
		}

		for j, ref := range svc.SecretRefs {
			if ref.SecretName == "" {
				return nil, fmt.Errorf("service %q: secretRefs[%d].secretName is required", name, j)
			}
		}
	}

	return &cfg, nil
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

	if s.Autoscaling != nil && s.Autoscaling.Enabled {
		body.Autoscaling = s.Autoscaling
	} else {
		body.Replicas = s.Replicas
	}

	return body
}
