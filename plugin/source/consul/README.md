# Consul plugin

The Consul plugin will save snapshots of a [Hashicorp Consul](https://www.consul.io/) server state.

The backup command will retrieve an atomic, point-in-time snapshot of the state of the Consul servers which includes key/value entries, service catalog, prepared queries, sessions, and ACLs.

If ACLs are enabled, a management token must be supplied in order to perform a backup.

### Configuration:

```toml
[source.consul]
  "http_addr" = "http://127.0.0.1:8500"
  "datacenter" = "us-east-1"
  "token" = "$CONSUL_TOKEN"
  "stale" = false
```

### Fields

 - `ca_file` - Path to a CA file to use for TLS when communicating with Consul.
 - `ca_path` - Path to a directory of CA certificates to use for TLS when communicating with Consul.
 - `client_cert` - Path to a client cert file to use for TLS when verify_incoming is enabled.
 - `client_key` - Path to a client key file to use for TLS when verify_incoming is enabled.
 - `tls_server_name` - The server name to use as the SNI host when connecting via TLS.
 - `token` - ACL token to use in the request.
 - `datacenter` - Name of the datacenter to query. If unspecified, the query will default to the datacenter of the Consul agent at the HTTP address.
 - `stale` - Permit any Consul server (non-leader) to respond to this request. This allows for lower latency and higher throughput, but can result in stale data. This option has no effect on non-read operations. The default value is false.
 - `http_addr` - Address of the Consul agent with the port. This can be an IP address or DNS address, but it must include the port. This can also be specified via the CONSUL_HTTP_ADDR environment variable. In Consul 0.8 and later, the default value is http://127.0.0.1:8500, and https can optionally be used instead.

### External dependencies

  - consul