## iai collections

Vector collections (knowledge bases) inside a pgvector database

### Synopsis

Manage vector collections within a database.

A collection is a vector store (knowledge base) that lives inside an existing
pgvector database, so every command requires --database. Use 'iai databases
create' first to provision the database.

### Options

```
  -h, --help   help for collections
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
* [iai collections chunks](iai_collections_chunks.md)	 - Manage the chunks (rows) in a collection
* [iai collections create](iai_collections_create.md)	 - Create a collection from a config file
* [iai collections delete](iai_collections_delete.md)	 - Delete a collection and all its data
* [iai collections describe](iai_collections_describe.md)	 - Describe a collection's configuration
* [iai collections documents](iai_collections_documents.md)	 - Inspect documents (chunks grouped by documentId)
* [iai collections list](iai_collections_list.md)	 - List collections in a database
* [iai collections patch](iai_collections_patch.md)	 - Update a collection's mutable config from a file
* [iai collections search](iai_collections_search.md)	 - Search a collection (single-lane vector search)
* [iai collections slots](iai_collections_slots.md)	 - Manage a collection's vector slots and their indexes
* [iai collections stats](iai_collections_stats.md)	 - Show a collection's chunk count, size, and index status

