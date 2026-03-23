package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
)

// DeviceFlowResult holds the session cookies after a successful device login.
type DeviceFlowResult struct {
	Cookies []*http.Cookie
	Email   string
}

// deviceAuthorizeResponse matches POST /api/v1/auth/cli/device/authorize.
type deviceAuthorizeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// deviceTokenResponse matches POST /api/v1/auth/cli/device/token.
type deviceTokenResponse struct {
	Success bool   `json:"success"`
	Email   string `json:"email"`
	Error   string `json:"error"`
}

// RunDeviceFlow performs the device authorization flow (US-010).
// It displays the user code, then polls until authorization completes or times out.
// The printFn is called to display messages to the user.
func RunDeviceFlow(
	ctx context.Context,
	hostname string,
	timeout time.Duration,
	printFn func(string),
) (*DeviceFlowResult, error) {
	httpClient := &http.Client{Timeout: 15 * time.Second}

	// 1. Request device code
	deviceReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		hostname+"/api/v1/auth/cli/device/authorize",
		bytes.NewReader([]byte("{}")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create device authorize request: %w", err)
	}
	deviceReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(deviceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"device authorize failed (%d): %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	var deviceResp deviceAuthorizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode device response: %w", err)
	}

	// 2. Display instructions
	printFn("\n")
	printFn("  Scan the QR code or open the link to sign in:\n\n")
	qrterminal.GenerateWithConfig(deviceResp.VerificationURIComplete, qrterminal.Config{
		Level:      qrterminal.L,
		Writer:     os.Stdout,
		HalfBlocks: true,
		QuietZone:  1,
	})
	printFn(fmt.Sprintf("\n  Your code: %s\n\n", deviceResp.UserCode))
	printFn(fmt.Sprintf("  Verification URL:  %s\n", deviceResp.VerificationURI))
	printFn(fmt.Sprintf("  Direct link:       %s\n", deviceResp.VerificationURIComplete))
	printFn("\n  Waiting for authorization... (press Ctrl+C to cancel)\n")

	// 3. Poll — use server-provided interval with a 2s floor for responsiveness
	interval := time.Duration(deviceResp.Interval) * time.Second
	if interval < 2*time.Second {
		interval = 2 * time.Second
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("login session expired. Please run 'iai login' again")
		case <-ticker.C:
			result, status, err := pollDeviceToken(timeoutCtx, hostname, deviceResp.DeviceCode)
			if err != nil {
				return nil, err
			}

			switch status {
			case "completed":
				return result, nil
			case "authorization_pending":
				// Continue polling
			case "slow_down":
				// Increase interval
				interval += 5 * time.Second
				ticker.Reset(interval)
			case "expired_token":
				return nil, fmt.Errorf("login session expired. Please run 'iai login' again")
			case "invalid_grant":
				return nil, fmt.Errorf("device code already used. Please run 'iai login' again")
			default:
				return nil, fmt.Errorf("unexpected polling status: %s", status)
			}
		}
	}
}

func pollDeviceToken(
	ctx context.Context,
	hostname, deviceCode string,
) (*DeviceFlowResult, string, error) {
	reqBody := map[string]string{"device_code": deviceCode}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encode poll request: %w", err)
	}

	// Use a client that does NOT follow redirects to capture Set-Cookie headers
	noRedirectClient := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	pollReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		hostname+"/api/v1/auth/cli/device/token",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create poll request: %w", err)
	}
	pollReq.Header.Set("Content-Type", "application/json")
	resp, err := noRedirectClient.Do(pollReq)
	if err != nil {
		return nil, "", fmt.Errorf("polling failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		cookies := resp.Cookies()
		var tokenResp deviceTokenResponse
		if err := json.Unmarshal(respBody, &tokenResp); err != nil {
			return nil, "", fmt.Errorf("failed to decode token response: %w", err)
		}
		if !tokenResp.Success {
			return nil, "", fmt.Errorf("token exchange failed: %s", tokenResp.Error)
		}
		return &DeviceFlowResult{
			Cookies: cookies,
			Email:   tokenResp.Email,
		}, "completed", nil

	case 428: // authorization_pending
		return nil, "authorization_pending", nil

	case http.StatusTooManyRequests: // slow_down
		return nil, "slow_down", nil

	case http.StatusGone: // expired
		return nil, "expired_token", nil

	case http.StatusBadRequest: // invalid_grant
		return nil, "invalid_grant", nil

	case http.StatusNotFound:
		return nil, "", fmt.Errorf("device code not found")

	default:
		return nil, "", fmt.Errorf(
			"unexpected response (%d): %s",
			resp.StatusCode,
			string(respBody),
		)
	}
}
