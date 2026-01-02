package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileName = "config.yaml"

type Config struct {
	SelectedOrg     string `yaml:"selected_organization,omitempty"`
	SelectedProject string `yaml:"selected_project,omitempty"`
}

func LoadConfig(cfgDirName string) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	path := filepath.Join(home, cfgDirName, configFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config yet; return an empty config.
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %w", path, err)
	}

	return &cfg, nil
}

func SaveConfig(cfgDirName string, cfg *Config) error {
	if cfg == nil {
		cfg = &Config{}
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	dir := filepath.Join(home, cfgDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory %q: %w", dir, err)
	}

	path := filepath.Join(dir, configFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open temp config file %q: %w", tmp, err)
	}

	_, writeErr := f.Write(data)
	closeErr := f.Close()

	if writeErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to write config file: %w", writeErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to close config file: %w", closeErr)
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to finalize config file: %w", err)
	}

	return nil
}

// GetSelectedOrg returns the currently configured selected
// organization name, or an empty string when no organization
// has been chosen yet.
func GetSelectedOrg(cfgDirName string) (string, error) {
	cfg, err := LoadConfig(cfgDirName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(cfg.SelectedOrg), nil
}

// GetSelectedProject returns the currently configured selected
// project name, or an empty string when no project
// has been chosen yet.
func GetSelectedProject(cfgDirName string) (string, error) {
	cfg, err := LoadConfig(cfgDirName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(cfg.SelectedProject), nil
}

// SelectProject updates the selected project name in the
// config file. An empty name clears the selection.
func SelectProject(cfgDirName, projectName string) error {
	cfg, err := LoadConfig(cfgDirName)
	if err != nil {
		return err
	}

	cfg.SelectedProject = strings.TrimSpace(projectName)
	return SaveConfig(cfgDirName, cfg)
}

// SelectOrg updates the selected organization name in the
// config file. An empty name clears the selection.
// When the organization changes, the selected project is cleared.
func SelectOrg(cfgDirName, orgName string) error {
	cfg, err := LoadConfig(cfgDirName)
	if err != nil {
		return err
	}

	newOrg := strings.TrimSpace(orgName)
	// If the organization is changing, clear the selected project
	if cfg.SelectedOrg != newOrg {
		cfg.SelectedProject = ""
	}
	cfg.SelectedOrg = newOrg
	return SaveConfig(cfgDirName, cfg)
}

// ResolveOrganization returns the organization name.
// Returns error if all three are empty.
func ResolveOrganization(cfgOrg, flagOrg, selectedOrg string) (string, error) {
	cfgOrg = strings.TrimSpace(cfgOrg)
	flagOrg = strings.TrimSpace(flagOrg)
	selectedOrg = strings.TrimSpace(selectedOrg)

	if cfgOrg != "" {
		return cfgOrg, nil
	}
	if flagOrg != "" {
		return flagOrg, nil
	}
	if selectedOrg != "" {
		return selectedOrg, nil
	}

	return "", fmt.Errorf("organization is required: provide via --organization flag, --cfg-file, or run 'iai organizations select'")
}

// ResolveProject returns the project name using this precedence:
// cfgProject > flagProject > selectedProject.
// Returns error if all are empty.
func ResolveProject(cfgProject, flagProject, selectedProject string) (string, error) {
	cfgProject = strings.TrimSpace(cfgProject)
	flagProject = strings.TrimSpace(flagProject)
	selectedProject = strings.TrimSpace(selectedProject)

	if cfgProject != "" {
		return cfgProject, nil
	}
	if flagProject != "" {
		return flagProject, nil
	}
	if selectedProject != "" {
		return selectedProject, nil
	}

	return "", fmt.Errorf("project is required: provide via --project flag, --cfg-file, or run 'iai projects select'")
}
