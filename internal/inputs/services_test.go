package inputs

import (
	"strings"
	"testing"
)

func TestValidateCPU(t *testing.T) {
	tests := []struct {
		name            string
		cpu             string
		wantErr         bool
		wantErrContains string
	}{
		{name: "valid whole number", cpu: "1", wantErr: false},
		{name: "valid decimal", cpu: "0.5", wantErr: false},
		{name: "valid millicores", cpu: "500m", wantErr: false},
		{name: "empty string", cpu: "", wantErr: true, wantErrContains: "cpu is required"},
		{name: "invalid text", cpu: "abc", wantErr: true, wantErrContains: "invalid cpu value"},
		{name: "invalid zero", cpu: "0", wantErr: true, wantErrContains: "invalid cpu value"},
		{name: "invalid zero millicores", cpu: "0m", wantErr: true, wantErrContains: "invalid cpu value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCPU(tt.cpu)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateCPU() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("ValidateCPU() error = %v, want error containing %q", err, tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateCPU() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateMemory(t *testing.T) {
	tests := []struct {
		name            string
		memory          string
		wantErr         bool
		wantErrContains string
	}{
		{name: "valid M unit", memory: "512M", wantErr: false},
		{name: "valid G unit", memory: "1G", wantErr: false},
		{name: "empty string", memory: "", wantErr: true, wantErrContains: "memory is required"},
		{name: "invalid no unit", memory: "512", wantErr: true, wantErrContains: "invalid memory value"},
		{name: "invalid text", memory: "abc", wantErr: true, wantErrContains: "invalid memory value"},
		{name: "invalid zero", memory: "0M", wantErr: true, wantErrContains: "invalid memory value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMemory(tt.memory)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateMemory() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("ValidateMemory() error = %v, want error containing %q", err, tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateMemory() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateService(t *testing.T) {
	tests := []struct {
		name            string
		input           ServiceInput
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "valid with replicas",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  3,
			},
			wantErr: false,
		},
		{
			name: "valid with autoscaling",
			input: ServiceInput{
				Name:            "my-service",
				Port:            8080,
				ImageType:       "external",
				ImageRepository: "docker.io/myrepo",
				ImageName:       "my-image",
				ImageTag:        "1.0.0",
				Memory:          "512M",
				CPU:             "1",
				Autoscaling: &AutoscalingInput{
					Enabled:       true,
					MinReplicas:   2,
					MaxReplicas:   10,
					CPUPercentage: 80,
					MemoryPercent: 70,
				},
			},
			wantErr: false,
		},
		{
			name: "valid with autoscaling CPU only",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Autoscaling: &AutoscalingInput{
					Enabled:       true,
					MinReplicas:   1,
					MaxReplicas:   5,
					CPUPercentage: 75,
				},
			},
			wantErr: false,
		},
		{
			name: "valid with autoscaling Memory only",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Autoscaling: &AutoscalingInput{
					Enabled:       true,
					MinReplicas:   1,
					MaxReplicas:   5,
					MemoryPercent: 80,
				},
			},
			wantErr: false,
		},
		{
			name: "empty service name",
			input: ServiceInput{
				Name:      "",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "service name is required",
		},
		{
			name: "port zero",
			input: ServiceInput{
				Name:      "my-service",
				Port:      0,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "service port must be greater than zero",
		},
		{
			name: "negative port",
			input: ServiceInput{
				Name:      "my-service",
				Port:      -1,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "service port must be greater than zero",
		},
		{
			name: "empty image name",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "image name is required",
		},
		{
			name: "empty image tag",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "image tag is required",
		},
		{
			name: "empty image type",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "image type is required",
		},
		{
			name: "invalid image type",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "invalid",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "image type must be either 'internal' or 'external'",
		},
		{
			name: "external image without repository",
			input: ServiceInput{
				Name:            "my-service",
				Port:            8080,
				ImageType:       "external",
				ImageRepository: "",
				ImageName:       "my-image",
				ImageTag:        "1.0.0",
				Memory:          "512M",
				CPU:             "1",
				Replicas:        1,
			},
			wantErr:         true,
			wantErrContains: "image repository is required for external images",
		},
		{
			name: "empty memory",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "memory is required",
		},
		{
			name: "empty cpu",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "cpu is required",
		},
		{
			name: "invalid cpu format",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "500Mi",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "invalid cpu value",
		},
		{
			name: "invalid memory format",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "invalid",
				CPU:       "1",
				Replicas:  1,
			},
			wantErr:         true,
			wantErrContains: "invalid memory value",
		},
		{
			name: "both replicas and autoscaling",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  3,
				Autoscaling: &AutoscalingInput{
					Enabled:       true,
					MinReplicas:   2,
					MaxReplicas:   10,
					CPUPercentage: 80,
				},
			},
			wantErr:         true,
			wantErrContains: "cannot specify both --replicas and --autoscaling-enabled",
		},
		{
			name: "neither replicas nor autoscaling",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Replicas:  0,
			},
			wantErr:         true,
			wantErrContains: "must specify either --replicas or --autoscaling-enabled",
		},
		{
			name: "autoscaling with min zero",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Autoscaling: &AutoscalingInput{
					Enabled:       true,
					MinReplicas:   0,
					MaxReplicas:   10,
					CPUPercentage: 80,
				},
			},
			wantErr:         true,
			wantErrContains: "--autoscaling-min-replicas must be greater than zero",
		},
		{
			name: "autoscaling with max zero",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Autoscaling: &AutoscalingInput{
					Enabled:       true,
					MinReplicas:   2,
					MaxReplicas:   0,
					CPUPercentage: 80,
				},
			},
			wantErr:         true,
			wantErrContains: "--autoscaling-max-replicas must be greater than zero",
		},
		{
			name: "autoscaling with min greater than max",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Autoscaling: &AutoscalingInput{
					Enabled:       true,
					MinReplicas:   10,
					MaxReplicas:   5,
					CPUPercentage: 80,
				},
			},
			wantErr:         true,
			wantErrContains: "--autoscaling-min-replicas cannot be greater than --autoscaling-max-replicas",
		},
		{
			name: "autoscaling without any metrics",
			input: ServiceInput{
				Name:      "my-service",
				Port:      8080,
				ImageType: "internal",
				ImageName: "my-image",
				ImageTag:  "1.0.0",
				Memory:    "512M",
				CPU:       "1",
				Autoscaling: &AutoscalingInput{
					Enabled:       true,
					MinReplicas:   2,
					MaxReplicas:   10,
					CPUPercentage: 0,
					MemoryPercent: 0,
				},
			},
			wantErr:         true,
			wantErrContains: "at least one of --autoscaling-cpu-percentage or --autoscaling-memory-percentage must be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateService(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateService() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("ValidateService() error = %v, want error containing %q", err, tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateService() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateAutoscaling(t *testing.T) {
	tests := []struct {
		name            string
		config          AutoscalingInput
		wantErr         bool
		wantErrContains string
	}{
		{
			name: "valid with both metrics",
			config: AutoscalingInput{
				Enabled:       true,
				MinReplicas:   2,
				MaxReplicas:   10,
				CPUPercentage: 80,
				MemoryPercent: 70,
			},
			wantErr: false,
		},
		{
			name: "valid with CPU only",
			config: AutoscalingInput{
				Enabled:       true,
				MinReplicas:   1,
				MaxReplicas:   5,
				CPUPercentage: 75,
			},
			wantErr: false,
		},
		{
			name: "valid with Memory only",
			config: AutoscalingInput{
				Enabled:       true,
				MinReplicas:   1,
				MaxReplicas:   5,
				MemoryPercent: 80,
			},
			wantErr: false,
		},
		{
			name: "min replicas zero",
			config: AutoscalingInput{
				Enabled:       true,
				MinReplicas:   0,
				MaxReplicas:   10,
				CPUPercentage: 80,
			},
			wantErr:         true,
			wantErrContains: "--autoscaling-min-replicas must be greater than zero",
		},
		{
			name: "max replicas zero",
			config: AutoscalingInput{
				Enabled:       true,
				MinReplicas:   2,
				MaxReplicas:   0,
				CPUPercentage: 80,
			},
			wantErr:         true,
			wantErrContains: "--autoscaling-max-replicas must be greater than zero",
		},
		{
			name: "min greater than max",
			config: AutoscalingInput{
				Enabled:       true,
				MinReplicas:   10,
				MaxReplicas:   5,
				CPUPercentage: 80,
			},
			wantErr:         true,
			wantErrContains: "--autoscaling-min-replicas cannot be greater than --autoscaling-max-replicas",
		},
		{
			name: "no metrics set",
			config: AutoscalingInput{
				Enabled:       true,
				MinReplicas:   2,
				MaxReplicas:   10,
				CPUPercentage: 0,
				MemoryPercent: 0,
			},
			wantErr:         true,
			wantErrContains: "at least one of --autoscaling-cpu-percentage or --autoscaling-memory-percentage must be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAutoscaling(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAutoscaling() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("ValidateAutoscaling() error = %v, want error containing %q", err, tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAutoscaling() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateServiceEnvVars(t *testing.T) {
	tests := []struct {
		name            string
		envVars         []string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:    "valid env vars",
			envVars: []string{"KEY1=value1", "KEY2=value2", "KEY3=value with spaces"},
			wantErr: false,
		},
		{
			name:    "valid env var with multiple equals",
			envVars: []string{"KEY=value=with=equals"},
			wantErr: false,
		},
		{
			name:    "empty slice",
			envVars: []string{},
			wantErr: false,
		},
		{
			name:    "nil slice",
			envVars: nil,
			wantErr: false,
		},
		{
			name:            "missing equals",
			envVars:         []string{"INVALID"},
			wantErr:         true,
			wantErrContains: "invalid --env value",
		},
		{
			name:            "empty key",
			envVars:         []string{"=value"},
			wantErr:         true,
			wantErrContains: "invalid --env value",
		},
		{
			name:            "whitespace only key",
			envVars:         []string{"  =value"},
			wantErr:         true,
			wantErrContains: "invalid --env value",
		},
		{
			name:            "one valid one invalid",
			envVars:         []string{"KEY1=value1", "INVALID"},
			wantErr:         true,
			wantErrContains: "invalid --env value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServiceEnvVars(tt.envVars)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateServiceEnvVars() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("ValidateServiceEnvVars() error = %v, want error containing %q", err, tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServiceEnvVars() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateServiceSecretRefs(t *testing.T) {
	tests := []struct {
		name            string
		secretRefs      []string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:       "valid secret refs",
			secretRefs: []string{"secret1", "secret2", "my-secret-3"},
			wantErr:    false,
		},
		{
			name:       "empty slice",
			secretRefs: []string{},
			wantErr:    false,
		},
		{
			name:       "nil slice",
			secretRefs: nil,
			wantErr:    false,
		},
		{
			name:            "empty string",
			secretRefs:      []string{""},
			wantErr:         true,
			wantErrContains: "invalid --secret value",
		},
		{
			name:            "whitespace only",
			secretRefs:      []string{"   "},
			wantErr:         true,
			wantErrContains: "invalid --secret value",
		},
		{
			name:            "one valid one invalid",
			secretRefs:      []string{"valid-secret", "  "},
			wantErr:         true,
			wantErrContains: "invalid --secret value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServiceSecretRefs(tt.secretRefs)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateServiceSecretRefs() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("ValidateServiceSecretRefs() error = %v, want error containing %q", err, tt.wantErrContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServiceSecretRefs() unexpected error = %v", err)
				}
			}
		})
	}
}
