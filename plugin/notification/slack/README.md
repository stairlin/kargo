# Slack plugin

The Slack plugin will send notifications to a slack channel via the webook
API to notify upon success/failure of a backup or restore operation.

### Configuration:

```toml
[[notifiers.slack]]
  username = "Kargo"
  url = "https://hooks.slack.com/services/uuid"
```

### Fields

- url
- channel
- username
