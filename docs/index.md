---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize Provider"
subcategory: ""
description: |-
  
---

# materialize Provider



## Example Usage

```terraform
# Configuration-based authentication
provider "materialize" {
  host     = local.host
  username = local.username
  password = local.password
  port     = local.port
  database = local.database
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `database` (String) The Materialize database
- `host` (String) Materialize host
- `password` (String, Sensitive) Materialize host
- `port` (Number) The Materialize port number to connect to at the server host
- `testing` (Boolean) Enable to test the provider locally
- `username` (String) Materialize username
