package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/auth"
	"github.com/Interactive-AI-Labs/interactive-cli/internal/files"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const loginTimeout = 5 * time.Minute

var (
	loginInteractive bool
	loginDevice      bool
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to InteractiveAI",
	Long: `Log in to InteractiveAI.

By default, opens your browser for SSO login (Google, GitHub, or email/password).
If the browser cannot be opened, the CLI automatically falls back to the device flow.

Use --device for headless/SSH environments. This displays a scannable QR code
and a verification code to enter on another device.

Use --interactive (or -i) for the classic email/password prompt.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case loginInteractive:
			return runInteractiveLogin(cmd)
		case loginDevice:
			return runDeviceLogin(cmd)
		default:
			return runBrowserLogin(cmd)
		}
	},
}

func init() {
	loginCmd.Flags().BoolVarP(&loginInteractive, "interactive", "i", false,
		"Use email/password prompt instead of browser")
	loginCmd.Flags().BoolVar(&loginDevice, "device", false,
		"Use device code flow (for headless/SSH environments)")
	rootCmd.AddCommand(loginCmd)
}

// runBrowserLogin is the default: opens the browser with PKCE (US-009).
// Falls back to device flow if the browser cannot be opened.
func runBrowserLogin(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "Opening browser to log in... (press Ctrl+C to cancel)")

	result, err := auth.RunBrowserFlow(cmd.Context(), hostname, loginTimeout)
	if err != nil {
		// If browser failed to open, fall back to device flow
		var browserErr *auth.BrowserOpenError
		if errors.As(err, &browserErr) {
			fmt.Fprintln(out, "Could not open browser. Falling back to device code flow...")
			return runDeviceLogin(cmd)
		}
		return fmt.Errorf("login failed: %w", err)
	}

	return saveCookiesAndPrint(out, result.Cookies, result.Email)
}

// runDeviceLogin uses the device authorization flow (US-010).
func runDeviceLogin(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	printFn := func(msg string) {
		fmt.Fprint(out, msg)
	}

	result, err := auth.RunDeviceFlow(cmd.Context(), hostname, loginTimeout, printFn)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	return saveCookiesAndPrint(out, result.Cookies, result.Email)
}

// runInteractiveLogin is the legacy email/password flow (US-011).
func runInteractiveLogin(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	reader := bufio.NewReader(in)

	fmt.Fprint(out, "email: ")
	email, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read email: %w", err)
	}
	email = strings.TrimSpace(email)

	fmt.Fprint(out, "Password: ")
	var password string
	// stdin is a terminal: read password with echo disabled
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		raw, err := term.ReadPassword(int(f.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Fprintln(out)
		password = string(raw)
	} else {
		// stdin is not a terminal: read as plain text
		p, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(p)
	}

	if email == "" || password == "" {
		return fmt.Errorf("email and password cannot be empty")
	}

	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to encode request body: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/auth/signin", hostname)
	req, err := http.NewRequestWithContext(
		cmd.Context(),
		http.MethodPost,
		url,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: defaultHTTPTimeout,
	}

	fmt.Fprintln(out, "Logging in to InteractiveAI...")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("login failed with status %s", resp.Status)
	}

	cookies := resp.Cookies()
	return saveCookiesAndPrint(out, cookies, email)
}

func saveCookiesAndPrint(out io.Writer, cookies []*http.Cookie, email string) error {
	if len(cookies) == 0 {
		fmt.Fprintln(out,
			"Warning: login succeeded but no cookies were returned by the server.")
		return nil
	}

	if err := files.SaveSessionCookies(cookies, cfgDirName, sessionFileName); err != nil {
		return fmt.Errorf("login succeeded but failed to store session cookies: %w", err)
	}

	if email != "" {
		fmt.Fprintf(out, "Logged in as %s\n", email)
	} else {
		fmt.Fprintf(out, "Login successful. %d cookie(s) stored for future commands.\n",
			len(cookies))
	}
	return nil
}
