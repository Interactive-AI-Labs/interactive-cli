package inputs

import (
	"strings"
	"testing"
)

func TestBuildMcpRequestBody(t *testing.T) {
	tests := []struct {
		name       string
		input      McpInput
		wantErr    bool
		errContain string
		wantSecret []string
	}{
		{
			name: "internal with secrets",
			input: McpInput{
				Port:       8080,
				ImageName:  "my-mcp",
				ImageTag:   "v1",
				SecretRefs: []string{"db-creds", "api-key"},
			},
			wantSecret: []string{"db-creds", "api-key"},
		},
		{
			name: "internal rejects empty secret name",
			input: McpInput{
				Port:       8080,
				ImageName:  "my-mcp",
				ImageTag:   "v1",
				SecretRefs: []string{"  "},
			},
			wantErr: true,
		},
		{
			name: "external rejects secret instead of silently dropping it",
			input: McpInput{
				EndpointURL: "https://mcp.acme.com/mcp",
				SecretRefs:  []string{"some-secret"},
			},
			wantErr:    true,
			errContain: "--env and --secret don't apply",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := BuildMcpRequestBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Fatalf("err = %q, want containing %q", err, tt.errContain)
				}
				return
			}
			if len(body.SecretRefs) != len(tt.wantSecret) {
				t.Fatalf("SecretRefs = %+v, want %v", body.SecretRefs, tt.wantSecret)
			}
			for i, want := range tt.wantSecret {
				if body.SecretRefs[i].SecretName != want {
					t.Errorf("SecretRefs[%d] = %q, want %q", i, body.SecretRefs[i].SecretName, want)
				}
			}
		})
	}
}
