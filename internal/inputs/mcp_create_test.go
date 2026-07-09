package inputs

import "testing"

func TestBuildMcpRequestBodyInternalSecrets(t *testing.T) {
	body, err := BuildMcpRequestBody(McpInput{
		Port:       8080,
		ImageName:  "my-mcp",
		ImageTag:   "v1",
		SecretRefs: []string{"db-creds", "api-key"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(body.SecretRefs) != 2 || body.SecretRefs[0].SecretName != "db-creds" || body.SecretRefs[1].SecretName != "api-key" {
		t.Errorf("SecretRefs = %+v", body.SecretRefs)
	}
}

func TestBuildMcpRequestBodyInternalSecretsRejectsEmptyName(t *testing.T) {
	_, err := BuildMcpRequestBody(McpInput{
		Port:       8080,
		ImageName:  "my-mcp",
		ImageTag:   "v1",
		SecretRefs: []string{"  "},
	})
	if err == nil {
		t.Fatal("empty secret name must be rejected")
	}
}

func TestBuildMcpRequestBodyExternalRejectsSecret(t *testing.T) {
	_, err := BuildMcpRequestBody(McpInput{
		EndpointURL: "https://mcp.acme.com/mcp",
		SecretRefs:  []string{"some-secret"},
	})
	if err == nil {
		t.Fatal("--secret with an external mcp must be rejected, not silently dropped")
	}
}
