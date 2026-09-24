package deployment

import "strings"

// UserEnv drops the MCP_KEY_* variables the operator reserves for MCP credentials and rejects when sent back.
func UserEnv(env []EnvVar) []EnvVar {
	var userEnv []EnvVar
	for _, e := range env {
		if !strings.HasPrefix(e.Name, "MCP_KEY_") {
			userEnv = append(userEnv, e)
		}
	}
	return userEnv
}
