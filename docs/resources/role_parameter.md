---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_role_parameter Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  Manages a system parameter in Materialize.
---

# materialize_role_parameter (Resource)

Manages a system parameter in Materialize.

## Example Usage

```terraform
# ALTER ROLE some_role SET transaction_isolation = 'strict serializable';
resource "materialize_role" "example_role" {
  name = "some_role"
}

resource "materialize_role_parameter" "example_role_parameter" {
  role_name      = materialize_role.example_role.name
  variable_name  = "transaction_isolation"
  variable_value = "strict serializable"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `role_name` (String) The name of the role to grant privilege to.
- `variable_name` (String) The name of the session variable to modify.
- `variable_value` (String) The value to assign to the session variable.

### Optional

- `region` (String) The region to use for the resource connection. If not set, the default region is used.

### Read-Only

- `id` (String) The ID of this resource.
