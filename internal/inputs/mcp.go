package inputs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
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
