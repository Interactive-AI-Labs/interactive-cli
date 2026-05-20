package files

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVersionCacheReadWrite(t *testing.T) {
	dir := t.TempDir()
	cfgDirName := filepath.Base(dir)
	t.Setenv("HOME", filepath.Dir(dir))

	// No cache yet
	v, ok := ReadVersionCache(cfgDirName)
	if ok {
		t.Fatalf("expected cache miss, got %q", v)
	}

	// Write and read back
	WriteVersionCache(cfgDirName, "0.29.0")
	v, ok = ReadVersionCache(cfgDirName)
	if !ok || v != "0.29.0" {
		t.Fatalf("ReadVersionCache() = (%q, %v), want (\"0.29.0\", true)", v, ok)
	}

	// Expired cache
	path := filepath.Join(dir, versionCacheFile)
	expired := versionCache{
		LatestVersion: "0.29.0",
		CheckedAt:     time.Now().Add(-25 * time.Hour).Unix(),
	}
	data, _ := json.Marshal(expired)
	_ = os.WriteFile(path, data, 0o644)

	_, ok = ReadVersionCache(cfgDirName)
	if ok {
		t.Fatal("expected cache miss on expired entry")
	}
}
