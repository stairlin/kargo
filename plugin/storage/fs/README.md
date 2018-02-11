# Filsystem plugin

The filesystem plugin will use the local file system to persist backups.

Kargo does not provide a mechanism to track backup ages and initiate a bulk deletion process. Therefore it is advised to set up a cron task to clean old backups from time to time.

### Configuration:

```toml
[processors.fs]
  path = "/path/to/backups"
```

### Fields

- path
