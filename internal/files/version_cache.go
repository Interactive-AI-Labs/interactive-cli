package files

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const versionCacheFile = "latest_version"

// VersionCache is the persisted state of the background update check.
// CheckedAt is when LatestVersion was last fetched from the module proxy;
// NotifiedAt is when the user was last shown the upgrade nudge.
type VersionCache struct {
	LatestVersion string `json:"latest_version"`
	CheckedAt     int64  `json:"checked_at"`
	NotifiedAt    int64  `json:"notified_at,omitempty"`
}

func versionCachePath(cfgDirName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, cfgDirName, versionCacheFile)
}

func ReadVersionCache(cfgDirName string) (VersionCache, bool) {
	path := versionCachePath(cfgDirName)
	if path == "" {
		return VersionCache{}, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return VersionCache{}, false
	}
	var c VersionCache
	if err := json.Unmarshal(data, &c); err != nil {
		return VersionCache{}, false
	}
	return c, true
}

func WriteVersionCache(cfgDirName string, c VersionCache) {
	path := versionCachePath(cfgDirName)
	if path == "" {
		return
	}

	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, data, 0o644)
}
