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

---

## 8. Development and Testing

### 8.1 Running Tests

This project includes both unit tests and integration tests, organized in separate files with build tags.

#### Unit Tests

Unit tests run fast and don't require external dependencies. They're located in `*_test.go` files without build tags.

Run all unit tests:

```bash
make test-unit
# or
go test -v ./...
```

#### Integration Tests

Integration tests use the `integration` build tag and may make real HTTP calls to test servers. They're located in `*_integration_test.go` files.

Run integration tests:

```bash
make test-integration
# or
go test -v -tags=integration ./...
```

#### All Tests

Run both unit and integration tests:

```bash
make test-all
```

#### Test Coverage

Generate a coverage report for unit tests only:

```bash
make test-coverage
```

This creates `coverage.html` which you can open in a browser to view detailed coverage information.

Generate a coverage report including integration tests:

```bash
make test-coverage-all
```

This creates `coverage-all.html` with comprehensive coverage data from both unit and integration tests.

### 8.2 Test Organization

- **Unit tests**: No build tags, run by default with `go test ./...`
- **Integration tests**: Use `//go:build integration` and `// +build integration` tags
- Test files follow these naming conventions:
  - `*_test.go` - Unit tests
  - `*_integration_test.go` - Integration tests

### 8.3 Writing Tests

When adding new tests:

1. Unit tests should be fast, deterministic, and not depend on external services
2. Integration tests should use the `integration` build tag at the top of the file:
   ```go
   //go:build integration
   // +build integration
   
   package mypackage
   ```
3. Use `t.TempDir()` for tests that need temporary directories
4. Follow Go testing best practices and table-driven test patterns where appropriate
