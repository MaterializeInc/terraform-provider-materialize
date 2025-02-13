---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_source_postgres Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  A Postgres source describes a PostgreSQL instance you want Materialize to read data from.
---

# materialize_source_postgres (Resource)

A Postgres source describes a PostgreSQL instance you want Materialize to read data from.

## Example Usage

```terraform
resource "materialize_source_postgres" "example_source_postgres" {
  name         = "source_postgres"
  schema_name  = "schema"
  cluster_name = "quickstart"
  publication  = "mz_source"

  postgres_connection {
    name = "pg_connection"
    # Optional parameters
    # database_name = "postgres"
    # schema_name = "public"
  }

  table {
    upstream_name        = "table1"
    upstream_schema_name = "schema1"
    name                 = "s1_table1"
  }

  table {
    upstream_name        = "table2"
    upstream_schema_name = "schema2"
    name                 = "s2_table2"
  }
}

# CREATE SOURCE schema.source_postgres
#   FROM POSTGRES CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source')
#   FOR TABLES (schema1.table1 AS s1_table1, schema2.table2 AS s2_table2);
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The identifier for the source.
- `postgres_connection` (Block List, Min: 1, Max: 1) The PostgreSQL connection to use in the source. (see [below for nested schema](#nestedblock--postgres_connection))
- `publication` (String) The PostgreSQL publication (the replication data set containing the tables to be streamed to Materialize).
- `table` (Block Set, Min: 1) Creates subsources for specific tables in the Postgres connection. (see [below for nested schema](#nestedblock--table))

### Optional

- `cluster_name` (String) The cluster to maintain this source.
- `comment` (String) Comment on an object in the database.
- `database_name` (String) The identifier for the source database in Materialize. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `expose_progress` (Block List, Max: 1) The name of the progress collection for the source. If this is not specified, the collection will be named `<src_name>_progress`. (see [below for nested schema](#nestedblock--expose_progress))
- `ownership_role` (String) The owernship role of the object.
- `region` (String) The region to use for the resource connection. If not set, the default region is used.
- `schema_name` (String) The identifier for the source schema in Materialize. Defaults to `public`.
- `text_columns` (List of String) Decode data as text for specific columns that contain PostgreSQL types that are unsupported in Materialize. Can only be updated in place when also updating a corresponding `table` attribute.

### Read-Only

- `id` (String) The ID of this resource.
- `qualified_sql_name` (String) The fully qualified name of the source.
- `size` (String) The size of the cluster maintaining this source.

<a id="nestedblock--postgres_connection"></a>
### Nested Schema for `postgres_connection`

Required:

- `name` (String) The postgres_connection name.

Optional:

- `database_name` (String) The postgres_connection database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The postgres_connection schema name. Defaults to `public`.


<a id="nestedblock--table"></a>
### Nested Schema for `table`

Required:

- `upstream_name` (String) The name of the table in the upstream Postgres database.

Optional:

- `database_name` (String) The database of the table in Materialize.
- `name` (String) The name of the table in Materialize.
- `schema_name` (String) The schema of the table in Materialize.
- `upstream_schema_name` (String) The schema of the table in the upstream Postgres database.


<a id="nestedblock--expose_progress"></a>
### Nested Schema for `expose_progress`

Required:

- `name` (String) The expose_progress name.

Optional:

- `database_name` (String) The expose_progress database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The expose_progress schema name. Defaults to `public`.

## Import

Import is supported using the following syntax:

```shell
# Sources can be imported using the source id:
terraform import materialize_source_postgres.example_source_postgres <region>:<source_id>

# Source id and information be found in the `mz_catalog.mz_sources` table
# The region is the region where the database is located (e.g. aws/us-east-1)
```
