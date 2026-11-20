package cmd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
)

func TestRequireLoginAuth(t *testing.T) {
	const refused = "this command requires iai login or JWT authentication; " +
		"API key authentication is not supported for report changes"
	tests := []struct {
		name, apiKey, token string
		session             bool
		want                string
	}{
		{name: "api key", apiKey: "pk:sk", want: refused},
		{name: "api key beside a session", apiKey: "pk:sk", session: true, want: refused},
		{name: "token beside an api key", apiKey: "pk:sk", token: "jwt", want: "<nil>"},
		{name: "token", token: "jwt", want: "<nil>"},
		{name: "session", session: true, want: "<nil>"},
		{name: "no auth", want: refused},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			savedKey, savedToken := apiKey, token
			t.Cleanup(func() { apiKey, token = savedKey, savedToken })
			apiKey, token = tt.apiKey, tt.token
			if tt.session {
				cookie := &http.Cookie{Name: "session", Value: "s"}
				if err := files.SaveSessionCookies(
					[]*http.Cookie{cookie},
					cfgDirName,
					sessionFileName,
				); err != nil {
					t.Fatal(err)
				}
			}
			if got := fmt.Sprint(requireLoginAuth("report changes")); got != tt.want {
				t.Errorf("requireLoginAuth() = %s, want %s", got, tt.want)
			}
		})
	}
}
