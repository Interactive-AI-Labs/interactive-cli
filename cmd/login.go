package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	internal "github.com/Interactive-AI-Labs/interactive-cli/internal"
	"github.com/spf13/cobra"
)

const (
	signInURL = "https://dev.interactive.ai/api/v1/auth/signin"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to InteractiveAI with a email and password",
	Long:  `Log in to InteractiveAI with a email and password`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		in := cmd.InOrStdin()

		// 1. Prompt for email and password.
		reader := bufio.NewReader(in)

		fmt.Fprint(out, "email: ")
		email, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
		email = strings.TrimSpace(email)

		fmt.Fprint(out, "Password: ")
		password, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(password)

		if email == "" || password == "" {
			return fmt.Errorf("email and password cannot be empty")
		}

		// 2. Prepare and send the sign-in request.
		reqBody := map[string]string{
			"email":    email,
			"password": password,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to encode request body: %w", err)
		}

		req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, signInURL, bytes.NewReader(bodyBytes))
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

		// 3. Handle non-successful status codes.
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			respText := strings.TrimSpace(string(respBody))

			if respText != "" {
				fmt.Fprintf(out, "Server response:\n%s\n", respText)
			}

			return fmt.Errorf("login failed with status %s", resp.Status)
		}

		// 4. Extract cookies from the response.
		cookies := resp.Cookies()
		if len(cookies) == 0 {
			fmt.Fprintln(out, "Warning: login succeeded but no cookies were returned by the server.")
		} else {
			if err := internal.SaveSessionCookies(cookies, cfgDirName, sessionFileName); err != nil {
				return fmt.Errorf("login succeeded but failed to store session cookies: %w", err)
			}
			fmt.Fprintf(out, "Login successful. %d cookie(s) stored for future commands.\n", len(cookies))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
