package output

import (
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

func formatEnvValue(env deployment.EnvVar) string {
	if env.ValueFrom != nil {
		ref := env.ValueFrom.SecretKeyRef
		return fmt.Sprintf("<secret: %s/%s>", ref.Name, ref.Key)
	}
	return env.Value
}
