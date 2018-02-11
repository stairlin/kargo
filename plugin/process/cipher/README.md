# Cipher plugin

The Cipher plugin will encrypt and authenticate backups. It also provide a key rotation system to periodically change the encryption key used to protect data written to the storage.

Keys must encoded in base 64 and be 64 characters long after decoding. Kargo provides a tool to generate random cipher keys `kargo generate cipher`.

### Configuration:

```toml
[processors.cipher]
  default = 0
  keys = [
    "$CIPHER_KEY_1", "B64_KEY"
  ]
```

### Fields

- keys
- default (key index)
