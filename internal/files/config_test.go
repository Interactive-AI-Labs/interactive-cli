package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("returns empty config when file does not exist", func(t *testing.T) {
		cfg, err := LoadConfig("nonexistent-dir-12345")
		if err != nil {
			t.Fatalf("LoadConfig() error = %v, want nil", err)
		}
		if cfg.SelectedOrg != "" {
			t.Errorf("LoadConfig() SelectedOrg = %q, want empty", cfg.SelectedOrg)
		}
	})

	t.Run("loads valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-config-" + t.Name()
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		orgName := "test-org"
		cfg := &Config{SelectedOrg: orgName}
		if err := SaveConfig(cfgDir, cfg); err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		loaded, err := LoadConfig(cfgDir)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if loaded.SelectedOrg != orgName {
			t.Errorf("LoadConfig() SelectedOrg = %q, want %q", loaded.SelectedOrg, orgName)
		}
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("saves config successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-config-" + t.Name()
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		cfg := &Config{SelectedOrg: "my-org"}
		err := SaveConfig(cfgDir, cfg)
		if err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		cfgPath := filepath.Join(testHome, cfgDir, configFileName)
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			t.Errorf("config file was not created at %q", cfgPath)
		}
	})

	t.Run("saves nil config as empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-config-" + t.Name()
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		err := SaveConfig(cfgDir, nil)
		if err != nil {
			t.Fatalf("SaveConfig() error = %v", err)
		}

		loaded, err := LoadConfig(cfgDir)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if loaded.SelectedOrg != "" {
			t.Errorf("expected empty SelectedOrg, got %q", loaded.SelectedOrg)
		}
	})
}

func TestGetSelectedOrg(t *testing.T) {
	t.Run("returns empty when no org selected", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-config-" + t.Name()
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		org, err := GetSelectedOrg(cfgDir)
		if err != nil {
			t.Fatalf("GetSelectedOrg() error = %v", err)
		}
		if org != "" {
			t.Errorf("GetSelectedOrg() = %q, want empty", org)
		}
	})

	t.Run("returns selected org", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-config-" + t.Name()
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		expected := "my-org"
		if err := SelectOrg(cfgDir, expected); err != nil {
			t.Fatalf("SelectOrg() error = %v", err)
		}

		org, err := GetSelectedOrg(cfgDir)
		if err != nil {
			t.Fatalf("GetSelectedOrg() error = %v", err)
		}
		if org != expected {
			t.Errorf("GetSelectedOrg() = %q, want %q", org, expected)
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-config-" + t.Name()
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		if err := SelectOrg(cfgDir, "  my-org  "); err != nil {
			t.Fatalf("SelectOrg() error = %v", err)
		}

		org, err := GetSelectedOrg(cfgDir)
		if err != nil {
			t.Fatalf("GetSelectedOrg() error = %v", err)
		}
		if org != "my-org" {
			t.Errorf("GetSelectedOrg() = %q, want %q", org, "my-org")
		}
	})
}

func TestSelectOrg(t *testing.T) {
	t.Run("sets organization", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-config-" + t.Name()
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		expected := "new-org"
		err := SelectOrg(cfgDir, expected)
		if err != nil {
			t.Fatalf("SelectOrg() error = %v", err)
		}

		cfg, err := LoadConfig(cfgDir)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.SelectedOrg != expected {
			t.Errorf("SelectOrg() set SelectedOrg = %q, want %q", cfg.SelectedOrg, expected)
		}
	})

	t.Run("clears organization with empty string", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-config-" + t.Name()
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		if err := SelectOrg(cfgDir, "some-org"); err != nil {
			t.Fatalf("SelectOrg() error = %v", err)
		}

		if err := SelectOrg(cfgDir, ""); err != nil {
			t.Fatalf("SelectOrg() error = %v", err)
		}

		cfg, err := LoadConfig(cfgDir)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.SelectedOrg != "" {
			t.Errorf("SelectOrg() SelectedOrg = %q, want empty", cfg.SelectedOrg)
		}
	})
}
