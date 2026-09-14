package inputs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

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
