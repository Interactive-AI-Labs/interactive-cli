## iai policies delete

Delete a policy

### Synopsis

Delete a policy and all its versions, or delete specific versions.

Without flags, deletes the policy and all its versions (requires confirmation).
Use --version to delete a specific version, or --label to delete versions with a
specific label. Use -f to skip the confirmation prompt.

Examples:
  iai policies delete safety-rules
  iai policies delete safety-rules -f
  iai policies delete safety-rules --version 3
  iai policies delete safety-rules --label staging

```
iai policies delete <name> [flags]
```

### Options

```
  -f, --force                 Skip confirmation prompt
  -h, --help                  help for delete
      --label string          Delete versions with this label only
  -o, --organization string   Organization name that owns the project
  -p, --project string        Project name that owns the prompts
      --version int           Delete a specific version only
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

* [iai policies](iai_policies.md)	 - Manage policies

