package files

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	versionCacheFile = "latest_version"
	versionCacheTTL  = 24 * time.Hour
)

type versionCache struct {
	LatestVersion string `json:"latest_version"`
	CheckedAt     int64  `json:"checked_at"`
}

func versionCachePath(cfgDirName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, cfgDirName, versionCacheFile)
}

func ReadVersionCache(cfgDirName string) (string, bool) {
	path := versionCachePath(cfgDirName)
	if path == "" {
		return "", false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	var c versionCache
	if err := json.Unmarshal(data, &c); err != nil {
		return "", false
	}
	if time.Since(time.Unix(c.CheckedAt, 0)) > versionCacheTTL {
		return "", false
	}
	return c.LatestVersion, true
}

func WriteVersionCache(cfgDirName, version string) {
	path := versionCachePath(cfgDirName)
	if path == "" {
		return
	}

	c := versionCache{
		LatestVersion: version,
		CheckedAt:     time.Now().Unix(),
	}
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, data, 0o644)
}
