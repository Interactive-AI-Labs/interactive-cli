## iai collections slots

Manage a collection's vector slots and their indexes

### Synopsis

Add, reindex, vacuum, inspect, and remove the vector slots of a collection.

### Options

```
  -h, --help   help for slots
```

### Options inherited from parent commands

```
      --api-key string               API key for authentication
      --cfg-file string              Path to YAML config file with organization, project, and optional service definitions
      --deployment-hostname string   Hostname for the deployment API (default "https://deployment.interactive.ai")
      --hostname string              Hostname for the API (default "https://app.interactive.ai")
```

### SEE ALSO

* [iai collections](iai_collections.md)	 - Vector collections (knowledge bases) inside a pgvector database
* [iai collections slots add](iai_collections_slots_add.md)	 - Add a vector slot
* [iai collections slots delete](iai_collections_slots_delete.md)	 - Delete a vector slot
* [iai collections slots progress](iai_collections_slots_progress.md)	 - Show a slot's index build progress
* [iai collections slots reindex](iai_collections_slots_reindex.md)	 - Rebuild a slot's index (online)
* [iai collections slots vacuum](iai_collections_slots_vacuum.md)	 - Vacuum a slot (reclaim space, refresh stats)

