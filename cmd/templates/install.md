## Install

The CLI is distributed through Go's package manager, so it must first be installed. Click on [this](https://go.dev/doc/install) link and follow the instructions to do so.

To validate the installation run:

```bash
go version
```

Once Go is installed, ensure Go binaries are in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Add this line to your shell profile (~/.bashrc, ~/.zshrc, etc.) to make it permanent.

Now install InteractiveAI's CLI with the following command:

```bash
go install github.com/Interactive-AI-Labs/interactive-cli/cmd/iai@latest
```

Verify the installation by running:

```bash
iai --help
```

---

