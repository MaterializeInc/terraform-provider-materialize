---
page_title: "Provider: Materialize"
description: "Manage Materialize resources with Terraform"
---

# Materialize Provider

This is a terraform provider plugin for managing [Materialize](https://materialize.com/) resources.

## Provider Configuration

```terraform
provider "materialize" {
  host     = local.host
  username = local.username
  password = local.password
  port     = local.port
  database = local.database
}
```

## Schema

### Required

- `host` - (String) The host of the Materialize instance.
- `username` - (String) The username to connect to the Materialize instance.
- `password` - (String) The app password to connect to the Materialize instance.
- `port` - (Int) The port to connect to the Materialize instance.
- `database` - (String) The database to connect to the Materialize instance.
