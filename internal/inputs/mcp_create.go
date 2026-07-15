package inputs

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// McpInput holds the flags for `iai mcps create`/`update`. Exactly one
// source must be set: --catalog-id, --external-url (both external), or
// --image-name/--image-tag/--port (internal). The server does the real
// validation; this builder only assembles the request body.
type McpInput struct {
	Type string // "internal" | "external" — inferred if empty

	// internal
	Port            int
	ImageType       string
	ImageRepository string
	ImageName       string
	ImageTag        string
	Memory          string
	CPU             string
	EnvVars         []string
	SecretRefs      []string // existing k8s Secrets loaded as env vars

	// external
	EndpointURL string
	CatalogID   string

	// AuthType is bearer | api_key | none. Inferred when empty: bearer if a
	// credential is given, otherwise none.
	AuthType         string
	Credential       string
	AuthHeader       string
	AuthHeaderPrefix string
	Headers          []string // raw KEY=VALUE pairs
}

// parseHeaderFlags turns repeated --header KEY=VALUE pairs into a map.
func parseHeaderFlags(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	headers := make(map[string]string, len(pairs))
	for _, p := range pairs {
		key, value, found := strings.Cut(p, "=")
		key = strings.TrimSpace(key)
		if !found || key == "" {
			return nil, fmt.Errorf("invalid --header %q: expected KEY=VALUE", p)
		}
		headers[key] = value
	}
	return headers, nil
}

func BuildMcpRequestBody(in McpInput) (clients.CreateMcpBody, error) {
	if err := ValidateServiceEnvVars(in.EnvVars); err != nil {
		return clients.CreateMcpBody{}, err
	}
	var env []clients.EnvVar
	for _, e := range in.EnvVars {
		parts := strings.SplitN(e, "=", 2)
		env = append(env, clients.EnvVar{
			Name:  strings.TrimSpace(parts[0]),
			Value: parts[1],
		})
	}

	if err := ValidateServiceSecretRefs(in.SecretRefs); err != nil {
		return clients.CreateMcpBody{}, err
	}
	var secretRefs []clients.SecretRef
	for _, name := range in.SecretRefs {
		secretRefs = append(secretRefs, clients.SecretRef{SecretName: strings.TrimSpace(name)})
	}

	headers, err := parseHeaderFlags(in.Headers)
	if err != nil {
		return clients.CreateMcpBody{}, err
	}

	mcpType := strings.TrimSpace(in.Type)
	switch {
	case mcpType != "":
		// explicit
	case in.CatalogID != "" || in.EndpointURL != "":
		mcpType = "external"
	default:
		mcpType = "internal"
	}

	// authType is required by the server; infer the common cases so existing
	// invocations keep working — a credential means bearer, none means no auth.
	authType := strings.TrimSpace(in.AuthType)
	if authType == "" {
		if strings.TrimSpace(in.Credential) != "" {
			authType = "bearer"
		} else {
			authType = "none"
		}
	}

	body := clients.CreateMcpBody{
		Type:             mcpType,
		CatalogID:        strings.TrimSpace(in.CatalogID),
		AuthType:         authType,
		Credential:       in.Credential,
		AuthHeader:       strings.TrimSpace(in.AuthHeader),
		AuthHeaderPrefix: in.AuthHeaderPrefix,
		Headers:          headers,
	}

	switch mcpType {
	case "external":
		if in.CatalogID == "" && in.EndpointURL == "" {
			return clients.CreateMcpBody{}, fmt.Errorf(
				"external mcps need --catalog-id or --external-url",
			)
		}
		if len(env) > 0 || len(secretRefs) > 0 {
			return clients.CreateMcpBody{}, fmt.Errorf(
				"--env and --secret don't apply to an external mcp",
			)
		}
		body.EndpointURL = strings.TrimSpace(in.EndpointURL)
	case "internal":
		if in.Port <= 0 {
			return clients.CreateMcpBody{}, fmt.Errorf("--port is required for an internal mcp")
		}
		if strings.TrimSpace(in.ImageName) == "" || strings.TrimSpace(in.ImageTag) == "" {
			return clients.CreateMcpBody{}, fmt.Errorf(
				"--image-name and --image-tag are required for an internal mcp",
			)
		}
		body.Port = in.Port
		body.Image = clients.ImageSpec{
			Type:       in.ImageType,
			Repository: in.ImageRepository,
			Name:       in.ImageName,
			Tag:        in.ImageTag,
		}
		body.Env = env
		body.SecretRefs = secretRefs
		if in.Memory != "" || in.CPU != "" {
			body.Resources = clients.Resources{Memory: in.Memory, CPU: in.CPU}
		}
	default:
		return clients.CreateMcpBody{}, fmt.Errorf(`--type must be "internal" or "external"`)
	}

	return body, nil
}
