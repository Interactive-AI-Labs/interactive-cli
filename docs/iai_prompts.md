## iai prompts

Versioned prompts for agents, evaluators, and guardrails

### Synopsis

Manage general-purpose text and chat prompts in InteractiveAI projects.

Unlike typed commands (routines, policies, glossaries, variables, macros),
prompts managed here have no enforced schema or structure. They support two
types: "text" (default) and "chat".

### Options

```
  -h, --help   help for prompts
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai](iai.md)	 - InteractiveAI's CLI
* [iai prompts create](iai_prompts_create.md)	 - Create a prompt
* [iai prompts delete](iai_prompts_delete.md)	 - Delete a prompt
* [iai prompts describe](iai_prompts_describe.md)	 - Describe a prompt in detail
* [iai prompts diff](iai_prompts_diff.md)	 - Compare two versions of a prompt
* [iai prompts labels](iai_prompts_labels.md)	 - Manage labels on prompt versions
* [iai prompts list](iai_prompts_list.md)	 - List prompts in a project
* [iai prompts update](iai_prompts_update.md)	 - Update a prompt (creates a new version)
* [iai prompts versions](iai_prompts_versions.md)	 - List versions of a prompt

