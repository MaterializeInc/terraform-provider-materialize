---
page_title: "materialize_secret Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
    A secret data source.
---

# materialize_secret (Data Source)

### Example Usage

```terraform
data "materialize_secret" "all" {}

data "materialize_secret" "materialize" {
  database_name = "materialize"
}

data "materialize_secret" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}
```

### Argument Reference

- `database_name` - (String) The name of the database to get the secret for. Defaults to `materialize`.
- `schema_name` - (String) The name of the schema to get the secret for. Defaults to `public`. Required if `database_name` is set.

### Attributes Reference

- `id` - (String) The ID of the secret.
- `name` - (String) The name of the secret.
- `database_name` - (String) The name of the database the secret is in.
- `schema_name` - (String) The name of the schema the secret is in.
