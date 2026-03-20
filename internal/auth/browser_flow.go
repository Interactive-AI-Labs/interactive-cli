package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BrowserFlowResult holds the session cookies after a successful browser login.
type BrowserFlowResult struct {
	Cookies []*http.Cookie
	Email   string
}

// authorizeResponse matches the backend POST /api/v1/auth/cli/authorize response.
type authorizeResponse struct {
	AuthURL     string `json:"auth_url"`
	CliAuthCode string `json:"cli_auth_code"`
	ExpiresIn   int    `json:"expires_in"`
}

// tokenResponse matches the backend POST /api/v1/auth/cli/token response.
type tokenResponse struct {
	Success bool   `json:"success"`
	Email   string `json:"email"`
	Error   string `json:"error"`
}

// RunBrowserFlow performs the full localhost callback + PKCE auth flow (US-009).
// It returns the session cookies on success, or an error.
// If the browser fails to open, it returns a *BrowserOpenError so the caller
// can fall back to the device flow.
func RunBrowserFlow(
	ctx context.Context,
	hostname string,
	timeout time.Duration,
) (*BrowserFlowResult, error) {
	// 1. Generate PKCE parameters
	codeVerifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeChallenge := CodeChallengeS256(codeVerifier)

	state, err := GenerateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// 2. Start localhost callback server
	callbackServer, err := NewCallbackServer()
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server: %w", err)
	}

	callbackServer.Start()

	redirectURI := callbackServer.RedirectURI()

	// 3. Call POST /api/v1/auth/cli/authorize
	authReq := map[string]string{
		"redirect_uri":          redirectURI,
		"code_challenge":        codeChallenge,
		"code_challenge_method": "S256",
		"state":                 state,
	}
	body, err := json.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode authorize request: %w", err)
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	authorizeReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		hostname+"/api/v1/auth/cli/authorize",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorize request: %w", err)
	}
	authorizeReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(authorizeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth initiation failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var authResp authorizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	// 4. Open browser
	if err := OpenBrowser(authResp.AuthURL); err != nil {
		callbackServer.Shutdown(context.Background())
		return nil, &BrowserOpenError{Err: err}
	}

	// 5. Wait for callback
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := callbackServer.Wait(timeoutCtx)

	// 6. Shut down localhost server immediately after receiving callback
	callbackServer.Shutdown(context.Background())

	if result.Err != nil {
		return nil, fmt.Errorf("waiting for browser callback: %w", result.Err)
	}

	// 7. Validate state
	if result.State != state {
		return nil, fmt.Errorf("state mismatch: possible CSRF attack")
	}

	// 8. Exchange authorization code for session cookies
	return exchangeCode(ctx, hostname, result.Code, codeVerifier, authResp.CliAuthCode)
}

func exchangeCode(
	ctx context.Context,
	hostname, code, codeVerifier, cliAuthCode string,
) (*BrowserFlowResult, error) {
	tokenReq := map[string]string{
		"code":          code,
		"code_verifier": codeVerifier,
		"cli_auth_code": cliAuthCode,
	}
	body, err := json.Marshal(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode token request: %w", err)
	}

	// Use a raw http.Client that does NOT follow redirects, so we can capture Set-Cookie headers
	noRedirectClient := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	tokenReqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost,
		hostname+"/api/v1/auth/cli/token",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	tokenReqHTTP.Header.Set("Content-Type", "application/json")
	resp, err := noRedirectClient.Do(tokenReqHTTP)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	cookies := resp.Cookies()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	if !tokenResp.Success {
		return nil, fmt.Errorf("token exchange failed: %s", tokenResp.Error)
	}

	return &BrowserFlowResult{
		Cookies: cookies,
		Email:   tokenResp.Email,
	}, nil
}

// BrowserOpenError indicates the browser could not be opened.
// The caller should fall back to the device flow.
type BrowserOpenError struct {
	Err error
}

func (e *BrowserOpenError) Error() string {
	return fmt.Sprintf("failed to open browser: %v", e.Err)
}

func (e *BrowserOpenError) Unwrap() error {
	return e.Err
}
