---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_system_parameter Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
  
---

# materialize_system_parameter (Data Source)



## Example Usage

```terraform
data "materialize_system_parameter" "all" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `name` (String) The name of the specific system parameter to fetch.
- `region` (String) The region in which the resource is located.

### Read-Only

- `id` (String) The ID of this resource.
- `parameters` (List of Object) (see [below for nested schema](#nestedatt--parameters))

<a id="nestedatt--parameters"></a>
### Nested Schema for `parameters`

Read-Only:

- `description` (String)
- `name` (String)
- `setting` (String)
