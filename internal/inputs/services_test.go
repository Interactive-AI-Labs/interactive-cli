package inputs

import (
	"strings"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
	"github.com/google/go-cmp/cmp"
)

func TestBuildServiceRequestBodyScaling(t *testing.T) {
	tests := []struct {
		name            string
		in              ServiceInput
		wantReplicas    int
		wantAutoscaling *clients.Autoscaling
	}{
		{
			name:         "replicas only",
			in:           ServiceInput{Replicas: 3},
			wantReplicas: 3,
		},
		{
			name: "autoscaling flags only",
			in: ServiceInput{
				AutoscalingMin:    2,
				AutoscalingMax:    10,
				AutoscalingCPU:    80,
				AutoscalingMemory: 85,
			},
			wantAutoscaling: &clients.Autoscaling{
				MinReplicas:      2,
				MaxReplicas:      10,
				CPUPercentage:    utils.ToPtr(80),
				MemoryPercentage: utils.ToPtr(85),
			},
		},
		{
			name: "single autoscaling flag triggers block",
			in:   ServiceInput{AutoscalingMin: 2},
			wantAutoscaling: &clients.Autoscaling{
				MinReplicas: 2,
			},
		},
		{
			name: "both set are passed through for backend to reject",
			in: ServiceInput{
				Replicas:       5,
				AutoscalingMin: 2,
				AutoscalingMax: 10,
			},
			wantReplicas: 5,
			wantAutoscaling: &clients.Autoscaling{
				MinReplicas: 2,
				MaxReplicas: 10,
			},
		},
		{
			name: "neither set",
			in:   ServiceInput{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildServiceRequestBody(tt.in)
			if err != nil {
				t.Fatalf("BuildServiceRequestBody() unexpected error = %v", err)
			}
			if got.Replicas != tt.wantReplicas {
				t.Errorf("Replicas = %d, want %d", got.Replicas, tt.wantReplicas)
			}
			if diff := cmp.Diff(tt.wantAutoscaling, got.Autoscaling); diff != "" {
				t.Errorf("Autoscaling mismatch (-want +got):\n%s", diff)
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
					t.Errorf(
						"ValidateServiceEnvVars() error = %v, want error containing %q",
						err,
						tt.wantErrContains,
					)
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
					t.Errorf(
						"ValidateServiceSecretRefs() error = %v, want error containing %q",
						err,
						tt.wantErrContains,
					)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServiceSecretRefs() unexpected error = %v", err)
				}
			}
		})
	}
}
