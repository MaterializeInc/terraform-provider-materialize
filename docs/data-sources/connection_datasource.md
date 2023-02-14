---
page_title: "materialize_connection Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
    A connection data source.
---

# materialize_connection (Data Source)

### Example Usage

```terraform
data "materialize_connection" "all" {}

data "materialize_connection" "materialize" {
  database_name = "materialize"
}

data "materialize_connection" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}
```

### Argument Reference

- `database_name` - (String) The name of the database to get the connection for. Defaults to `materialize`.
- `schema_name` - (String) The name of the schema to get the connection for. Defaults to `public`. Required if `database_name` is set.

### Attributes Reference

- `id` - (String) The ID of the connection.
- `name` - (String) The name of the connection.
- `database_name` - (String) The name of the database the connection is in.
- `schema_name` - (String) The name of the schema the connection is in.
- `type` - (String) The type of the connection.
