package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionCacheReadWrite(t *testing.T) {
	dir := t.TempDir()
	cfgDirName := filepath.Base(dir)
	t.Setenv("HOME", filepath.Dir(dir))

	// No cache yet
	if c, ok := ReadVersionCache(cfgDirName); ok {
		t.Fatalf("expected cache miss, got %+v", c)
	}

	// Write and read back
	in := VersionCache{LatestVersion: "0.29.0", CheckedAt: 1700000000, NotifiedAt: 1700000100}
	WriteVersionCache(cfgDirName, in)
	out, ok := ReadVersionCache(cfgDirName)
	if !ok || out != in {
		t.Fatalf("ReadVersionCache() = (%+v, %v), want (%+v, true)", out, ok, in)
	}

	// Corrupt cache
	path := filepath.Join(dir, versionCacheFile)
	_ = os.WriteFile(path, []byte("{"), 0o644)
	if _, ok := ReadVersionCache(cfgDirName); ok {
		t.Fatal("expected cache miss on corrupt entry")
	}
}
