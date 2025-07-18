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
  name             = "example_webhook"
  cluster_name     = materialize_cluster.cluster.name
  body_format      = "json"
  check_expression = "headers->'x-mz-api-key' = secret"
  include_headers {
    not = ["x-mz-api-key"]
  }

  check_options {
    field {
      headers = true
    }
  }

  check_options {
    field {
      secret {
        name          = materialize_secret.password.name
        database_name = materialize_secret.password.database_name
        schema_name   = materialize_secret.password.schema_name
      }
    }
    alias = "secret"
  }
}

# CREATE SOURCE example_webhook IN CLUSTER cluster FROM WEBHOOK
#   BODY FORMAT json
#   INCLUDE HEADERS ( NOT 'x-mz-api-key' )
#   CHECK (
#     WITH ( HEADERS, SECRET materialize.public.password AS secret)
#     headers->'x-mz-api-key' = secret
#   );
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `body_format` (String) The body format of the webhook.
- `name` (String) The identifier for the source.

### Optional

- `check_expression` (String) The check expression for the webhook.
- `check_options` (Block List) The check options for the webhook. (see [below for nested schema](#nestedblock--check_options))
- `cluster_name` (String) The cluster to maintain this source.
- `comment` (String) Comment on an object in the database.
- `database_name` (String) The identifier for the source database in Materialize. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `include_header` (Block List) Map a header value from a request into a column. (see [below for nested schema](#nestedblock--include_header))
- `include_headers` (Block List, Max: 1) Include headers in the webhook. (see [below for nested schema](#nestedblock--include_headers))
- `ownership_role` (String) The owernship role of the object.
- `region` (String) The region to use for the resource connection. If not set, the default region is used.
- `schema_name` (String) The identifier for the source schema in Materialize. Defaults to `public`.

### Read-Only

- `id` (String) The ID of this resource.
- `qualified_sql_name` (String) The fully qualified name of the source.
- `size` (String) The size of the cluster maintaining this source.
- `url` (String) The webhook URL that can be used to send data to this source.

<a id="nestedblock--check_options"></a>
### Nested Schema for `check_options`

Required:

- `field` (Block List, Min: 1, Max: 1) The field for the check options. (see [below for nested schema](#nestedblock--check_options--field))

Optional:

- `alias` (String) The alias for the check options.
- `bytes` (Boolean) Change type to `bytea`.

<a id="nestedblock--check_options--field"></a>
### Nested Schema for `check_options.field`

Optional:

- `body` (Boolean) The body for the check options.
- `headers` (Boolean) The headers for the check options.
- `secret` (Block List, Max: 1) The secret for the check options. (see [below for nested schema](#nestedblock--check_options--field--secret))

<a id="nestedblock--check_options--field--secret"></a>
### Nested Schema for `check_options.field.secret`

Required:

- `name` (String) The secret name.

Optional:

- `database_name` (String) The secret database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The secret schema name. Defaults to `public`.




<a id="nestedblock--include_header"></a>
### Nested Schema for `include_header`

Required:

- `header` (String) The name for the header.

Optional:

- `alias` (String) The alias for the header.
- `bytes` (Boolean) Change type to `bytea`.


<a id="nestedblock--include_headers"></a>
### Nested Schema for `include_headers`

Optional:

- `all` (Boolean) Include all headers.
- `not` (List of String) Headers that should be excluded.
- `only` (List of String) Headers that should be included.

## Import

Import is supported using the following syntax:

```shell
# Sources can be imported using the source id:
terraform import materialize_source_webhook.example_source_webhook <region>:<source_id>

# Source id and information be found in the `mz_catalog.mz_sources` table
# The region is the region where the database is located (e.g. aws/us-east-1)
```
