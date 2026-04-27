package inputs

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
)

// Helpers for building partial-update patches. Shared by services.go and
// agents.go.

func setJSON(patch map[string]json.RawMessage, key string, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to encode %s: %w", key, err)
	}
	patch[key] = raw
	return nil
}

func anyChanged(changed func(string) bool, names ...string) bool {
	return slices.ContainsFunc(names, changed)
}

// setEnvPatch handles the env field: emits null on clear, replaces the full
// list on --env, leaves it alone otherwise. Rejects clear combined with --env.
func setEnvPatch(patch clients.UpdatePatch, envVars []string, changed, clear bool) error {
	if clear && changed {
		return fmt.Errorf("--clear-env cannot be combined with --env")
	}
	if clear {
		patch["env"] = json.RawMessage("null")
		return nil
	}
	if !changed {
		return nil
	}
	if len(envVars) == 0 {
		return fmt.Errorf(
			"--env requires at least one NAME=VALUE argument; use --clear-env to remove all variables",
		)
	}
	if err := ValidateServiceEnvVars(envVars); err != nil {
		return err
	}
	env := []clients.EnvVar{}
	for _, e := range envVars {
		parts := strings.SplitN(e, "=", 2)
		env = append(env, clients.EnvVar{
			Name: strings.TrimSpace(parts[0]), Value: parts[1],
		})
	}
	return setJSON(patch, "env", env)
}

// setSecretRefsPatch handles the secretRefs field: emits null on clear,
// replaces the full list on --secret, leaves it alone otherwise. Rejects clear
// combined with --secret.
func setSecretRefsPatch(patch clients.UpdatePatch, refs []string, changed, clear bool) error {
	if clear && changed {
		return fmt.Errorf("--clear-secret cannot be combined with --secret")
	}
	if clear {
		patch["secretRefs"] = json.RawMessage("null")
		return nil
	}
	if !changed {
		return nil
	}
	if len(refs) == 0 {
		return fmt.Errorf(
			"--secret requires at least one secret name; use --clear-secret to remove all secret references",
		)
	}
	if err := ValidateServiceSecretRefs(refs); err != nil {
		return err
	}
	out := []clients.SecretRef{}
	for _, name := range refs {
		out = append(out, clients.SecretRef{SecretName: strings.TrimSpace(name)})
	}
	return setJSON(patch, "secretRefs", out)
}

// setEndpointPatch adds the endpoint flag to the patch when the user explicitly
// passed --endpoint (true or false).
func setEndpointPatch(patch clients.UpdatePatch, endpoint bool, changed bool) error {
	if !changed {
		return nil
	}
	return setJSON(patch, "endpoint", endpoint)
}

// ScheduleInput collects the schedule-related flags both update commands accept.
type ScheduleInput struct {
	Uptime          string
	Downtime        string
	Timezone        string
	UptimeChanged   bool
	DowntimeChanged bool
	TimezoneChanged bool
	Clear           bool
}

// setSchedulePatch handles the schedule field for both services and agents:
// emits null on Clear, builds a partial object from the changed sub-flags
// otherwise. Rejects clear combined with any --schedule-* setter.
func setSchedulePatch(patch clients.UpdatePatch, in ScheduleInput) error {
	anySet := in.UptimeChanged || in.DowntimeChanged || in.TimezoneChanged
	if in.Clear && anySet {
		return fmt.Errorf("--clear-schedule cannot be combined with --schedule-* flags")
	}
	if in.Clear {
		patch["schedule"] = json.RawMessage("null")
		return nil
	}
	if !anySet {
		return nil
	}
	sched := map[string]any{}
	if in.UptimeChanged {
		sched["uptime"] = in.Uptime
	}
	if in.DowntimeChanged {
		sched["downtime"] = in.Downtime
	}
	if in.TimezoneChanged {
		sched["timezone"] = in.Timezone
	}
	return setJSON(patch, "schedule", sched)
}
