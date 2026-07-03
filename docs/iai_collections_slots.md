## iai collections slots

Manage a collection's vector slots and their indexes

### Synopsis

Add, reindex, vacuum, inspect, and remove a collection's vector slots.

A slot is a named vector space (a column) on a collection: if a collection is a
table and a chunk is a row, a slot is a vector column down every row. A
collection can have several — e.g. a dense slot for embeddings and a sparse slot
for keywords — and each chunk holds one vector per slot. The slot's index is
what makes searching that column fast.

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

* [iai collections](iai_collections.md)	 - Knowledge bases (searchable tables of chunks) inside a pgvector database
* [iai collections slots add](iai_collections_slots_add.md)	 - Add a vector slot
* [iai collections slots delete](iai_collections_slots_delete.md)	 - Delete a vector slot
* [iai collections slots progress](iai_collections_slots_progress.md)	 - Show a slot's index build progress
* [iai collections slots reindex](iai_collections_slots_reindex.md)	 - Rebuild a slot's index (online)
* [iai collections slots vacuum](iai_collections_slots_vacuum.md)	 - Vacuum a slot (reclaim space, refresh stats)

