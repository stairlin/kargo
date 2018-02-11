# InfluxDB plugin

The InfluxDB plugin will backup an [InfluxDB](https://www.influxdata.com/) time-series database.

InfluxDB has the ability to snapshot an instance at a point-in-time and restore it. All backups are full backups. InfluxDB does not yet support incremental backups. There are two types of data to backup, the metastore and the metrics themselves. The metastore is backed up in its entirety. The metrics are backed up per-database in a separate operation from the metastore backup.

InfluxDB’s metastore contains internal information about the status of the system, including user information, database/shard metadata, CQs, RPs, and subscriptions.

 > Note: Restoring from backup is only supported while the InfluxDB daemon is stopped.

If the deamon is running, the restore will quietly fail.

### Configuration:

```toml
[source.influxdb]
  "host" = "127.0.0.1"
  "port" = "8088"
  "user" = ""
  "password" = ""
  "db" = ""
  "metadir" = "/var/lib/influxdb/meta"
  "datadir" = "/var/lib/influxdb/data"
```

### Fields

  - host
  - port
  - user
  - password
  - `db` - This is the database that you would like to restore the data to. This option is required if no `metadir` option is provided.
  - `metadir` - This is the path to the meta directory where you would like the metastore backup recovered to. For packaged installations, this should be specified as `/var/lib/influxdb/meta`.
  - `datadir` - This is the path to the data directory where you would like the database backup recovered to. For packaged installations, this should be specified as `/var/lib/influxdb/data`.

### External dependencies

  - tar
  - influxd

### Restore prior to v1.4

Prior to InfluxDB 1.4, it was not possible to restore a remote node. So kargo should run from the same instance.

### InfluxDB OSS and InfluxEnterprise

Backups are not interchangeable between InfluxDB OSS and InfluxEnterprise. You cannot restore an OSS backup to an InfluxEnterprise data node, nor can you restore an InfluxEnterprise backup to an OSS instance.