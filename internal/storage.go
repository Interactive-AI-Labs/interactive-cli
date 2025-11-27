package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type sessionCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Domain   string    `json:"domain,omitempty"`
	Path     string    `json:"path,omitempty"`
	Expires  time.Time `json:"expires,omitempty"`
	Secure   bool      `json:"secure,omitempty"`
	HTTPOnly bool      `json:"http_only,omitempty"`
}

type sessionCookieFile struct {
	Cookies []sessionCookie `json:"cookies"`
	SavedAt time.Time       `json:"saved_at"`
}

func SaveSessionCookies(cookies []*http.Cookie, cfgDirName, sessionFileName string) error {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	dir := filepath.Join(home, cfgDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create session directory %q: %w", dir, err)
	}

	path := filepath.Join(dir, sessionFileName)

	fileData := sessionCookieFile{
		Cookies: make([]sessionCookie, 0, len(cookies)),
		SavedAt: time.Now().UTC(),
	}

	for _, c := range cookies {
		if c == nil {
			continue
		}
		sc := sessionCookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HttpOnly,
		}
		if !c.Expires.IsZero() {
			sc.Expires = c.Expires
		}
		fileData.Cookies = append(fileData.Cookies, sc)
	}

	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode session cookies: %w", err)
	}

	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open temp session file %q: %w", tmp, err)
	}

	_, writeErr := f.Write(data)
	closeErr := f.Close()

	if writeErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to write session file: %w", writeErr)
	}
	if closeErr != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to close session file: %w", closeErr)
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to finalize session file: %w", err)
	}

	return nil
}

func LoadSessionCookies(cfgDirName, sessionFileName string) ([]*http.Cookie, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	path := filepath.Join(home, cfgDirName, sessionFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No session file, not an error
		}
		return nil, fmt.Errorf("failed to read session file %q: %w", path, err)
	}

	var fileData sessionCookieFile
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("failed to decode session cookies: %w", err)
	}

	cookies := make([]*http.Cookie, 0, len(fileData.Cookies))
	for _, sc := range fileData.Cookies {
		c := &http.Cookie{
			Name:     sc.Name,
			Value:    sc.Value,
			Domain:   sc.Domain,
			Path:     sc.Path,
			Secure:   sc.Secure,
			HttpOnly: sc.HTTPOnly,
		}
		if !sc.Expires.IsZero() {
			c.Expires = sc.Expires
		}
		cookies = append(cookies, c)
	}

	return cookies, nil
}
