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

func TestBuildMcpRequestBodyAuth(t *testing.T) {
	tests := []struct {
		name         string
		input        McpInput
		wantAuthType string
		wantHeaders  map[string]string
		wantErr      string
	}{
		{
			name:         "credential infers bearer",
			input:        McpInput{EndpointURL: "https://x.io/mcp", Credential: "tok"},
			wantAuthType: "bearer",
		},
		{
			name:         "no credential infers none",
			input:        McpInput{EndpointURL: "https://x.io/mcp"},
			wantAuthType: "none",
		},
		{
			name: "explicit auth-type wins over inference",
			input: McpInput{
				EndpointURL: "https://x.io/mcp",
				Credential:  "k",
				AuthType:    "api_key",
			},
			wantAuthType: "api_key",
		},
		{
			name: "headers parsed into a map",
			input: McpInput{
				EndpointURL: "https://x.io/mcp",
				Headers:     []string{"X-Org=acme", "X-Trace=1"},
			},
			wantAuthType: "none",
			wantHeaders:  map[string]string{"X-Org": "acme", "X-Trace": "1"},
		},
		{
			name:    "malformed header rejected",
			input:   McpInput{EndpointURL: "https://x.io/mcp", Headers: []string{"bogus"}},
			wantErr: "expected KEY=VALUE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := BuildMcpRequestBody(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if body.AuthType != tt.wantAuthType {
				t.Errorf("AuthType = %q, want %q", body.AuthType, tt.wantAuthType)
			}
			if len(body.Headers) != len(tt.wantHeaders) {
				t.Fatalf("Headers = %v, want %v", body.Headers, tt.wantHeaders)
			}
			for k, want := range tt.wantHeaders {
				if body.Headers[k] != want {
					t.Errorf("Headers[%q] = %q, want %q", k, body.Headers[k], want)
				}
			}
		})
	}
}
