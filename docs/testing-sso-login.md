# Testing SSO Browser & Device Flow Login (DEV-202)

## Prerequisites

- **platform-backend** running locally (default: `http://localhost:3000`)
- **CLI** built from the `DEV-202/cli-sso-browser-auth` branch

### Build the CLI

```bash
git checkout DEV-202/cli-sso-browser-auth
go build -o iai .
```

### Point CLI to local backend

All commands below use the `--hostname` flag to target your local backend:

```bash
./iai login --hostname http://localhost:3000
```

Alternatively, export the env var once:

```bash
export INTERACTIVE_HOSTNAME=http://localhost:3000
```

---

## Test Cases

### 1. Browser SSO Login (default flow)

This is the default when no flags are passed.

```bash
./iai login --hostname http://localhost:3000
```

**Expected behavior:**

1. CLI prints `Opening browser to log in...`
2. Browser opens to the platform SSO page (Google, GitHub, or email/password)
3. After completing login in the browser, the callback page shows **"Login successful!"**
4. CLI prints `Logged in as <email>` and returns to the prompt
5. Session cookies are saved to `~/.interactiveai/session_cookies.json`

**Verify session persists:**

```bash
./iai organizations list --hostname http://localhost:3000
```

### 2. Browser Login — User Denies / Cancels in Browser

1. Run `./iai login --hostname http://localhost:3000`
2. When the browser opens, deny access or navigate away without completing login

**Expected behavior:**

- Browser shows **"Login failed"** error page (not the success page)
- CLI prints an error message (e.g., `authorization error: access_denied`)

### 3. Browser Login — Ctrl+C Cancellation

1. Run `./iai login --hostname http://localhost:3000`
2. Press **Ctrl+C** before completing login in the browser

**Expected behavior:**

- CLI exits cleanly without hanging
- No orphan processes or goroutine leaks

### 4. Browser Fails to Open → Device Flow Fallback

Simulate a missing browser by overriding the PATH:

```bash
PATH=/usr/bin:/bin ./iai login --hostname http://localhost:3000
```

Or test on a headless/SSH environment.

**Expected behavior:**

1. CLI prints `Could not open browser. Falling back to device code flow...`
2. CLI displays a verification URL and user code
3. Opening the URL on another device and entering the code completes login

### 5. Device Code Flow (explicit)

```bash
./iai login --device --hostname http://localhost:3000
```

**Expected behavior:**

1. CLI prints a verification URL and a user code
2. CLI prints `Waiting for authorization... (press Ctrl+C to cancel)`
3. Open the verification URL in any browser, enter the code
4. CLI prints `Logged in as <email>` once authorized
5. Session cookies are saved

**Additional checks:**

- Pressing Ctrl+C during polling cancels cleanly
- Waiting past the expiry prints `login session expired`

### 6. Interactive Email/Password Login (legacy)

```bash
./iai login -i --hostname http://localhost:3000
```

**Expected behavior:**

1. CLI prompts for `email:` then `Password:` (password input is hidden)
2. On valid credentials, prints `Logged in as <email>`
3. On invalid credentials, prints `login failed with status 401 Unauthorized` (or similar)

### 7. Logout

```bash
./iai logout
```

**Expected behavior:**

- Session cookies file is removed
- Subsequent authenticated commands fail with an auth error

---

## Quick Checklist

| # | Scenario | Command | Pass? |
|---|----------|---------|-------|
| 1 | Browser SSO login (happy path) | `./iai login` | |
| 2 | User denies auth in browser | deny in browser | |
| 3 | Ctrl+C during browser wait | Ctrl+C | |
| 4 | Fallback to device flow | no browser available | |
| 5 | Device flow (explicit) | `./iai login --device` | |
| 6 | Device flow Ctrl+C | Ctrl+C during poll | |
| 7 | Device flow expiry | wait for timeout | |
| 8 | Interactive login (valid creds) | `./iai login -i` | |
| 9 | Interactive login (bad creds) | `./iai login -i` | |
| 10 | Logout clears session | `./iai logout` | |
