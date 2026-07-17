package inputs

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// McpInput holds the flags for `iai mcps create`/`update`; the server does the real validation.
type McpInput struct {
	Type string // "internal" | "external" — inferred if empty

	// internal
	Port            int
	Path            string // default "/mcp"
	ImageType       string
	ImageRepository string
	ImageName       string
	ImageTag        string
	Memory          string
	CPU             string
	EnvVars         []string
	SecretRefs      []string // existing secrets loaded as env vars

	// external
	EndpointURL string
	CatalogID   string

	AuthType         string // bearer | api_key | custom | none — inferred when empty, see BuildMcpRequestBody
	Credential       string
	AuthHeader       string
	AuthHeaderPrefix string
	Headers          []string // raw KEY=VALUE pairs
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

	var headers map[string]string
	if len(in.Headers) > 0 {
		headers = make(map[string]string, len(in.Headers))
		for _, p := range in.Headers {
			key, value, found := strings.Cut(p, "=")
			key = strings.TrimSpace(key)
			if !found || key == "" {
				return clients.CreateMcpBody{}, fmt.Errorf(
					"invalid --header %q: expected KEY=VALUE",
					p,
				)
			}
			headers[key] = value
		}
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

	// a header override implies custom, otherwise a credential implies bearer, else none
	authType := strings.TrimSpace(in.AuthType)
	if authType == "" {
		switch {
		case strings.TrimSpace(in.AuthHeader) != "" || strings.TrimSpace(in.AuthHeaderPrefix) != "":
			authType = "custom"
		case strings.TrimSpace(in.Credential) != "":
			authType = "bearer"
		default:
			authType = "none"
		}
	}

	body := clients.CreateMcpBody{
		Type:      mcpType,
		CatalogID: strings.TrimSpace(in.CatalogID),
		Auth: clients.McpAuthBody{
			Type:         authType,
			Credential:   in.Credential,
			Header:       strings.TrimSpace(in.AuthHeader),
			HeaderPrefix: in.AuthHeaderPrefix,
		},
		Headers: headers,
	}

	switch mcpType {
	case "external":
		if in.CatalogID == "" && in.EndpointURL == "" {
			return clients.CreateMcpBody{}, fmt.Errorf(
				"external mcps need --catalog-id or --external-url",
			)
		}
		if len(env) > 0 || len(secretRefs) > 0 || in.Path != "" {
			return clients.CreateMcpBody{}, fmt.Errorf(
				"--env, --secret, and --path don't apply to an external mcp — the path is part of --external-url",
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
		body.Path = strings.TrimSpace(in.Path)
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
