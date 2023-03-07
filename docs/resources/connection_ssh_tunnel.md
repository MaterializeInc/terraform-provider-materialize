---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_connection_ssh_tunnel Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  The connection resource allows you to manage connections in Materialize.
---

# materialize_connection_ssh_tunnel (Resource)

The connection resource allows you to manage connections in Materialize.

## Example Usage

```terraform
# Create SSH Connection
resource "materialize_connection_ssh_tunnel" "example_ssh_connection" {
  name        = "ssh_example_connection"
  schema_name = "public"
  host        = "example.com"
  port        = 22
  user        = "example"
}

# CREATE CONNECTION ssh_example_connection TO SSH TUNNEL (
#    HOST 'example.com',
#    PORT 22,
#    USER 'example'
# );
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `host` (String) The host of the SSH tunnel.
- `name` (String) The name of the connection.
- `port` (Number) The port of the SSH tunnel.
- `user` (String) The user of the SSH tunnel.

### Optional

- `database_name` (String) The identifier for the connection database.
- `schema_name` (String) The identifier for the connection schema.

### Read-Only

- `id` (String) The ID of this resource.
- `qualified_name` (String) The fully qualified name of the connection.

## Import

Import is supported using the following syntax:

```shell
#Connections can be imported using the connection id:
terraform import materialize_connection_ssh_tunnel.example <connection_id>
```