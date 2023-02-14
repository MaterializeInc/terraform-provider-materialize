---
page_title: "materialize_sink Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
    A sink data source.
---

# materialize_sink (Data Source)

### Example Usage

```terraform
data "materialize_sink" "all" {}

data "materialize_sink" "materialize" {
  database_name = "materialize"
}

data "materialize_sink" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}
```

### Argument Reference

- `database_name` - (String) The name of the database to get the sink for. Defaults to `materialize`.
- `schema_name` - (String) The name of the schema to get the sink for. Defaults to `public`. Required if `database_name` is set.

### Attributes Reference

- `id` - (String) The ID of the sink.
- `name` - (String) The name of the sink.
- `database_name` - (String) The name of the database the sink is in.
- `schema_name` - (String) The name of the schema the sink is in.
- `type` - (String) The type of the sink.
- `size` - (String) The size of the sink.
- `envelope_type` - (String) The envelope type of the sink.
- `connection_name` - (String) The name of the connection the sink is using.
- `cluster_name` - (String) The name of the cluster the sink is using.
