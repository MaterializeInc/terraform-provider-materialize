---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_source_webhook Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  A webhook source describes a webhook you want Materialize to read data from.
---

# materialize_source_webhook (Resource)

A webhook source describes a webhook you want Materialize to read data from.

## Example Usage

```terraform
resource "materialize_source_webhook" "example_webhook" {
  name            = "example_webhook"
  cluster_name    = materialize_cluster.example_cluster.name
  body_format     = "json"
  include_headers = false
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `body_format` (String) The body format of the webhook.
- `name` (String) The identifier for the source.

### Optional

- `check_expression` (String) The check expression for the webhook.
- `check_options` (List of String) The check options for the webhook.
- `cluster_name` (String) The cluster to maintain this source.
- `database_name` (String) The identifier for the source database. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `include_headers` (Boolean) Include headers in the webhook.
- `ownership_role` (String) The owernship role of the object.
- `schema_name` (String) The identifier for the source schema. Defaults to `public`.
- `size` (String) The size of the source.

### Read-Only

- `id` (String) The ID of this resource.
- `qualified_sql_name` (String) The fully qualified name of the source.
- `subsource` (List of Object) Subsources of a source. (see [below for nested schema](#nestedatt--subsource))

<a id="nestedatt--subsource"></a>
### Nested Schema for `subsource`

Read-Only:

- `database_name` (String)
- `name` (String)
- `schema_name` (String)

## Import

Import is supported using the following syntax:

```shell
# Sources can be imported using the source id:
terraform import materialize_source_webhook.example_source_webhook <source_id>
```