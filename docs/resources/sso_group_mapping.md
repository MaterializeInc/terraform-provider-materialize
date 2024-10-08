---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_sso_group_mapping Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  The SSO group role mapping resource allows you to set the roles for an SSO group. This allows you to automatically assign additional roles according to your identity provider groups
---

# materialize_sso_group_mapping (Resource)

The SSO group role mapping resource allows you to set the roles for an SSO group. This allows you to automatically assign additional roles according to your identity provider groups

## Example Usage

```terraform
resource "materialize_sso_group_mapping" "example_sso_group_mapping" {
  group         = "admins"
  sso_config_id = materialize_sso_config.example_sso_config.id
  roles         = ["Admin"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `group` (String) The name of the SSO group.
- `roles` (Set of String) List of role names associated with the group.
- `sso_config_id` (String) The ID of the associated SSO configuration.

### Read-Only

- `enabled` (Boolean) Whether the group mapping is enabled.
- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# Retrieve the SSO Configuration ID and Group Mapping ID using the materialize_sso_config data source
# Example of using the data source in your configuration:
#
# data "materialize_sso_config" "all" {}
#
# output "sso_config_output" {
#     value = data.materialize_sso_config.all
# }
#
# The SSO Configuration ID and Group Mapping ID can be retrieved using the following command:
# terraform output sso_config_output

# Import command:
terraform import materialize_sso_group_mapping.example <sso_config_id>:<sso_group_mapping_id>

# Note: Replace <sso_config_id> with the actual ID of the SSO configuration
# and <sso_group_mapping_id> with the actual ID of the group mapping you want to import
```
