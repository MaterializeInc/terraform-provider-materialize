---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_view Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
  
---

# materialize_view (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `database_name` (String) Limit views to a specific database
- `schema_name` (String) Limit views to a specific schema within a specific database

### Read-Only

- `id` (String) The ID of this resource.
- `views` (List of Object) The views in the account (see [below for nested schema](#nestedatt--views))

<a id="nestedatt--views"></a>
### Nested Schema for `views`

Read-Only:

- `database_name` (String)
- `id` (String)
- `name` (String)
- `schema_name` (String)


