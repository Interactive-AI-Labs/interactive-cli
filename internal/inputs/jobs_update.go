package inputs

import (
	"encoding/json"
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

type JobUpdateInput struct {
	JobInput
	ClearEnv, ClearSecret, ClearStackId, ClearSchedule, ClearRetention bool
	ClearCommand, ClearArgs                                            bool
}

func BuildJobUpdatePatch(
	in JobUpdateInput,
	changed func(string) bool,
) (deployment.UpdatePatch, error) {
	patch := deployment.UpdatePatch{}
	if changed("type") {
		if in.Type != "image" && in.Type != "script" {
			return nil, fmt.Errorf("--type must be image or script")
		}
		if err := setJSON(patch, "type", in.Type); err != nil {
			return nil, err
		}
		if in.Type == "script" {
			if anyChanged(
				changed,
				"image-type",
				"image-repository",
				"image-name",
				"image-tag",
				"command",
				"args",
			) {
				return nil, fmt.Errorf(
					"image, command and args flags are only available for image jobs",
				)
			}
			for _, field := range []string{"image", "command", "args"} {
				patch[field] = json.RawMessage("null")
			}
		} else {
			if anyChanged(changed, "script", "pyproject") {
				return nil, fmt.Errorf(
					"--script and --pyproject are only available for script jobs",
				)
			}
			patch["script"], patch["pyproject"] = json.RawMessage("null"), json.RawMessage("null")
		}
	}

	image := map[string]string{}
	for _, field := range []struct{ flag, key, value string }{
		{"image-type", "type", in.ImageType},
		{"image-repository", "repository", in.ImageRepository},
		{"image-name", "name", in.ImageName},
		{"image-tag", "tag", in.ImageTag},
	} {
		if changed(field.flag) {
			image[field.key] = field.value
		}
	}
	if len(image) > 0 {
		if err := setJSON(patch, "image", image); err != nil {
			return nil, err
		}
	}
	resources := map[string]string{}
	if changed("cpu") {
		resources["cpu"] = in.CPU
	}
	if changed("memory") {
		resources["memory"] = in.Memory
	}
	if len(resources) > 0 {
		if err := setJSON(patch, "resources", resources); err != nil {
			return nil, err
		}
	}
	for _, field := range []struct {
		flag  string
		value []string
		clear bool
	}{
		{"command", in.Command, in.ClearCommand}, {"args", in.Args, in.ClearArgs},
	} {
		if field.clear && changed(field.flag) {
			return nil, fmt.Errorf(
				"--clear-%s cannot be combined with --%s",
				field.flag,
				field.flag,
			)
		}
		if field.clear {
			patch[field.flag] = json.RawMessage("null")
		} else if changed(field.flag) {
			if err := setJSON(patch, field.flag, field.value); err != nil {
				return nil, err
			}
		}
	}
	for _, field := range []struct{ name, value string }{{"script", in.Script}, {"pyproject", in.Pyproject}} {
		if changed(field.name) {
			if err := setJSON(patch, field.name, field.value); err != nil {
				return nil, err
			}
		}
	}
	if err := setEnvPatch(patch, in.EnvVars, changed("env"), in.ClearEnv); err != nil {
		return nil, err
	}
	if err := setSecretRefsPatch(
		patch,
		in.SecretRefs,
		changed("secret"),
		in.ClearSecret,
	); err != nil {
		return nil, err
	}
	if err := setStackIdPatch(patch, in.StackId, changed("stack-id"), in.ClearStackId); err != nil {
		return nil, err
	}
	if in.ClearSchedule && anyChanged(changed, "schedule", "schedule-timezone") {
		return nil, fmt.Errorf(
			"--clear-schedule cannot be combined with --schedule or --schedule-timezone",
		)
	}
	if in.ClearSchedule {
		patch["cron"] = json.RawMessage("null")
	} else if anyChanged(changed, "schedule", "schedule-timezone") {
		cron := map[string]string{}
		if changed("schedule") {
			cron["schedule"] = in.Schedule
		}
		if changed("schedule-timezone") {
			cron["timezone"] = in.Timezone
		}
		if err := setJSON(patch, "cron", cron); err != nil {
			return nil, err
		}
	}
	for _, field := range []struct {
		name  string
		value any
	}{{"timeout", in.Timeout}, {"retries", in.Retries}} {
		if changed(field.name) {
			if err := setJSON(patch, field.name, field.value); err != nil {
				return nil, err
			}
		}
	}
	if in.ClearRetention &&
		anyChanged(changed, "retention-ttl", "retention-successful-runs", "retention-failed-runs") {
		return nil, fmt.Errorf("--clear-retention cannot be combined with --retention-* flags")
	}
	retention := map[string]any{}
	for _, field := range []struct {
		flag, key string
		value     *int32
	}{
		{"retention-ttl", "ttl", in.Retention.TTL},
		{"retention-successful-runs", "successfulRuns", in.Retention.SuccessfulRuns},
		{"retention-failed-runs", "failedRuns", in.Retention.FailedRuns},
	} {
		if changed(field.flag) {
			retention[field.key] = field.value
		}
	}
	if in.ClearRetention {
		patch["retention"] = json.RawMessage("null")
	} else if len(retention) > 0 {
		if err := setJSON(patch, "retention", retention); err != nil {
			return nil, err
		}
	}
	return patch, nil
}
