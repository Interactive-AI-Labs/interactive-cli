package replay

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/inputs"
)

// APIKeyEnv is the environment variable that supplies the agent's bearer.
const APIKeyEnv = "INTERACTIVE_AGENT_API_KEY"

type SecretReader interface {
	GetSecret(ctx context.Context, orgID, projectID, name string) (*deployment.SecretInfo, error)
}

// ResolveBearer picks the agent's API key: the flag, then the env var, then
// the value the agent itself receives — from its env list or, walking its
// secret refs in order, the first secret that holds the referenced key.
func ResolveBearer(
	ctx context.Context,
	secrets SecretReader,
	orgID, projectID string,
	agent *deployment.DescribeAgentResponse,
	flagValue, envValue string,
	errW io.Writer,
) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if envValue != "" {
		return envValue, nil
	}

	name := inputs.RuntimeAPIKeyRef(agent.AgentConfig)
	if name == "" {
		return "", fmt.Errorf(
			"could not resolve the agent's api key: agentConfig.runtime.api_key is not a ${VAR} reference; "+
				"pass --agent-api-key or set %s",
			APIKeyEnv,
		)
	}

	for _, e := range agent.Env {
		if e.Name == name && e.Value != "" {
			fmt.Fprintf(errW, "using %s from the agent's env\n", name)
			return e.Value, nil
		}
	}

	for _, ref := range agent.SecretRefs {
		secret, err := secrets.GetSecret(ctx, orgID, projectID, ref.SecretName)
		if err != nil {
			return "", fmt.Errorf(
				"failed to read secret %q while resolving %s (pass --agent-api-key or set %s to skip): %w",
				ref.SecretName,
				name,
				APIKeyEnv,
				err,
			)
		}
		value, ok := secret.Data[name]
		if !ok || value == "" {
			continue
		}
		// Secret values arrive base64-encoded. Decode only when the result is
		// printable ASCII: a raw key that happens to be valid base64 would
		// otherwise be silently mangled into a wrong credential, and every real
		// API key is printable while random decoded bytes almost never are.
		if decoded, err := base64.StdEncoding.DecodeString(value); err == nil &&
			isPrintableASCII(decoded) {
			value = string(decoded)
		}
		fmt.Fprintf(errW, "using %s from secret %s\n", name, ref.SecretName)
		return value, nil
	}

	return "", fmt.Errorf(
		"could not resolve the agent's api key: %s not found in the agent's env or secrets; "+
			"pass --agent-api-key or set %s", name, APIKeyEnv,
	)
}

func isPrintableASCII(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}
