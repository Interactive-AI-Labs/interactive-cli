package versioncheck

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
)

const (
	goProxyURL   = "https://proxy.golang.org/github.com/!interactive-!a!i-!labs/interactive-cli/@latest"
	fetchTimeout = 10 * time.Second
)

type proxyInfo struct {
	Version string `json:"Version"`
}

// FetchLatestVersion queries the Go module proxy for the latest tagged version.
// A zero timeout uses the default fetch timeout.
func FetchLatestVersion(timeout time.Duration) (string, error) {
	if timeout == 0 {
		timeout = fetchTimeout
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(goProxyURL)
	if err != nil {
		if os.IsTimeout(err) {
			return "", fmt.Errorf("module proxy request timed out after %s", timeout)
		}
		return "", fmt.Errorf("failed to reach module proxy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("module proxy returned %s", resp.Status)
	}

	var info proxyInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("failed to parse proxy response: %w", err)
	}

	return strings.TrimPrefix(info.Version, "v"), nil
}

// GetLatestVersion returns the latest version, using a local cache to avoid
// hitting the network on every invocation. cfgDirName is the dotfile directory
// name (e.g. ".interactiveai").
func GetLatestVersion(cfgDirName string) (string, error) {
	if v, ok := files.ReadVersionCache(cfgDirName); ok {
		return v, nil
	}

	v, err := FetchLatestVersion(0)
	if err != nil {
		return "", err
	}

	files.WriteVersionCache(cfgDirName, v)
	return v, nil
}

// IsNewer returns true if latest is a higher semver than current.
// Prerelease latest versions are never considered newer (users should only
// be nudged toward stable releases). A prerelease current is considered
// older than its matching stable latest (e.g. 0.28.1-beta.1 < 0.28.1).
func IsNewer(current, latest string) bool {
	if isPrerelease(latest) {
		return false
	}
	cur := parseSemver(current)
	lat := parseSemver(latest)
	if cur == nil || lat == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if lat[i] > cur[i] {
			return true
		}
		if lat[i] < cur[i] {
			return false
		}
	}
	// Base versions are equal — current is older only if it's a prerelease.
	return isPrerelease(current)
}

// IsGoVersionSufficient returns true if goVer >= minVer.
func IsGoVersionSufficient(goVer, minVer string) bool {
	got := parseSemver(goVer)
	need := parseSemver(minVer)
	if got == nil || need == nil {
		return true
	}
	for i := 0; i < 3; i++ {
		if got[i] > need[i] {
			return true
		}
		if got[i] < need[i] {
			return false
		}
	}
	return true
}

func isPrerelease(v string) bool {
	v = strings.TrimPrefix(v, "v")
	return strings.Contains(v, "-")
}

func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	// Strip pre-release suffix for comparison
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) == 2 {
		parts = append(parts, "0")
	}
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}
