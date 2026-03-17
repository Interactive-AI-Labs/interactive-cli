## iai glossaries delete

Delete a glossary

### Synopsis

Delete a glossary definition and all its versions, or delete specific versions.

Without flags, deletes the glossary entry and all its versions. Use --version to
delete a specific version, or --label to delete versions with a specific label.

The project is selected with --project or via 'iai projects select'.

Examples:
  iai glossaries delete finance-terms
  iai glossaries delete finance-terms --version 3
  iai glossaries delete finance-terms --label staging

```
iai glossaries delete <name> [flags]
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
```

### SEE ALSO

* [iai glossaries](iai_glossaries.md)	 - Manage glossary definitions

