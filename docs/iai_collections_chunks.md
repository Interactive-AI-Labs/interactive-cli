## iai collections chunks

Manage the chunks (rows) in a collection

### Synopsis

Upsert, inspect, and delete the chunks stored in a collection.

### Options

```
  -h, --help   help for chunks
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
* [iai collections chunks bulk-delete](iai_collections_chunks_bulk-delete.md)	 - Delete many chunks by ids, metadata filter, or all
* [iai collections chunks count](iai_collections_chunks_count.md)	 - Count chunks, optionally scoped by a metadata filter or id prefix
* [iai collections chunks delete](iai_collections_chunks_delete.md)	 - Delete a single chunk by id
* [iai collections chunks get](iai_collections_chunks_get.md)	 - Get a single chunk
* [iai collections chunks list](iai_collections_chunks_list.md)	 - List chunks (keyset-paginated)
* [iai collections chunks patch](iai_collections_chunks_patch.md)	 - Update a chunk's metadata and/or text from a file
* [iai collections chunks upsert](iai_collections_chunks_upsert.md)	 - Upsert chunks from a file

