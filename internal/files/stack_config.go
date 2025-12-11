package files

import (
	"fmt"
	"os"

	clients "github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"gopkg.in/yaml.v3"
)

type StackConfig struct {
	Organization string                   `yaml:"organization"`
	Project      string                   `yaml:"project"`
	StackId      string                   `yaml:"stack-id"`
	Services     map[string]ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	Version     string              `yaml:"version,omitempty"`
	ServicePort int                 `yaml:"servicePort"`
	Image       clients.ImageSpec   `yaml:"image"`
	Resources   clients.Resources   `yaml:"resources"`
	Env         []clients.EnvVar    `yaml:"env,omitempty"`
	SecretRefs  []clients.SecretRef `yaml:"secretRefs,omitempty"`
	Endpoint    bool                `yaml:"endpoint,omitempty"`
	Replicas    int                 `yaml:"replicas"`
}

func LoadStackConfig(path string) (*StackConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg StackConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// If services are provided, stack-id is required
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
		if svc.Replicas < 1 {
			return nil, fmt.Errorf("service %q: replicas must be at least 1", name)
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
	return clients.CreateServiceBody{
		ServicePort: s.ServicePort,
		Image:       s.Image,
		Resources:   s.Resources,
		Env:         s.Env,
		SecretRefs:  s.SecretRefs,
		Endpoint:    s.Endpoint,
		Replicas:    s.Replicas,
		StackId:     stackId,
	}
}
