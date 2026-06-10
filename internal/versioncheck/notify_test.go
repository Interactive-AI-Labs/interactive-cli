package versioncheck

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
)

func TestPendingNotification(t *testing.T) {
	dir := t.TempDir()
	cfgDirName := filepath.Base(dir)
	t.Setenv("HOME", filepath.Dir(dir))

	// No cache → nothing to announce.
	if _, ok := PendingNotification(cfgDirName, "0.34.0"); ok {
		t.Fatal("expected no notification without a cache")
	}

	files.WriteVersionCache(cfgDirName, files.VersionCache{
		LatestVersion: "0.35.0",
		CheckedAt:     time.Now().Unix(),
	})

	// First sighting of a newer release notifies and stamps the cache.
	latest, ok := PendingNotification(cfgDirName, "0.34.0")
	if !ok || latest != "0.35.0" {
		t.Fatalf("PendingNotification() = (%q, %v), want (\"0.35.0\", true)", latest, ok)
	}
	if c, _ := files.ReadVersionCache(cfgDirName); c.NotifiedAt == 0 {
		t.Fatal("expected NotifiedAt to be stamped")
	}

	// Within the notify interval it stays quiet.
	if _, ok := PendingNotification(cfgDirName, "0.34.0"); ok {
		t.Fatal("expected notification to be throttled")
	}

	// Once the stamp ages out, it fires again.
	c, _ := files.ReadVersionCache(cfgDirName)
	c.NotifiedAt = time.Now().Add(-notifyInterval - time.Hour).Unix()
	files.WriteVersionCache(cfgDirName, c)
	if _, ok := PendingNotification(cfgDirName, "0.34.0"); !ok {
		t.Fatal("expected notification after the interval elapsed")
	}

	// Up to date → quiet regardless of stamps.
	if _, ok := PendingNotification(cfgDirName, "0.35.0"); ok {
		t.Fatal("expected no notification when already up to date")
	}
}

func TestRefreshCacheSkipsFreshCache(t *testing.T) {
	dir := t.TempDir()
	cfgDirName := filepath.Base(dir)
	t.Setenv("HOME", filepath.Dir(dir))

	in := files.VersionCache{
		LatestVersion: "9.9.9",
		CheckedAt:     time.Now().Unix(),
		NotifiedAt:    1700000000,
	}
	files.WriteVersionCache(cfgDirName, in)

	// A fresh cache must short-circuit before any network access.
	RefreshCache(cfgDirName)

	if out, _ := files.ReadVersionCache(cfgDirName); out != in {
		t.Fatalf("RefreshCache rewrote a fresh cache: %+v", out)
	}
}
