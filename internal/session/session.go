package session

import (
	"fmt"
	"os"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
)

type Session struct {
	cfgDirName      string
	envOrganization string
	envProject      string
}

func NewSession(cfgDirName string) *Session {
	return &Session{
		cfgDirName:      cfgDirName,
		envOrganization: strings.TrimSpace(os.Getenv("INTERACTIVE_ORGANIZATION")),
		envProject:      strings.TrimSpace(os.Getenv("INTERACTIVE_PROJECT")),
	}
}

// ResolveOrganization returns the organization name using this precedence:
// flagOrg > cfgOrg > INTERACTIVE_ORGANIZATION > selectedOrg.
// Returns an error if all are empty.
func (s *Session) ResolveOrganization(cfgOrg, flagOrg string) (string, error) {
	cfgOrg = strings.TrimSpace(cfgOrg)
	flagOrg = strings.TrimSpace(flagOrg)

	if flagOrg != "" {
		return flagOrg, nil
	}
	if cfgOrg != "" {
		return cfgOrg, nil
	}
	if s.envOrganization != "" {
		return s.envOrganization, nil
	}

	selectedOrg, err := files.GetSelectedOrg(s.cfgDirName)
	if err != nil {
		return "", fmt.Errorf("failed to load selected organization: %w", err)
	}

	selectedOrg = strings.TrimSpace(selectedOrg)
	if selectedOrg != "" {
		return selectedOrg, nil
	}

	return "", fmt.Errorf(
		"organization is required: provide via --organization flag, --cfg-file, INTERACTIVE_ORGANIZATION, or run 'iai organizations select'",
	)
}

// ResolveProject returns the project name using this precedence:
// flagProject > cfgProject > INTERACTIVE_PROJECT > selectedProject.
// Returns error if all are empty.
func (s *Session) ResolveProject(cfgProject, flagProject string) (string, error) {
	flagProject = strings.TrimSpace(flagProject)
	cfgProject = strings.TrimSpace(cfgProject)

	if flagProject != "" {
		return flagProject, nil
	}
	if cfgProject != "" {
		return cfgProject, nil
	}
	if s.envProject != "" {
		return s.envProject, nil
	}

	selectedProject, err := files.GetSelectedProject(s.cfgDirName)
	if err != nil {
		return "", fmt.Errorf("failed to load selected project: %w", err)
	}

	selectedProject = strings.TrimSpace(selectedProject)
	if selectedProject != "" {
		return selectedProject, nil
	}

	return "", fmt.Errorf(
		"project is required: provide via --project flag, --cfg-file, INTERACTIVE_PROJECT, or run 'iai projects select'",
	)
}
