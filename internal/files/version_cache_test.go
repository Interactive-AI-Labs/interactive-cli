package files

import (
	"os"
	"path/filepath"
	"testing"
)

// tempCfgDir points HOME at a temp dir and returns the cfg dir name and path.
func tempCfgDir(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", filepath.Dir(dir))
	return filepath.Base(dir), dir
}

func TestVersionCache(t *testing.T) {
	t.Run("misses when no cache exists", func(t *testing.T) {
		cfgDirName, _ := tempCfgDir(t)
		if c, ok := ReadVersionCache(cfgDirName); ok {
			t.Fatalf("expected cache miss, got %+v", c)
		}
	})

	t.Run("round-trips a written entry", func(t *testing.T) {
		cfgDirName, _ := tempCfgDir(t)
		in := VersionCache{LatestVersion: "0.29.0", CheckedAt: 1700000000, NotifiedAt: 1700000100}
		WriteVersionCache(cfgDirName, in)
		out, ok := ReadVersionCache(cfgDirName)
		if !ok || out != in {
			t.Fatalf("ReadVersionCache() = (%+v, %v), want (%+v, true)", out, ok, in)
		}
	})

	t.Run("misses on corrupt entry", func(t *testing.T) {
		cfgDirName, dir := tempCfgDir(t)
		path := filepath.Join(dir, versionCacheFile)
		if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		if _, ok := ReadVersionCache(cfgDirName); ok {
			t.Fatal("expected cache miss on corrupt entry")
		}
	})
}
