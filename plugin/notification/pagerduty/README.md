# Pagerduty plugin

The Pagerduty plugin will trigger Pagerduty alerts when a backup or restore
operation failed.

The plugin needs an [API token](https://support.pagerduty.com/docs/using-the-api#section-generating-an-api-key). It currently uses the Event V2 API.

### Configuration:

```toml
[[notifiers.pagerduty]]
  key = "my_key"
```

### Fields

- key
