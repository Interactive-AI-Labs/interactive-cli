## iai

InteractiveAI's CLI

### Synopsis

InteractiveAI's CLI to interact with its platform.

Use the subcommands below to manage your organizations, projects, agents, services, secrets, prompts, routines, policies, variables, glossaries, macros, and other components.

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
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
  -h, --help                         help for iai
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai agents](iai_agents.md)	 - Manage agents
* [iai comments](iai_comments.md)	 - Manage comments
* [iai completion](iai_completion.md)	 - Generate the autocompletion script for the specified shell
* [iai dataset-items](iai_dataset-items.md)	 - Manage dataset items
* [iai dataset-runs](iai_dataset-runs.md)	 - Manage dataset runs
* [iai datasets](iai_datasets.md)	 - Manage evaluation datasets
* [iai glossaries](iai_glossaries.md)	 - Manage glossary definitions
* [iai images](iai_images.md)	 - Build and manage container images
* [iai login](iai_login.md)	 - Log in to InteractiveAI
* [iai logout](iai_logout.md)	 - Log out of InteractiveAI
* [iai macros](iai_macros.md)	 - Manage macros
* [iai metrics](iai_metrics.md)	 - Manage observability metrics
* [iai observations](iai_observations.md)	 - Manage observations
* [iai organizations](iai_organizations.md)	 - Manage organizations
* [iai policies](iai_policies.md)	 - Manage policies
* [iai projects](iai_projects.md)	 - Manage projects
* [iai prompts](iai_prompts.md)	 - Manage prompts
* [iai queue-items](iai_queue-items.md)	 - Manage annotation queue items
* [iai queues](iai_queues.md)	 - Manage annotation queues
* [iai replicas](iai_replicas.md)	 - Manage service replicas
* [iai routines](iai_routines.md)	 - Manage routines
* [iai run-items](iai_run-items.md)	 - Manage dataset run items
* [iai score-configs](iai_score-configs.md)	 - Manage score configs
* [iai scores](iai_scores.md)	 - Manage scores
* [iai secrets](iai_secrets.md)	 - Manage secrets
* [iai services](iai_services.md)	 - Manage services
* [iai sessions](iai_sessions.md)	 - Manage sessions
* [iai stacks](iai_stacks.md)	 - Manage stacks
* [iai traces](iai_traces.md)	 - Manage traces
* [iai variables](iai_variables.md)	 - Manage variables
* [iai vector-stores](iai_vector-stores.md)	 - Manage vector stores

