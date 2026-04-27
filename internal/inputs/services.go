package inputs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/utils"
)

type ServiceInput struct {
	Port            int
	ImageType       string
	ImageRepository string
	ImageName       string
	ImageTag        string
	Memory          string
	CPU             string
	Endpoint        bool

	Replicas          int
	AutoscalingMin    int
	AutoscalingMax    int
	AutoscalingCPU    int
	AutoscalingMemory int

	EnvVars    []string
	SecretRefs []string

	HealthcheckPath         string
	HealthcheckInitialDelay int

	ScheduleUptime   string
	ScheduleDowntime string
	ScheduleTimezone string
}

func BuildServiceRequestBody(in ServiceInput) (clients.CreateServiceBody, error) {
	if err := ValidateServiceEnvVars(in.EnvVars); err != nil {
		return clients.CreateServiceBody{}, err
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
		return clients.CreateServiceBody{}, err
	}
	var secretRefs []clients.SecretRef
	for _, name := range in.SecretRefs {
		secretRefs = append(secretRefs, clients.SecretRef{
			SecretName: strings.TrimSpace(name),
		})
	}

	reqBody := clients.CreateServiceBody{
		ServicePort: in.Port,
		Image: clients.ImageSpec{
			Type:       in.ImageType,
			Repository: in.ImageRepository,
			Name:       in.ImageName,
			Tag:        in.ImageTag,
		},
		Resources: clients.Resources{
			Memory: in.Memory,
			CPU:    in.CPU,
		},
		Env:        env,
		SecretRefs: secretRefs,
		Endpoint:   in.Endpoint,
	}

	if in.AutoscalingMin > 0 || in.AutoscalingMax > 0 || in.AutoscalingCPU > 0 ||
		in.AutoscalingMemory > 0 {
		as := &clients.Autoscaling{
			MinReplicas: in.AutoscalingMin,
			MaxReplicas: in.AutoscalingMax,
		}
		if in.AutoscalingCPU > 0 {
			as.CPUPercentage = utils.ToPtr(in.AutoscalingCPU)
		}
		if in.AutoscalingMemory > 0 {
			as.MemoryPercentage = utils.ToPtr(in.AutoscalingMemory)
		}
		reqBody.Autoscaling = as
	}
	if in.Replicas > 0 {
		reqBody.Replicas = in.Replicas
	}

	if in.HealthcheckPath != "" || in.HealthcheckInitialDelay != 0 {
		hc := &clients.Healthcheck{
			Path: in.HealthcheckPath,
		}
		if in.HealthcheckInitialDelay > 0 {
			hc.InitialDelaySeconds = utils.ToPtr(in.HealthcheckInitialDelay)
		}
		reqBody.Healthcheck = hc
	}

	if in.ScheduleUptime != "" || in.ScheduleDowntime != "" || in.ScheduleTimezone != "" {
		reqBody.Schedule = &clients.Schedule{
			Uptime:   in.ScheduleUptime,
			Downtime: in.ScheduleDowntime,
			Timezone: in.ScheduleTimezone,
		}
	}

	return reqBody, nil
}

// ServiceUpdateFlags is the set of cobra flag names BuildServiceUpdatePatch
// inspects via the `changed` predicate. Keep in sync with cmd/services.go.
var ServiceUpdateFlags = struct {
	Port              string
	ImageType         string
	ImageRepository   string
	ImageName         string
	ImageTag          string
	Memory            string
	CPU               string
	Replicas          string
	AutoscalingMin    string
	AutoscalingMax    string
	AutoscalingCPU    string
	AutoscalingMemory string
	Env               string
	Secret            string
	Endpoint          string
	HealthcheckPath   string
	HealthcheckDelay  string
	ScheduleUptime    string
	ScheduleDowntime  string
	ScheduleTimezone  string
}{
	Port:              "port",
	ImageType:         "image-type",
	ImageRepository:   "image-repository",
	ImageName:         "image-name",
	ImageTag:          "image-tag",
	Memory:            "memory",
	CPU:               "cpu",
	Replicas:          "replicas",
	AutoscalingMin:    "autoscaling-min-replicas",
	AutoscalingMax:    "autoscaling-max-replicas",
	AutoscalingCPU:    "autoscaling-cpu-percentage",
	AutoscalingMemory: "autoscaling-memory-percentage",
	Env:               "env",
	Secret:            "secret",
	Endpoint:          "endpoint",
	HealthcheckPath:   "healthcheck-path",
	HealthcheckDelay:  "healthcheck-initial-delay",
	ScheduleUptime:    "schedule-uptime",
	ScheduleDowntime:  "schedule-downtime",
	ScheduleTimezone:  "schedule-timezone",
}

