package inputs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/platform"
)

// ResolveCredential reads stdin when requested to keep the value out of shell history.
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

type McpUpdateInput struct {
	ImageName, ImageTag                 string
	Port                                int
	Endpoint                            bool
	Path, Memory, CPU, StackId          string
	SecretRefs                          []string
	Auth                                platform.McpAuth
	Description                         string
	EnvVars                             []string
	ClearEnv, ClearSecret, ClearStackID bool
}

// BuildMcpUpdatePatch keeps omitted fields distinct from explicit empty values.
func BuildMcpUpdatePatch(
	in McpUpdateInput,
	changed func(string) bool,
) (platform.McpUpdateRequest, error) {
	patch := platform.McpUpdateRequest{}
	if changed("description") {
		patch["description"] = in.Description
	}

	auth := map[string]any{}
	for _, field := range []struct {
		flag, key string
		value     any
	}{
		{"auth-type", "type", in.Auth.Type},
		{"credential", "credential", in.Auth.Credential},
		{"auth-header", "header_name", in.Auth.HeaderName},
		{"auth-header-prefix", "header_prefix", in.Auth.HeaderPrefix},
	} {
		if changed(field.flag) || (field.flag == "credential" && in.Auth.Credential != nil) {
			auth[field.key] = field.value
		}
	}
	if len(auth) > 0 {
		patch["auth"] = auth
	}

	workload := map[string]any{}
	if anyChanged(changed, "image-name", "image-tag") {
		img := map[string]any{}
		if changed("image-name") {
			img["name"] = in.ImageName
		}
		if changed("image-tag") {
			img["tag"] = in.ImageTag
		}
		workload["image"] = img
	}
	for _, field := range []struct {
		flag, key string
		value     any
	}{
		{"port", "port", in.Port},
		{"path", "path", in.Path},
		{"memory", "memory", in.Memory},
		{"cpu", "cpu", in.CPU},
		{"endpoint", "endpoint", in.Endpoint},
	} {
		if changed(field.flag) {
			workload[field.key] = field.value
		}
	}

	stack := deployment.UpdatePatch{}
	if err := setStackIdPatch(
		stack,
		in.StackId,
		changed("stack-id"),
		in.ClearStackID,
	); err != nil {
		return nil, err
	}
	if value, ok := stack["stackId"]; ok {
		workload["stack_id"] = value
	}

	env, err := ResolveMcpEnvVars(in.EnvVars, changed("env"), in.ClearEnv)
	if err != nil {
		return nil, err
	}
	if env != nil {
		workload["env"] = env
	}

	refs, err := ResolveMcpSecretRefs(in.SecretRefs, changed("secret"), in.ClearSecret)
	if err != nil {
		return nil, err
	}
	if refs != nil {
		workload["secret_refs"] = refs
	}

	if len(workload) > 0 {
		patch["workload"] = workload
	}

	return patch, nil
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
	if args == nil {
		return nil, fmt.Errorf("invalid tool arguments: must be a JSON object, got null")
	}
	return args, nil
}

// ResolveMcpEnvVars turns repeated --env NAME=VALUE flags into an mcp workload's
// env list. Nil leaves the deployed list alone; an empty list clears it.
func ResolveMcpEnvVars(envVars []string, changed, clear bool) ([]platform.McpEnvVar, error) {
	if clear && changed {
		return nil, fmt.Errorf("--clear-env cannot be combined with --env")
	}
	if clear {
		return []platform.McpEnvVar{}, nil
	}
	if !changed {
		return nil, nil
	}
	if len(envVars) == 0 {
		return nil, fmt.Errorf(
			"--env requires at least one NAME=VALUE argument; use --clear-env to remove all variables",
		)
	}
	if err := ValidateServiceEnvVars(envVars); err != nil {
		return nil, err
	}
	env := make([]platform.McpEnvVar, 0, len(envVars))
	for _, e := range envVars {
		name, value, _ := strings.Cut(e, "=")
		env = append(env, platform.McpEnvVar{Name: strings.TrimSpace(name), Value: value})
	}
	return env, nil
}

// ResolveMcpSecretRefs is ResolveMcpEnvVars for --secret/--clear-secret, over secret names.
func ResolveMcpSecretRefs(refs []string, changed, clear bool) ([]string, error) {
	if clear && changed {
		return nil, fmt.Errorf("--clear-secret cannot be combined with --secret")
	}
	if clear {
		return []string{}, nil
	}
	if !changed {
		return nil, nil
	}
	if len(refs) == 0 {
		return nil, fmt.Errorf(
			"--secret requires at least one secret name; use --clear-secret to remove all secret references",
		)
	}
	if err := ValidateServiceSecretRefs(refs); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(refs))
	for _, name := range refs {
		names = append(names, strings.TrimSpace(name))
	}
	return names, nil
}
