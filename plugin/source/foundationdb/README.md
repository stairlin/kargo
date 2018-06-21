# FoundationDB plugin

The Foundation plugin will backup a [FoundationDB](https://www.foundationdb.org/) multi-model data store.

FoundationDB’s backup tool makes a consistent, point-in-time backup of a FoundationDB database without downtime. Like FoundationDB itself, the backup/restore software is distributed, with multiple backup agents cooperating to perform a backup or restore faster than a single machine can send or receive data and to continue the backup process seamlessly even when some backup agents fail.

The FoundationDB database usually cannot maintain a consistent snapshot long enough to read the entire database, so full backup consists of an inconsistent copy of the data with a log of database changes that took place during the creation of that inconsistent copy. During a restore, the inconsistent copy and the log of changes are combined to reconstruct a consistent, point-in-time snapshot of the original database.

 > Warning: It is your responsibility to ensure that no clients are accessing the database while it is being restored. During the restore process the database is in an inconsistent state, and writes that happen during the restore process might be partially or completely overwritten by restored data.

### Configuration:

```toml
[source.foundationdb]
  "cluster" = "/etc/foundationdb/fdb.cluster"
  "tag" = "kargo"
```

### Fields

  - `cluster` - Path to the cluster file
  - `tag` - A “tag” is a named slot in which a backup task executes. Backups on different named tags make progress and are controlled independently, though their executions are handled by the same set of backup agent processes. Any number of unique backup tags can be active at once. It the tag is not specified, the default tag name “default” is used.
  - `fdbbackup` - When this value is left empty, the value will be taken from the $PATH.
  - `fdbrestore` - When this value is left empty, the value will be taken from the $PATH.
  - `backup_agent` - When this value is left empty, the value will be taken from the $PATH.

### External dependencies

  - tar
  - fdbbackup
  - fdbrestore
  - backup_agent

### [Disaster Recovery](https://apple.github.io/foundationdb/backups.html#backup-vs-dr)

Backing up one database to another is a special form of backup is called DR backup or just DR for short. DR stands for Disaster Recovery, as it can be used to keep two geographically separated databases in close synchronization to recover from a catastrophic disaster.

This plugin does not support DR.