// BuildServiceUpdatePatch produces a partial-update body containing only the
// fields whose flags the user explicitly set. `changed` reports whether a flag
// name was provided on the command line (typically cmd.Flags().Changed).
func BuildServiceUpdatePatch(
	in ServiceInput,
	clearEnv, clearSecret, clearHealthcheck, clearSchedule bool,
	changed func(string) bool,
) (clients.UpdatePatch, error) {
	f := ServiceUpdateFlags
	patch := clients.UpdatePatch{}

	autoscalingFlags := []string{
		f.AutoscalingMin, f.AutoscalingMax, f.AutoscalingCPU, f.AutoscalingMemory,
	}
	if changed(f.Replicas) && anyChanged(changed, autoscalingFlags...) {
		return nil, fmt.Errorf(
			"--replicas and --autoscaling-* flags are mutually exclusive",
		)
	}
	if clearHealthcheck &&
		(changed(f.HealthcheckPath) || changed(f.HealthcheckDelay)) {
		return nil, fmt.Errorf(
			"--clear-healthcheck cannot be combined with --healthcheck-* flags",
		)
	}

	if changed(f.Port) {
		if err := setJSON(patch, "servicePort", in.Port); err != nil {
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
	if err := setEndpointPatch(patch, in.Endpoint, changed(f.Endpoint)); err != nil {
		return nil, err
	}

	if changed(f.Replicas) {
		if err := setJSON(patch, "replicas", in.Replicas); err != nil {
			return nil, err
		}
	}

	if anyChanged(changed, autoscalingFlags...) {
		as := map[string]any{}
		if changed(f.AutoscalingMin) {
			as["minReplicas"] = in.AutoscalingMin
		}
		if changed(f.AutoscalingMax) {
			as["maxReplicas"] = in.AutoscalingMax
		}
		if changed(f.AutoscalingCPU) {
			as["cpuPercentage"] = in.AutoscalingCPU
		}
		if changed(f.AutoscalingMemory) {
			as["memoryPercentage"] = in.AutoscalingMemory
		}
		if err := setJSON(patch, "autoscaling", as); err != nil {
			return nil, err
		}
	}

	switch {
	case clearHealthcheck:
		patch["healthcheck"] = json.RawMessage("null")
	case anyChanged(changed, f.HealthcheckPath, f.HealthcheckDelay):
		hc := map[string]any{}
		if changed(f.HealthcheckPath) {
			hc["path"] = in.HealthcheckPath
		}
		if changed(f.HealthcheckDelay) {
			hc["initialDelaySeconds"] = in.HealthcheckInitialDelay
		}
		if err := setJSON(patch, "healthcheck", hc); err != nil {
			return nil, err
		}
	}

	if err := setSchedulePatch(patch, ScheduleInput{
		Uptime:          in.ScheduleUptime,
		Downtime:        in.ScheduleDowntime,
		Timezone:        in.ScheduleTimezone,
		UptimeChanged:   changed(f.ScheduleUptime),
		DowntimeChanged: changed(f.ScheduleDowntime),
		TimezoneChanged: changed(f.ScheduleTimezone),
		Clear:           clearSchedule,
	}); err != nil {
		return nil, err
	}

	return patch, nil
}

func ValidateServiceEnvVars(envVars []string) error {
	for _, e := range envVars {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return fmt.Errorf("invalid --env value %q; expected NAME=VALUE", e)
		}
	}
	return nil
}

func ValidateServiceSecretRefs(secretRefs []string) error {
	for _, name := range secretRefs {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return fmt.Errorf("invalid --secret value %q; name must not be empty", name)
		}
	}
	return nil
}
