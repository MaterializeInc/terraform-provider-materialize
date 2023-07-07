---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_grant_system Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  Manages the system privileges for roles.
---

# materialize_grant_system (Resource)

Manages the system privileges for roles.

## Example Usage

```terraform
# Grant role the privilege CREATEDB
resource "materialize_grant_system" "role_createdb" {
  role_name = "role"
  privilege = "CREATEDB"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `privilege` (String) The system privilege to grant.
- `role_name` (String) The name of the role to grant privilege to.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
#Grants can be imported using the concatenation of GRANT SYSTEM, the id of the role and the privilege 
terraform import materialize_grant_system.example GRANT SYSTEM|<role_id>|<privilege>
```