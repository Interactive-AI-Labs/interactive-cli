package inputs

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// McpInput holds the flags for `iai mcps create`/`update`. Exactly one
// source must be set: --catalog-id, --endpoint-url (both external), or
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

	// external
	EndpointURL string
	CatalogID   string

	Credential string
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

	mcpType := strings.TrimSpace(in.Type)
	switch {
	case mcpType != "":
		// explicit
	case in.CatalogID != "" || in.EndpointURL != "":
		mcpType = "external"
	default:
		mcpType = "internal"
	}

	body := clients.CreateMcpBody{
		Type:       mcpType,
		CatalogID:  strings.TrimSpace(in.CatalogID),
		Credential: in.Credential,
	}

	switch mcpType {
	case "external":
		if in.CatalogID != "" && in.EndpointURL != "" {
			return clients.CreateMcpBody{}, fmt.Errorf(
				"--catalog-id and --endpoint-url are mutually exclusive",
			)
		}
		if in.CatalogID == "" && in.EndpointURL == "" {
			return clients.CreateMcpBody{}, fmt.Errorf(
				"external mcps need --catalog-id or --endpoint-url",
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
		if in.Memory != "" || in.CPU != "" {
			body.Resources = clients.Resources{Memory: in.Memory, CPU: in.CPU}
		}
	default:
		return clients.CreateMcpBody{}, fmt.Errorf(`--type must be "internal" or "external"`)
	}

	return body, nil
}
