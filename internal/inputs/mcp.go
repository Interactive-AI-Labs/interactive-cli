package inputs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

var (
	validMcpAuthTypes  = []string{"api_key", "bearer", "none"}
	validMcpTransports = []string{"streamable_http", "sse"}
)

func ValidateMcpAuth(authType, credential string) error {
	if !slices.Contains(validMcpAuthTypes, authType) {
		return fmt.Errorf(
			"invalid --auth-type %q: must be one of %s",
			authType,
			strings.Join(validMcpAuthTypes, ", "),
		)
	}
	if authType == "none" && credential != "" {
		return fmt.Errorf("--credential must not be set when --auth-type is 'none'")
	}
	if authType != "none" && credential == "" {
		return fmt.Errorf("--credential is required when --auth-type is %q", authType)
	}
	return nil
}

func ValidateMcpTransport(transport string) error {
	if !slices.Contains(validMcpTransports, transport) {
		return fmt.Errorf(
			"invalid --transport %q: must be one of %s",
			transport,
			strings.Join(validMcpTransports, ", "),
		)
	}
	return nil
}

func ParseHeaderFlags(pairs []string) (map[string]string, error) {
	headers := make(map[string]string, len(pairs))
	for _, p := range pairs {
		key, value, found := strings.Cut(p, "=")
		if !found || key == "" {
			return nil, fmt.Errorf("invalid --header %q: expected KEY=VALUE", p)
		}
		headers[key] = value
	}
	return headers, nil
}

// ResolveCredential reads the credential from stdin when --credential-stdin is
// set, which keeps the secret out of the process list and shell history.
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

// CatalogEndpointURL returns the canonical endpoint of a catalog entry, which
// the backend requires when creating a connection from the catalog.
func CatalogEndpointURL(entries []clients.McpCatalogEntry, catalogID string) (string, error) {
	for _, e := range entries {
		if e.ID == catalogID {
			if strings.TrimSpace(e.EndpointURL) == "" {
				return "", fmt.Errorf(
					"catalog entry %q has no managed endpoint; create a custom connector with --endpoint-url instead",
					catalogID,
				)
			}
			return e.EndpointURL, nil
		}
	}
	return "", fmt.Errorf("catalog entry %q not found; see 'iai connectors catalog'", catalogID)
}
