package files

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestLiveCookies(t *testing.T) {
	tests := []struct {
		name       string
		setCookies []string
		want       []string
	}{
		{
			name: "login replaces legacy cookies under the same name",
			setCookies: []string{
				"__Secure-next-auth.session-token=; Max-Age=0; Path=/; Secure",
				"next-auth.session-token=; Max-Age=0; Path=/",
				"__Secure-next-auth.session-token=123:abc; Path=/; Secure; HttpOnly",
			},
			want: []string{"__Secure-next-auth.session-token=123:abc; Path=/; Secure; HttpOnly"},
		},
		{
			name:       "only deletions",
			setCookies: []string{"next-auth.csrf-token=; Max-Age=0; Path=/"},
			want:       []string{},
		},
		{
			name:       "cookie without Max-Age",
			setCookies: []string{"session=abc; Path=/"},
			want:       []string{"session=abc; Path=/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LiveCookies(parseSetCookies(t, tt.setCookies))

			if want := parseSetCookies(t, tt.want); !reflect.DeepEqual(got, want) {
				t.Errorf("LiveCookies() = %+v, want %+v", got, want)
			}
		})
	}
}

func parseSetCookies(t *testing.T, lines []string) []*http.Cookie {
	t.Helper()
	cookies := make([]*http.Cookie, 0, len(lines))
	for _, line := range lines {
		c, err := http.ParseSetCookie(line)
		if err != nil {
			t.Fatalf("ParseSetCookie(%q) error = %v", line, err)
		}
		cookies = append(cookies, c)
	}
	return cookies
}

func TestSaveSessionCookies(t *testing.T) {
	t.Run("saves cookies successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-session-" + t.Name()
		sessionFile := "session.json"
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		cookies := []*http.Cookie{
			{
				Name:     "session",
				Value:    "abc123",
				Domain:   "example.com",
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
			},
		}

		err := SaveSessionCookies(cookies, cfgDir, sessionFile)
		if err != nil {
			t.Fatalf("SaveSessionCookies() error = %v", err)
		}

		sessionPath := filepath.Join(testHome, cfgDir, sessionFile)
		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			t.Errorf("session file was not created at %q", sessionPath)
		}
	})

	t.Run("handles nil cookies", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-session-" + t.Name()
		sessionFile := "session.json"
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		cookies := []*http.Cookie{nil, {Name: "valid", Value: "value"}, nil}
		err := SaveSessionCookies(cookies, cfgDir, sessionFile)
		if err != nil {
			t.Fatalf("SaveSessionCookies() error = %v", err)
		}

		loaded, err := LoadSessionCookies(cfgDir, sessionFile)
		if err != nil {
			t.Fatalf("LoadSessionCookies() error = %v", err)
		}
		if len(loaded) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(loaded))
		}
	})

	t.Run("saves cookies with expiry", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-session-" + t.Name()
		sessionFile := "session.json"
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		expiry := time.Now().Add(24 * time.Hour).UTC()
		cookies := []*http.Cookie{
			{
				Name:    "session",
				Value:   "abc123",
				Expires: expiry,
			},
		}

		err := SaveSessionCookies(cookies, cfgDir, sessionFile)
		if err != nil {
			t.Fatalf("SaveSessionCookies() error = %v", err)
		}

		loaded, err := LoadSessionCookies(cfgDir, sessionFile)
		if err != nil {
			t.Fatalf("LoadSessionCookies() error = %v", err)
		}
		if len(loaded) != 1 {
			t.Fatalf("expected 1 cookie, got %d", len(loaded))
		}

		if loaded[0].Expires.IsZero() {
			t.Error("expected expiry to be set")
		}
	})
}

func TestLoadSessionCookies(t *testing.T) {
	t.Run("returns nil when file does not exist", func(t *testing.T) {
		cookies, err := LoadSessionCookies("nonexistent-dir-12345", "session.json")
		if err != nil {
			t.Fatalf("LoadSessionCookies() error = %v, want nil", err)
		}
		if cookies != nil {
			t.Errorf("LoadSessionCookies() = %v, want nil", cookies)
		}
	})

	t.Run("loads saved cookies", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-session-" + t.Name()
		sessionFile := "session.json"
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		expected := []*http.Cookie{
			{
				Name:     "session",
				Value:    "abc123",
				Domain:   "example.com",
				Path:     "/api",
				Secure:   true,
				HttpOnly: false,
			},
			{
				Name:   "token",
				Value:  "xyz789",
				Domain: "api.example.com",
			},
		}

		if err := SaveSessionCookies(expected, cfgDir, sessionFile); err != nil {
			t.Fatalf("SaveSessionCookies() error = %v", err)
		}

		loaded, err := LoadSessionCookies(cfgDir, sessionFile)
		if err != nil {
			t.Fatalf("LoadSessionCookies() error = %v", err)
		}

		if len(loaded) != len(expected) {
			t.Fatalf("loaded %d cookies, want %d", len(loaded), len(expected))
		}

		for i, cookie := range loaded {
			if cookie.Name != expected[i].Name {
				t.Errorf("cookie[%d].Name = %q, want %q", i, cookie.Name, expected[i].Name)
			}
			if cookie.Value != expected[i].Value {
				t.Errorf("cookie[%d].Value = %q, want %q", i, cookie.Value, expected[i].Value)
			}
			if cookie.Domain != expected[i].Domain {
				t.Errorf("cookie[%d].Domain = %q, want %q", i, cookie.Domain, expected[i].Domain)
			}
			if cookie.Path != expected[i].Path {
				t.Errorf("cookie[%d].Path = %q, want %q", i, cookie.Path, expected[i].Path)
			}
			if cookie.Secure != expected[i].Secure {
				t.Errorf("cookie[%d].Secure = %v, want %v", i, cookie.Secure, expected[i].Secure)
			}
			if cookie.HttpOnly != expected[i].HttpOnly {
				t.Errorf(
					"cookie[%d].HttpOnly = %v, want %v",
					i,
					cookie.HttpOnly,
					expected[i].HttpOnly,
				)
			}
		}
	})
}

func TestDeleteSessionCookies(t *testing.T) {
	t.Run("deletes existing session file", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfgDir := ".test-session-" + t.Name()
		sessionFile := "session.json"
		home := os.Getenv("HOME")
		testHome := filepath.Join(tmpDir, "home")
		os.Setenv("HOME", testHome)
		defer os.Setenv("HOME", home)

		cookies := []*http.Cookie{{Name: "session", Value: "abc123"}}
		if err := SaveSessionCookies(cookies, cfgDir, sessionFile); err != nil {
			t.Fatalf("SaveSessionCookies() error = %v", err)
		}

		sessionPath := filepath.Join(testHome, cfgDir, sessionFile)
		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			t.Fatal("session file was not created")
		}

		err := DeleteSessionCookies(cfgDir, sessionFile)
		if err != nil {
			t.Fatalf("DeleteSessionCookies() error = %v", err)
		}

		if _, err := os.Stat(sessionPath); !os.IsNotExist(err) {
			t.Error("session file still exists after deletion")
		}
	})

	t.Run("returns nil when file does not exist", func(t *testing.T) {
		err := DeleteSessionCookies("nonexistent-dir-12345", "session.json")
		if err != nil {
			t.Errorf("DeleteSessionCookies() error = %v, want nil", err)
		}
	})
}
