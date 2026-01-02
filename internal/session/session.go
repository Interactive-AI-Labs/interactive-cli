package session

import (
	"fmt"
	"strings"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
)

type Session struct {
	cfgDirName string
}

func NewSession(cfgDirName string) *Session {
	return &Session{
		cfgDirName: cfgDirName,
	}
}

// ResolveOrganization returns the organization name using this precedence:
// cfgOrg > flagOrg > selectedOrg.
// Returns error if all three are empty.
func (s *Session) ResolveOrganization(cfgOrg, flagOrg string) (string, error) {
	cfgOrg = strings.TrimSpace(cfgOrg)
	flagOrg = strings.TrimSpace(flagOrg)

	if cfgOrg != "" {
		return cfgOrg, nil
	}
	if flagOrg != "" {
		return flagOrg, nil
	}

	selectedOrg, err := files.GetSelectedOrg(s.cfgDirName)
	if err != nil {
		return "", fmt.Errorf("failed to load selected organization: %w", err)
	}

	selectedOrg = strings.TrimSpace(selectedOrg)
	if selectedOrg != "" {
		return selectedOrg, nil
	}

	return "", fmt.Errorf("organization is required: provide via --organization flag, --cfg-file, or run 'iai organizations select'")
}

// ResolveProject returns the project name using this precedence:
// cfgProject > flagProject > selectedProject.
// Returns error if all are empty.
func (s *Session) ResolveProject(cfgProject, flagProject string) (string, error) {
	cfgProject = strings.TrimSpace(cfgProject)
	flagProject = strings.TrimSpace(flagProject)

	if cfgProject != "" {
		return cfgProject, nil
	}
	if flagProject != "" {
		return flagProject, nil
	}

	selectedProject, err := files.GetSelectedProject(s.cfgDirName)
	if err != nil {
		return "", fmt.Errorf("failed to load selected project: %w", err)
	}

	selectedProject = strings.TrimSpace(selectedProject)
	if selectedProject != "" {
		return selectedProject, nil
	}

	return "", fmt.Errorf("project is required: provide via --project flag, --cfg-file, or run 'iai projects select'")
}
