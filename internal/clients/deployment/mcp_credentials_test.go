package deployment

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUserEnv(t *testing.T) {
	credential := &EnvVarValueFrom{
		SecretKeyRef: SecretKeyRef{Name: "tools-dev", Key: "MCP_API_KEY"},
	}
	tests := []struct {
		name string
		env  []EnvVar
		want []EnvVar
	}{
		{
			name: "drops the managed MCP credential reference",
			env: []EnvVar{
				{Name: "LOG_LEVEL", Value: "debug"},
				{Name: "MCP_KEY_TOOLS_DEV", ValueFrom: credential},
				{Name: "DB_HOST", Value: "db"},
			},
			want: []EnvVar{
				{Name: "LOG_LEVEL", Value: "debug"},
				{Name: "DB_HOST", Value: "db"},
			},
		},
		{
			name: "drops a leftover MCP_KEY_ literal too, since the whole prefix is reserved",
			env: []EnvVar{
				{Name: "MCP_KEY_TOOLS_DEV", Value: ""},
				{Name: "LOG_LEVEL", Value: "debug"},
			},
			want: []EnvVar{{Name: "LOG_LEVEL", Value: "debug"}},
		},
		{
			name: "keeps a secret reference under another name",
			env:  []EnvVar{{Name: "CUSTOM_TOKEN", ValueFrom: credential}},
			want: []EnvVar{{Name: "CUSTOM_TOKEN", ValueFrom: credential}},
		},
		{
			name: "only managed references leaves nothing",
			env:  []EnvVar{{Name: "MCP_KEY_TOOLS_DEV", ValueFrom: credential}},
		},
		{
			name: "no env stays empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.want, UserEnv(tt.env)); diff != "" {
				t.Errorf("UserEnv() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
