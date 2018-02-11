# Amazon S3 plugin

The S3 plugin will use [Amazon S3](https://aws.amazon.com/s3/) to persist backups.

Kargo does not provide a mechanism to track backup ages and initiate a bulk deletion process. Therefore it is advised to set up a lifecycle policy on the S3 bucket to clean old backups from time to time.

### Configuration:

```toml
[processors.s3]
  id = "<your_id>"
  secret = "<your_secret>"
  token = "<your_token>"
  folder = ""
  region = "eu-central-1"
  bucket = "db-backups"
  debug = false
```

### Fields

 - id
 - secret
 - token (optional)
 - folder (optional)
 - region
 - bucket
 - debug