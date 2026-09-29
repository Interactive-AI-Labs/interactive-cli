package inputs

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

type JobInput struct {
	Type                                            string
	ImageType, ImageRepository, ImageName, ImageTag string
	Command, Args                                   []string
	Script, Pyproject                               string
	Memory, CPU                                     string
	EnvVars, SecretRefs                             []string
	StackId                                         string
	Schedule, Timezone                              string
	Timeout                                         *int64
	Retries                                         *int32
	Retention                                       deployment.JobRetention
}

func ValidateJobArgs(args []string) error {
	for _, arg := range args {
		if strings.TrimSpace(arg) == "" {
			return fmt.Errorf("arguments must not be empty")
		}
	}
	return nil
}

// ReadJobFile reads a script or project file without changing its contents.
func ReadJobFile(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read job file %q: %w", path, err)
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("job file %q must contain valid UTF-8", path)
	}
	return string(data), nil
}

func BuildJobRequestBody(in JobInput) (deployment.CreateJobBody, error) {
	body := deployment.CreateJobBody{
		Type:      in.Type,
		Resources: deployment.Resources{CPU: in.CPU, Memory: in.Memory},
		StackId:   in.StackId,
		Timeout:   in.Timeout,
		Retries:   in.Retries,
	}
	switch in.Type {
	case "image":
		if in.Script != "" || in.Pyproject != "" {
			return deployment.CreateJobBody{}, fmt.Errorf(
				"--script and --pyproject are only available for script jobs",
			)
		}
		body.Image = &deployment.ImageSpec{
			Type:       in.ImageType,
			Repository: in.ImageRepository,
			Name:       in.ImageName,
			Tag:        in.ImageTag,
		}
		body.Command, body.Args = in.Command, in.Args
	case "script":
		if in.ImageType != "" || in.ImageRepository != "" || in.ImageName != "" ||
			in.ImageTag != "" ||
			len(in.Command) != 0 ||
			len(in.Args) != 0 {
			return deployment.CreateJobBody{}, fmt.Errorf(
				"image, command and args flags are only available for image jobs",
			)
		}
		if strings.TrimSpace(in.Script) == "" || strings.TrimSpace(in.Pyproject) == "" {
			return deployment.CreateJobBody{}, fmt.Errorf(
				"--script and --pyproject are required for script jobs",
			)
		}
		body.Script, body.Pyproject = in.Script, in.Pyproject
	default:
		return deployment.CreateJobBody{}, fmt.Errorf("--type must be image or script")
	}

	if err := ValidateServiceEnvVars(in.EnvVars); err != nil {
		return deployment.CreateJobBody{}, err
	}
	for _, value := range in.EnvVars {
		name, value, _ := strings.Cut(value, "=")
		body.Env = append(body.Env, deployment.EnvVar{Name: strings.TrimSpace(name), Value: value})
	}
	if err := ValidateServiceSecretRefs(in.SecretRefs); err != nil {
		return deployment.CreateJobBody{}, err
	}
	for _, name := range in.SecretRefs {
		body.SecretRefs = append(
			body.SecretRefs,
			deployment.SecretRef{SecretName: strings.TrimSpace(name)},
		)
	}

	if in.Timezone != "" && in.Schedule == "" {
		return deployment.CreateJobBody{}, fmt.Errorf("--schedule-timezone requires --schedule")
	}
	if in.Schedule != "" {
		body.Cron = &deployment.JobCron{Schedule: in.Schedule, Timezone: in.Timezone}
	}
	if in.Retention.TTL != nil || in.Retention.SuccessfulRuns != nil ||
		in.Retention.FailedRuns != nil {
		body.Retention = &in.Retention
	}
	return body, nil
}
