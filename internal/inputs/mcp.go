package inputs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// ResolveCredential reads from stdin when fromStdin is set, keeping the secret out of the process list and shell history.
func ResolveCredential(in io.Reader, credential string, fromStdin bool) (string, error) {
	if !fromStdin {
		return credential, nil
	}
	data, err := io.ReadAll(in)
	if err != nil {
		return "", fmt.Errorf("failed to read credential from stdin: %w", err)
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}

func ResolveToolArgs(inline, file string) (map[string]any, error) {
	raw := inline
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read --args-file %q: %w", file, err)
		}
		raw = string(data)
	}
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return nil, fmt.Errorf("invalid tool arguments: must be a JSON object: %w", err)
	}
	// null unmarshals to a nil map without error, so reject it like any non-object.
	if args == nil {
		return nil, fmt.Errorf("invalid tool arguments: must be a JSON object, got null")
	}
	return args, nil
}

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

// McpUpdateFlags is the set of cobra flag names BuildMcpUpdatePatch inspects via the `changed` predicate. Keep in sync with cmd/mcps.go.
var McpUpdateFlags = struct {
	Port             string
	Path             string
	ImageType        string
	ImageRepository  string
	ImageName        string
	ImageTag         string
	Memory           string
	CPU              string
	Env              string
	Secret           string
	AuthType         string
	Credential       string
	AuthHeader       string
	AuthHeaderPrefix string
	Header           string
}{
	Port:             "port",
	Path:             "path",
	ImageType:        "image-type",
	ImageRepository:  "image-repository",
	ImageName:        "image-name",
	ImageTag:         "image-tag",
	Memory:           "memory",
	CPU:              "cpu",
	Env:              "env",
	Secret:           "secret",
	AuthType:         "auth-type",
	Credential:       "credential",
	AuthHeader:       "auth-header",
	AuthHeaderPrefix: "auth-header-prefix",
	Header:           "header",
}

// BuildMcpUpdatePatch produces a partial-update body containing only the fields whose flags the user explicitly set.
func BuildMcpUpdatePatch(
	in McpInput,
	clearEnv, clearSecret, clearHeaders bool,
	changed func(string) bool,
) (clients.UpdatePatch, error) {
	f := McpUpdateFlags
	patch := clients.UpdatePatch{}

	if changed(f.Port) {
		if err := setJSON(patch, "port", in.Port); err != nil {
			return nil, err
		}
	}
	if changed(f.Path) {
		if err := setJSON(patch, "path", strings.TrimSpace(in.Path)); err != nil {
			return nil, err
		}
	}

	if anyChanged(changed, f.ImageType, f.ImageRepository, f.ImageName, f.ImageTag) {
		img := map[string]any{}
		if changed(f.ImageType) {
			img["type"] = in.ImageType
		}
		if changed(f.ImageRepository) {
			img["repository"] = in.ImageRepository
		}
		if changed(f.ImageName) {
			img["name"] = in.ImageName
		}
		if changed(f.ImageTag) {
			img["tag"] = in.ImageTag
		}
		if err := setJSON(patch, "image", img); err != nil {
			return nil, err
		}
	}

	if anyChanged(changed, f.Memory, f.CPU) {
		res := map[string]any{}
		if changed(f.Memory) {
			res["memory"] = in.Memory
		}
		if changed(f.CPU) {
			res["cpu"] = in.CPU
		}
		if err := setJSON(patch, "resources", res); err != nil {
			return nil, err
		}
	}

	if err := setEnvPatch(patch, in.EnvVars, changed(f.Env), clearEnv); err != nil {
		return nil, err
	}
	if err := setSecretRefsPatch(patch, in.SecretRefs, changed(f.Secret), clearSecret); err != nil {
		return nil, err
	}

	if anyChanged(changed, f.AuthType, f.Credential, f.AuthHeader, f.AuthHeaderPrefix) {
		newType := strings.TrimSpace(in.AuthType)
		auth := map[string]any{}
		if changed(f.AuthType) {
			auth["type"] = newType
		}
		switch {
		case changed(f.Credential):
			auth["credential"] = in.Credential
		case changed(f.AuthType) && newType == "none":
			// switching to none implicitly drops any credential the mcp already has
			auth["credential"] = ""
		}
		if changed(f.AuthHeader) {
			auth["header"] = strings.TrimSpace(in.AuthHeader)
		}
		if changed(f.AuthHeaderPrefix) {
			auth["headerPrefix"] = in.AuthHeaderPrefix
		}
		if err := setJSON(patch, "auth", auth); err != nil {
			return nil, err
		}
	}

	switch {
	case clearHeaders && changed(f.Header):
		return nil, fmt.Errorf("--clear-headers cannot be combined with --header")
	case clearHeaders:
		patch["headers"] = json.RawMessage("null")
	case changed(f.Header):
		headers, err := parseHeaderFlags(in.Headers)
		if err != nil {
			return nil, err
		}
		if err := setJSON(patch, "headers", headers); err != nil {
			return nil, err
		}
	}

	return patch, nil
}
