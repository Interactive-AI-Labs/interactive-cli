## iai secrets update

Update keys in a secret

### Synopsis

Update one or more keys in an existing secret.

By default, only the specified keys are updated (merge/upsert). Existing keys
not included in the update are preserved.

With --replace, ALL secret data is replaced. Any keys not included in the new
data will be permanently deleted.

The project is selected with --project or via 'iai projects select'.

Secret data can be provided via:
  --data KEY=VALUE         (can be repeated)
  --from-env-file FILE     (KEY=VALUE pairs, one per line)

When both are provided, --data values take precedence.

Examples:
  # Update a single key (other keys preserved)
  iai secrets update my-secret -d API_KEY=new-value

  # Update multiple keys (other keys preserved)
  iai secrets update my-secret -d API_KEY=val1 -d DB_PASS=val2

  # Replace all keys (keys not provided will be deleted)
  iai secrets update my-secret -d API_KEY=val1 --replace

```
iai secrets update <secret_name> [flags]
```

### Options

```
  -d, --data stringArray       Secret data in KEY=VALUE form (repeatable)
      --from-env-file string   Path to env file with KEY=VALUE pairs (one per line)
  -h, --help                   help for update
  -o, --organization string    Organization name that owns the project
  -p, --project string         Project name that owns the secrets
      --replace                Replace all secret data (keys not provided will be deleted)
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai secrets](iai_secrets.md)	 - Manage secrets

