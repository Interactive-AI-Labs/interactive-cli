package auth

import (
	"os/exec"
	"runtime"
)

// OpenBrowser opens the given URL in the user's default browser.
// Returns an error if the browser cannot be opened.
func OpenBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Run()
}
