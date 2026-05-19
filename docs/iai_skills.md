## iai skills

Manage Interactive Copilot skills (not to be confused with context items that configure the Interactive Agent)

### Synopsis

Manage Interactive Copilot skills for the interactive-copilot service.

IMPORTANT: These are Interactive Copilot skills, not to be confused with
context items that configure the Interactive Agent. Skills are loaded by the
Copilot runtime and injected into Copilot conversations as context. They
have no effect on the Interactive Agent.

Each Copilot skill is a free-form markdown bundle. It carries a short description
and an "intents" list of natural-language triggers (stored in config.skill) that
the Copilot uses to route incoming queries to the right skill at runtime.

### Options

```
  -h, --help   help for skills
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
* [iai skills create](iai_skills_create.md)	 - Create a skill
* [iai skills delete](iai_skills_delete.md)	 - Delete a skill
* [iai skills describe](iai_skills_describe.md)	 - Describe a skill in detail
* [iai skills diff](iai_skills_diff.md)	 - Compare two versions of a skill
* [iai skills list](iai_skills_list.md)	 - List skills in a project
* [iai skills update](iai_skills_update.md)	 - Update a skill (creates a new version)
* [iai skills versions](iai_skills_versions.md)	 - List versions of a skill

