---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_app_password Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  
---

# materialize_app_password (Resource)



## Example Usage

```terraform
# Create a service user and app password
resource "materialize_role" "production_dashboard" {
  name = "svc_production_dashboard"
}
resource "materialize_app_password" "production_dashboard_app_password" {
  name  = "production_dashboard_app_password"
  type  = "service"
  user  = materialize_role.production_dashboard.name
  roles = ["Member"]
}
resource "materialize_database_grant" "database_grant_usage" {
  role_name     = materialize_role.production_dashboard.name
  privilege     = "USAGE"
  database_name = "production_analytics"
}

# Create a personal app password for the current user
resource "materialize_app_password" "example_app_password" {
  name = "example_app_password_name"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A human-readable name for the app password.

### Optional

- `roles` (List of String) The roles to assign to the app password. Allowed values are 'Member' and 'Admin'. Only valid with service-type app passwords.
- `type` (String) The type of the app password: personal or service.
- `user` (String) The user to associate with the app password. Only valid with service-type app passwords.

### Read-Only

- `created_at` (String) The time at which the app password was created.
- `id` (String) The ID of this resource.
- `password` (String, Sensitive) The value of the app password.
- `secret` (String, Sensitive)

## Import

Import is supported using the following syntax:

```shell
# App passwords can be imported using the app password id:
terraform import materialize_app_password.example_app_password <app_password_id>
```
