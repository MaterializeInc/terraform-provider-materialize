---
page_title: "materialize_source Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
    A source data source.
---

# materialize_source (Data Source)

### Example Usage

```terraform
data "materialize_source" "all" {}

data "materialize_source" "materialize" {
  database_name = "materialize"
}

data "materialize_source" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}
```

### Argument Reference

- `database_name` - (String) The name of the database to get the source for. Defaults to `materialize`.
- `schema_name` - (String) The name of the schema to get the source for. Defaults to `public`. Required if `database_name` is set.

### Attributes Reference

- `sources` - (List) A list of sources.

### Nested Schema for `sources`

- `id` - (String) The ID of the source.
- `name` - (String) The name of the source.
- `database_name` - (String) The name of the database the source is in.
- `schema_name` - (String) The name of the schema the source is in.
- `type` - (String) The type of the source.
- `size` - (String) The size of the source.
- `envelope_type` - (String) The envelope type of the source.
- `connection_name` - (String) The name of the connection the source is using.
- `cluster_name` - (String) The name of the cluster the source is using.
