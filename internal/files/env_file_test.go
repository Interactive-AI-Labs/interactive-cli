package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	t.Run("parses valid env file", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := `KEY1=value1
KEY2=value2
KEY3=value with spaces`
		if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		result, err := ParseEnvFile(envFile)
		if err != nil {
			t.Fatalf("ParseEnvFile() error = %v", err)
		}

		expected := map[string]string{
			"KEY1": "value1",
			"KEY2": "value2",
			"KEY3": "value with spaces",
		}

		if len(result) != len(expected) {
			t.Errorf("got %d entries, want %d", len(result), len(expected))
		}

		for key, want := range expected {
			got, ok := result[key]
			if !ok {
				t.Errorf("missing key %q", key)
				continue
			}
			if got != want {
				t.Errorf("result[%q] = %q, want %q", key, got, want)
			}
		}
	})

	t.Run("skips empty lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := `KEY1=value1

KEY2=value2

`
		if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		result, err := ParseEnvFile(envFile)
		if err != nil {
			t.Fatalf("ParseEnvFile() error = %v", err)
		}

		if len(result) != 2 {
			t.Errorf("got %d entries, want 2", len(result))
		}
	})

	t.Run("skips comments", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := `# This is a comment
KEY1=value1
# Another comment
KEY2=value2`
		if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		result, err := ParseEnvFile(envFile)
		if err != nil {
			t.Fatalf("ParseEnvFile() error = %v", err)
		}

		if len(result) != 2 {
			t.Errorf("got %d entries, want 2", len(result))
		}
	})

	t.Run("handles values with equals sign", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := `KEY1=value=with=equals`
		if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		result, err := ParseEnvFile(envFile)
		if err != nil {
			t.Fatalf("ParseEnvFile() error = %v", err)
		}

		if result["KEY1"] != "value=with=equals" {
			t.Errorf("result[KEY1] = %q, want %q", result["KEY1"], "value=with=equals")
		}
	})

	t.Run("trims whitespace from keys and values", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := `  KEY1  =  value1  `
		if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		result, err := ParseEnvFile(envFile)
		if err != nil {
			t.Fatalf("ParseEnvFile() error = %v", err)
		}

		if result["KEY1"] != "value1" {
			t.Errorf("result[KEY1] = %q, want %q", result["KEY1"], "value1")
		}
	})

	t.Run("returns error for missing equals", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := `KEY1=value1
KEY2_NO_EQUALS
KEY3=value3`
		if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := ParseEnvFile(envFile)
		if err == nil {
			t.Fatal("ParseEnvFile() expected error, got nil")
		}

		if !strings.Contains(err.Error(), "missing '=' separator") {
			t.Errorf("error should mention 'missing '=' separator', got: %v", err)
		}
		if !strings.Contains(err.Error(), "line 2") {
			t.Errorf("error should mention line 2, got: %v", err)
		}
	})

	t.Run("returns error for empty key", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		content := `KEY1=value1
=value_no_key
KEY3=value3`
		if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := ParseEnvFile(envFile)
		if err == nil {
			t.Fatal("ParseEnvFile() expected error, got nil")
		}

		if !strings.Contains(err.Error(), "empty key") {
			t.Errorf("error should mention 'empty key', got: %v", err)
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		_, err := ParseEnvFile("/nonexistent/file.env")
		if err == nil {
			t.Fatal("ParseEnvFile() expected error, got nil")
		}

		if !strings.Contains(err.Error(), "failed to open file") {
			t.Errorf("error should mention 'failed to open file', got: %v", err)
		}
	})

	t.Run("handles empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte(""), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		result, err := ParseEnvFile(envFile)
		if err != nil {
			t.Fatalf("ParseEnvFile() error = %v", err)
		}

		if len(result) != 0 {
			t.Errorf("got %d entries, want 0", len(result))
		}
	})
}
