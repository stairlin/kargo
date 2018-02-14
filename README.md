# 🚢 Kargo - Backup & Restore [![CircleCI](https://circleci.com/gh/stairlin/kargo.svg?style=svg&circle-token=a0e2b657eb18a1c303535c8d122ba8a09d0a9f98)](https://circleci.com/gh/stairlin/kargo) [![Go Report Card](https://goreportcard.com/badge/github.com/stairlin/kargo)](https://goreportcard.com/report/github.com/stairlin/kargo)

Kargo is a CLI written in Go for transporting everything everywhere, which makes it the ideal companion to backup & restore data.

The design is unapologetically taken from Telegraf ❤️ with its plugin system. That allows developers to create adapters for any kind of databases, storages, or notifiers.

For example, you might want to backup a Couchbase database, compress and encrypt the backup, and store it on Amazon S3. You also may want to send a Pagerduty alert when a backup failed or timed out. Kargo can be seen as a database-agnostic backup tool.

Kargo is plugin-driven and has the concept of 4 distinct plugins:

    1. Source Plugins backup and restore a source, such as a database, a folder, or anything else.
    2. Processor Plugins encode/decode backup data (compress, cipher, ...).
    3. Store Plugins persist backups and can be queried.
    4. Notification Plugins send notifications on success/failure.

New plugins are designed to be easy to contribute, we'll eagerly accept pull requests and will manage the set of plugins that Kargo supports.

## Installation

You can download the binaries from the [release](https://github.com/stairlin/kargo/releases) section.

From Homebrew (macOS):

1. Run `brew tap stairlin/homebrew-tap`
2. Run `brew install stairlin/tap/kargo`

From Source:

Kargo is currently tested on go 1.9.

1. [Install Go](https://golang.org/doc/install)
2. [Setup your GOPATH](https://golang.org/doc/code.html#GOPATH)
3. Run `go get -d github.com/stairlin/kargo`
4. Run `cd $GOPATH/src/github.com/stairlin/kargo`
5. Run `go get -v -t -d ./...`
6. Run `go run main.go`

## Usage

Backup & restore data:

```shell
kargo backup
kargo restore my_backup_key
```

List backups:

```shell
kargo list
kargo list --from 2018-02-14
kargo list --prefix foo --to 2017-12-31
kargo list --pattern ^[a-z]{3}.*
kargo list --limit 50
```

Backup with a custom key name

```shell
kargo backup --key dat_snapshot
```

Pull a backup to local disk & restore data:

```shell
kargo pull my_backup_key
kargo restore --local my_backup_key
```

Help:

```shell
kargo help
```

## Configuration

### Environment variables

Environment variables can be used anywhere in the config file, simply prepend them with $. The variable must be within quotes (ie, "$STR_VAR"). Numbers, booleans, and other data types are not supported.

### Configuration file locations

The location of the configuration file can be set via the --config command line flag.

Order on which Kargo looks for a configuration file:

    1. --config flag
    1. kargo.toml in the working directory
    2. $KARGO_CONFIG environment variable
    4. default location `/etc/kargo/kargo.toml`

Show the current configuration:

```shell
kargo config
```

Example of a configuration file.

```toml
[agent]
  debug = true

[source.dir]
  path = "/oh/my/dir/"

[processors.gzip]

[processors.cipher]
  default = 0
  keys = ["my_base64_key", "another_key"]

[storage.s3]
  id = "foo"
  secret = "bar"
  token = ""
  bucket = "my_bucket"
  region = "eu-central-1"

[[notifiers.pagerduty]]
  key = "api_key"

[[notifiers.slack]]
  username = "My Project"
  url = "https://hooks.slack.com/services/foo"
```

## Plugins

### Sources

1. [Consul](./plugin/source/consul)
2. [Couchbase](./plugin/source/couchbase)
3. [Directory](./plugin/source/dir)
4. [InfluxDB](./plugin/source/influxdb)
5. [PostgreSQL](./plugin/source/postgresql)

### Storages

1. [Filesystem](./plugin/storage/fs)
2. [Amazon S3](./plugin/storage/s3)

### Processors

1. [Cipher](./plugin/process/cipher)
2. [GZip](./plugin/process/gzip)

### Notifiers

1. [Pagerduty](./plugin/notification/pagerduty)
2. [Slack](./plugin/notification/slack)

## Dependencies

Most dependencies are packed into the binary, such as AWS S3, gzip, Pagerduty, etc. However, `Source` plugins mainly rely on shell commands to work, so these dependencies must be installed separately and set to the $PATH.

## Internal design

Kargo piggyback on the powerful Go [I/O library](https://golang.org/pkg/io/) to keep the memory and disk footprint minimal. In most cases, data is being streamed from the source to the storage with no or minimal internal buffering. There are plugins, such as `cipher` that must work data by chunks for obvious reasons, so it will use
a small buffer.

It is worth noting that most `Source` plugins cannot stream data right away. Indeed, they have to create a temporary file that contains the backup before. `postgres` is currently the only plugin that can stream data directly, thanks to `pg_dump`.

## Contributing

This project is still unstable and not production ready. However, new plugins and bug fixes are welcome.

There is currently no guidelines, no red tape to contribute. Enjoy while it last. :)