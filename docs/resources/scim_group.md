---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_scim_group Resource - terraform-provider-materialize"
subcategory: ""
description: |-
  The SCIM group resource allows you to manage user groups in Frontegg.
---

# materialize_scim_group (Resource)

The SCIM group resource allows you to manage user groups in Frontegg.

## Example Usage

```terraform
# Create a SCIM group
resource "materialize_scim_group" "example_scim_group" {
  name        = "example-scim-group"
  description = "Example SCIM group"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the SCIM group.

### Optional

- `description` (String) A description of the SCIM group.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
# SCIM Group ID can be found using the `materialize_scim_groups` data source
terraform import materialize_scim_group.example_scim_group <scim_group_id>
```