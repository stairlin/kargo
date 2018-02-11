# PostgreSQL plugin

The PostgreSQL plugin will backup a [Postgre](https://www.postgresql.org/) SQL database.

### Configuration:

```toml
[source.postgresql]
  "host" = "127.0.0.1"
  "port" = "5432"
  "user" = "postgres"
  "password" = ""
  "db" = ""
```

### Fields

  - host
  - port
  - user
  - password
  - db

### External dependencies

  - pg_dump
  - pg_restore