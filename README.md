# InteractiveAI CLI (macOS)

This repository contains the **InteractiveAI CLI**, a Go-based command-line tool for interacting with the Interactive AI platform.

These instructions are tailored to **macOS** users who want to install and run the CLI from source using Go.

---

## 1. Prerequisites

### 1.1 Install Homebrew (if you don’t have it)

Homebrew is the standard package manager for macOS.

Check if it’s installed:

```bash
brew --version
```

If that prints a version, you’re good. Otherwise, install Homebrew:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Follow the on-screen instructions to add Homebrew to your `PATH` if prompted.

---

## 2. Install Go with Homebrew

Use Homebrew to install Go:

```bash
brew install go
```

Verify your Go installation:

```bash
go version
```

You should see a recent Go version printed.

---

## 3. Install the InteractiveAI CLI with `go install`

You can install the CLI directly from this repository using `go install`.

From any directory, run:

```bash
go install github.com/Interactive-AI-Labs/interactive-cli/cmd/iai@latest
```

This will:

- Download the source.
- Build the CLI.
- Place the resulting binary in your Go bin directory (usually `$(go env GOPATH)/bin` or `$HOME/go/bin`).

---

## 4. Ensure the CLI is on your PATH

By default, Go places binaries in `$(go env GOPATH)/bin`. On macOS this is often:

```bash
$HOME/go/bin
```

Check your Go bin path:

```bash
go env GOPATH
```

Append `/bin` to that path and ensure it’s in your `PATH`. For example, if `GOPATH` is `/Users/you/go`, then:

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

(Use `~/.bash_profile` or `~/.bashrc` if you’re using `bash`.)

---

## 5. Verify the installation

After updating your `PATH`, you should be able to run the CLI directly:

```bash
iai --help
```

You should see the CLI usage information.

Common commands to get started:

```bash
iai login
iai organizations list
iai projects list
```

(Depending on the CLI version, some commands may require you to have selected an organization or to be logged in first.)

---

## 6. Updating the CLI

To update to the latest version, simply run:

```bash
go install github.com/Interactive-AI-Labs/interactive-cli/cmd/iai@latest
```

This will rebuild and reinstall the latest version over your existing binary.

---

## 7. Troubleshooting

- **Command not found (`iai`):**
  - Ensure `$(go env GOPATH)/bin` is in your `PATH`.
  - Open a new terminal session after updating your shell config.

- **Go not found (`go: command not found`):**
  - Make sure `brew install go` completed successfully.
  - Confirm `go` is on your `PATH`:
    ```bash
    which go
    ```

- **Network or module errors:**
  - Try again after checking your network connection.
  - Make sure you can reach GitHub from your machine.

If you continue to have issues, capture the exact command and error message and share it when asking for help.
