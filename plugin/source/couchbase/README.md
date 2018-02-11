# Couchbase plugin

The Couchbase plugin will backup a [Couchbase](https://www.couchbase.com/) NoSQL database with the community backup tools.

By default, all nodes and all buckets will be included in the backup. Over the time, as the database grows, this could become problematic. In that case, it is possible to only backup a single node via the `single_node` flag.

This plugin does not support bucket-specific and incremental backups.

### Configuration:

```toml
[source.couchbase]
  "host" = "127.0.0.1"
  "port" = "8091"
  "user" = "backup"
  "password" = "$COUCHBASE_PWD"
  "single_node" = false
  "rehash" = true
```

### Fields

 - `host`
 - `port` - Couchbase uses 8091 by default
 - `user`
 - `password`
 - `single_node` - Backup a single node instead of the whole cluster
 - `rehash` - To restore the data to a cluster with a different operating system. This option rehashes the information and distribute the data to the appropriate node within the cluster.

### External dependencies

  - cbbackup
  - cbrestore
  - tar