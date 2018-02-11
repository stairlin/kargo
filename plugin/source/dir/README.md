# Directory plugin

The directory plugin will backup a directory on the filesystem.

It is worth noting that symlinks are not followed during the backup process.

### Configuration:

```toml
[source.dir]
  "path" = "/path/to/dir"
```

### Fields

 - path

### External dependencies

  - tar