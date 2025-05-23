---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_connection_postgres Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  A Postgres connection establishes a link to a single database of a PostgreSQL server.
---

# materialize_connection_postgres (Resource)

A Postgres connection establishes a link to a single database of a PostgreSQL server.

## Example Usage

```terraform
# Create a Postgres Connection
resource "materialize_connection_postgres" "example_postgres_connection" {
  name = "example_postgres_connection"
  host = "instance.foo000.us-west-1.rds.amazonaws.com"
  port = 5432
  user {
    secret {
      name          = "example"
      database_name = "database"
      schema_name   = "schema"
    }
  }
  password {
    name          = "example"
    database_name = "database"
    schema_name   = "schema"
  }
  database = "example"
}

# CREATE CONNECTION example_postgres_connection TO POSTGRES (
#     HOST 'instance.foo000.us-west-1.rds.amazonaws.com',
#     PORT 5432,
#     USER SECRET "database"."schema"."example"
#     PASSWORD SECRET "database"."schema"."example",
#     DATABASE 'example'
# );


# Create a Postgres Connection with SSH tunnel & plain text user
resource "materialize_connection_postgres" "example_postgres_connection" {
  name     = "example_postgres_connection"
  host     = "instance.foo000.us-west-1.rds.amazonaws.com"
  port     = 5432
  database = "example"

  user {
    text = "my_user"
  }
  password {
    name          = "example"
    database_name = "database"
    schema_name   = "schema"
  }
  ssh_tunnel {
    name = "example"
  }
}

# CREATE CONNECTION example_postgres_connection TO POSTGRES (
#     HOST 'instance.foo000.us-west-1.rds.amazonaws.com',
#     PORT 5432,
#     USER "my_user",
#     PASSWORD SECRET "database"."schema"."example",
#     DATABASE 'example',
#     SSH TUNNEL "example"
# );
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `database` (String) The target Postgres database.
- `host` (String) The Postgres database hostname.
- `name` (String) The identifier for the connection.
- `user` (Block List, Min: 1, Max: 1) The Postgres database username.. Can be supplied as either free text using `text` or reference to a secret object using `secret`. (see [below for nested schema](#nestedblock--user))

### Optional

- `aws_privatelink` (Block List, Max: 1) The AWS PrivateLink configuration for the Postgres database. (see [below for nested schema](#nestedblock--aws_privatelink))
- `comment` (String) Comment on an object in the database.
- `database_name` (String) The identifier for the connection database in Materialize. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `ownership_role` (String) The owernship role of the object.
- `password` (Block List, Max: 1) The Postgres database password. (see [below for nested schema](#nestedblock--password))
- `port` (Number) The Postgres database port.
- `region` (String) The region to use for the resource connection. If not set, the default region is used.
- `schema_name` (String) The identifier for the connection schema in Materialize. Defaults to `public`.
- `ssh_tunnel` (Block List, Max: 1) The SSH tunnel configuration for the Postgres database. (see [below for nested schema](#nestedblock--ssh_tunnel))
- `ssl_certificate` (Block List, Max: 1) The client certificate for the Postgres database.. Can be supplied as either free text using `text` or reference to a secret object using `secret`. (see [below for nested schema](#nestedblock--ssl_certificate))
- `ssl_certificate_authority` (Block List, Max: 1) The CA certificate for the Postgres database.. Can be supplied as either free text using `text` or reference to a secret object using `secret`. (see [below for nested schema](#nestedblock--ssl_certificate_authority))
- `ssl_key` (Block List, Max: 1) The client key for the Postgres database. (see [below for nested schema](#nestedblock--ssl_key))
- `ssl_mode` (String) The SSL mode for the Postgres database.
- `validate` (Boolean) If the connection should wait for validation.

### Read-Only

- `id` (String) The ID of this resource.
- `qualified_sql_name` (String) The fully qualified name of the connection.

<a id="nestedblock--user"></a>
### Nested Schema for `user`

Optional:

- `secret` (Block List, Max: 1) The `user` secret value. Conflicts with `text` within this block. (see [below for nested schema](#nestedblock--user--secret))
- `text` (String, Sensitive) The `user` text value. Conflicts with `secret` within this block

<a id="nestedblock--user--secret"></a>
### Nested Schema for `user.secret`

Required:

- `name` (String) The user name.

Optional:

- `database_name` (String) The user database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The user schema name. Defaults to `public`.



<a id="nestedblock--aws_privatelink"></a>
### Nested Schema for `aws_privatelink`

Required:

- `name` (String) The aws_privatelink name.

Optional:

- `database_name` (String) The aws_privatelink database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The aws_privatelink schema name. Defaults to `public`.


<a id="nestedblock--password"></a>
### Nested Schema for `password`

Required:

- `name` (String) The password name.

Optional:

- `database_name` (String) The password database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The password schema name. Defaults to `public`.


<a id="nestedblock--ssh_tunnel"></a>
### Nested Schema for `ssh_tunnel`

Required:

- `name` (String) The ssh_tunnel name.

Optional:

- `database_name` (String) The ssh_tunnel database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The ssh_tunnel schema name. Defaults to `public`.


<a id="nestedblock--ssl_certificate"></a>
### Nested Schema for `ssl_certificate`

Optional:

- `secret` (Block List, Max: 1) The `ssl_certificate` secret value. Conflicts with `text` within this block. (see [below for nested schema](#nestedblock--ssl_certificate--secret))
- `text` (String, Sensitive) The `ssl_certificate` text value. Conflicts with `secret` within this block

<a id="nestedblock--ssl_certificate--secret"></a>
### Nested Schema for `ssl_certificate.secret`

Required:

- `name` (String) The ssl_certificate name.

Optional:

- `database_name` (String) The ssl_certificate database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The ssl_certificate schema name. Defaults to `public`.



<a id="nestedblock--ssl_certificate_authority"></a>
### Nested Schema for `ssl_certificate_authority`

Optional:

- `secret` (Block List, Max: 1) The `ssl_certificate_authority` secret value. Conflicts with `text` within this block. (see [below for nested schema](#nestedblock--ssl_certificate_authority--secret))
- `text` (String, Sensitive) The `ssl_certificate_authority` text value. Conflicts with `secret` within this block

<a id="nestedblock--ssl_certificate_authority--secret"></a>
### Nested Schema for `ssl_certificate_authority.secret`

Required:

- `name` (String) The ssl_certificate_authority name.

Optional:

- `database_name` (String) The ssl_certificate_authority database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The ssl_certificate_authority schema name. Defaults to `public`.



<a id="nestedblock--ssl_key"></a>
### Nested Schema for `ssl_key`

Required:

- `name` (String) The ssl_key name.

Optional:

- `database_name` (String) The ssl_key database name. Defaults to `MZ_DATABASE` environment variable if set or `materialize` if environment variable is not set.
- `schema_name` (String) The ssl_key schema name. Defaults to `public`.

## Import

Import is supported using the following syntax:

```shell
# Connections can be imported using the connection id:
terraform import materialize_connection_postgres.example <region>:<connection_id>

# Connection id and information be found in the `mz_catalog.mz_connections` table
# The region is the region where the database is located (e.g. aws/us-east-1)
```
