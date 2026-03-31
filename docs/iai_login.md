## iai login

Log in to InteractiveAI

### Synopsis

Log in to InteractiveAI.

By default, opens your browser for SSO login (Google, GitHub, or email/password).
If the browser cannot be opened, the CLI automatically falls back to the device flow.

Use --device for headless/SSH environments. This displays a scannable QR code
and a verification code to enter on another device.

Use --interactive (or -i) for the classic email/password prompt.

```
iai login [flags]
```

### Options

```
      --device        Use device code flow (for headless/SSH environments)
  -h, --help          help for login
  -i, --interactive   Use email/password prompt instead of browser
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
      --token string                 JWT Bearer token for user-level auth, issued via OAuth or copilot token exchange (env: INTERACTIVE_TOKEN)
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI

