## iai

InteractiveAI's CLI

### Synopsis

InteractiveAI's CLI to interact with its platform.

Use the subcommands below to manage your organizations, projects, services, secrets, and other components.

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


### Options

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.dev.interactive.ai")
  -h, --help                         help for iai
      --hostname string              Hostname for the API (default "https://dev.interactive.ai")
```

### SEE ALSO

* [iai completion](iai_completion.md)	 - Generate the autocompletion script for the specified shell
* [iai images](iai_images.md)	 - Build and manage container images
* [iai login](iai_login.md)	 - Log in to InteractiveAI with a email and password
* [iai logout](iai_logout.md)	 - Log out of InteractiveAI
* [iai logs](iai_logs.md)	 - Show logs for a specific replica
* [iai organizations](iai_organizations.md)	 - Manage organizations
* [iai projects](iai_projects.md)	 - Manage projects
* [iai replicas](iai_replicas.md)	 - Manage service replicas
* [iai secrets](iai_secrets.md)	 - Manage secrets
* [iai services](iai_services.md)	 - Manage services

