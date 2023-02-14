---
page_title: "materialize_schema Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
    A schema data source.
---

# materialize_schema (Data Source)

### Example Usage

```terraform
data "materialize_schema" "all" {}

data "materialize_schema" "materialize" {
  database_name = "materialize"
}
```

### Argument Reference

- `database_name` - (String) The name of the database to get the schema for. Defaults to `materialize`.

### Attributes Reference

- `id` - (String) The ID of the schema.
- `name` - (String) The name of the schema.
- `database_name` - (String) The name of the database the schema is in.
