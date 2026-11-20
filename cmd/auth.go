package cmd

import (
	"fmt"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/auth"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
)

// requireLoginAuth refuses action unless the command runs as a user (a token or iai login).
func requireLoginAuth(action string) error {
	refusal := auth.LoginRequiredError(action)
	// A token wins over an API key, as it does on every request.
	if token != "" {
		return nil
	}
	if apiKey != "" {
		return refusal
	}

	cookies, err := files.LoadSessionCookies(cfgDirName, sessionFileName)
	if err != nil {
		return fmt.Errorf("failed to load login session: %w", err)
	}
	if len(cookies) == 0 {
		return refusal
	}

	return nil
}
