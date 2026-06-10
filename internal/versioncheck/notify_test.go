package versioncheck

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
)

// seedCache points HOME at a temp dir and writes the given cache entry,
// returning the cfg dir name to pass to the functions under test.
func seedCache(t *testing.T, c *files.VersionCache) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", filepath.Dir(dir))
	cfgDirName := filepath.Base(dir)
	if c != nil {
		files.WriteVersionCache(cfgDirName, *c)
	}
	return cfgDirName
}

func TestPendingNotification(t *testing.T) {
	t.Run("quiet without a cache", func(t *testing.T) {
		cfgDirName := seedCache(t, nil)
		if _, ok := PendingNotification(cfgDirName, "0.34.0"); ok {
			t.Fatal("expected no notification without a cache")
		}
	})

	t.Run("notifies on newer version and stamps cache", func(t *testing.T) {
		cfgDirName := seedCache(t, &files.VersionCache{LatestVersion: "0.35.0"})
		latest, ok := PendingNotification(cfgDirName, "0.34.0")
		if !ok || latest != "0.35.0" {
			t.Fatalf("PendingNotification() = (%q, %v), want (\"0.35.0\", true)", latest, ok)
		}
		if c, _ := files.ReadVersionCache(cfgDirName); c.NotifiedAt == 0 {
			t.Fatal("expected NotifiedAt to be stamped")
		}
	})

	t.Run("throttled within the notify interval", func(t *testing.T) {
		cfgDirName := seedCache(t, &files.VersionCache{
			LatestVersion: "0.35.0",
			NotifiedAt:    time.Now().Unix(),
		})
		if _, ok := PendingNotification(cfgDirName, "0.34.0"); ok {
			t.Fatal("expected notification to be throttled")
		}
	})

	t.Run("fires again after the interval elapses", func(t *testing.T) {
		cfgDirName := seedCache(t, &files.VersionCache{
			LatestVersion: "0.35.0",
			NotifiedAt:    time.Now().Add(-notifyInterval - time.Hour).Unix(),
		})
		if _, ok := PendingNotification(cfgDirName, "0.34.0"); !ok {
			t.Fatal("expected notification after the interval elapsed")
		}
	})

	t.Run("quiet when already up to date", func(t *testing.T) {
		cfgDirName := seedCache(t, &files.VersionCache{LatestVersion: "0.35.0"})
		if _, ok := PendingNotification(cfgDirName, "0.35.0"); ok {
			t.Fatal("expected no notification when already up to date")
		}
	})
}

func TestRefreshCache(t *testing.T) {
	t.Run("skips fetch when cache is fresh", func(t *testing.T) {
		in := files.VersionCache{
			LatestVersion: "9.9.9",
			CheckedAt:     time.Now().Unix(),
			NotifiedAt:    1700000000,
		}
		cfgDirName := seedCache(t, &in)

		// A fresh cache must short-circuit before any network access.
		RefreshCache(cfgDirName)

		if out, _ := files.ReadVersionCache(cfgDirName); out != in {
			t.Fatalf("RefreshCache rewrote a fresh cache: %+v", out)
		}
	})
}
