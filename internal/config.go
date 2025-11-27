package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileName = "config.yaml"

// Config represents user-level CLI configuration persisted on disk.
type Config struct {
	// SelectedOrganization holds the human-readable organization name
	// chosen by the user via a future "select-organization" command.
	SelectedOrganization string `yaml:"selected_organization,omitempty"`
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

// SaveConfig writes the YAML configuration to the user's home directory,
// creating the directory if necessary.
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

// GetSelectedOrganization returns the currently configured selected
// organization name (trimmed), or an empty string when no organization
// has been chosen yet.
func GetSelectedOrganization(cfgDirName string) (string, error) {
	cfg, err := LoadConfig(cfgDirName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(cfg.SelectedOrganization), nil
}

// SetSelectedOrganization updates the selected organization name in the
// config file. An empty name clears the selection.
func SetSelectedOrganization(cfgDirName, orgName string) error {
	cfg, err := LoadConfig(cfgDirName)
	if err != nil {
		return err
	}

	cfg.SelectedOrganization = strings.TrimSpace(orgName)
	return SaveConfig(cfgDirName, cfg)
}
