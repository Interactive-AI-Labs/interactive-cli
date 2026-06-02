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

* [iai agents](iai_agents.md)	 - Deploy AI agents with policies, routines, and tools
* [iai comments](iai_comments.md)	 - Annotate traces, observations, and sessions
* [iai completion](iai_completion.md)	 - Generate the autocompletion script for the specified shell
* [iai connectors](iai_connectors.md)	 - Manage MCP connectors in a project
* [iai databases](iai_databases.md)	 - PostgreSQL instances with extension support, including pgvector
* [iai dataset-items](iai_dataset-items.md)	 - Manage items in evaluation datasets
* [iai dataset-runs](iai_dataset-runs.md)	 - Run evaluations against datasets
* [iai datasets](iai_datasets.md)	 - Create and list evaluation datasets
* [iai glossaries](iai_glossaries.md)	 - Domain vocabularies for consistent term interpretation
* [iai images](iai_images.md)	 - Build and push container images
* [iai login](iai_login.md)	 - Authenticate with InteractiveAI
* [iai logout](iai_logout.md)	 - Clear local session
* [iai macros](iai_macros.md)	 - Pre-approved response templates used in routines
* [iai metrics](iai_metrics.md)	 - Query aggregated observability metrics
* [iai observations](iai_observations.md)	 - Inspect spans within traces
* [iai organizations](iai_organizations.md)	 - Switch or list organizations
* [iai policies](iai_policies.md)	 - Single-step behavioral rules for agents
* [iai projects](iai_projects.md)	 - Switch or list projects
* [iai prompts](iai_prompts.md)	 - Versioned prompts for agents, evaluators, and guardrails
* [iai queue-items](iai_queue-items.md)	 - Manage items in annotation queues
* [iai queues](iai_queues.md)	 - Annotation queues for human review workflows
* [iai replicas](iai_replicas.md)	 - Inspect service replicas
* [iai routines](iai_routines.md)	 - Multi-step behavioral processes for agents
* [iai run-items](iai_run-items.md)	 - Inspect results of evaluation runs
* [iai score-configs](iai_score-configs.md)	 - Define scoring schemas for evaluation
* [iai scores](iai_scores.md)	 - Read and write evaluation scores
* [iai secrets](iai_secrets.md)	 - Encrypted key-value pairs for services and agents
* [iai services](iai_services.md)	 - Deploy and manage HTTP services
* [iai sessions](iai_sessions.md)	 - Browse trace-derived conversation sessions
* [iai skills](iai_skills.md)	 - Manage Interactive Copilot skills (not to be confused with context items that configure the Interactive Agent)
* [iai stacks](iai_stacks.md)	 - Declarative resource sync from config files
* [iai traces](iai_traces.md)	 - Browse agent decision traces with full attribution
* [iai update](iai_update.md)	 - Update iai to the latest version
* [iai variables](iai_variables.md)	 - Contextual attributes referenced in policies and routines

