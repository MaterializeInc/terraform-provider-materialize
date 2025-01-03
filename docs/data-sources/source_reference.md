---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_source_reference Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
  The materialize_source_reference data source retrieves a list of available upstream references for a given Materialize source. These references represent potential tables that can be created based on the source, but they do not necessarily indicate references the source is already ingesting. This allows users to see all upstream data that could be materialized into tables.
---

# materialize_source_reference (Data Source)

The `materialize_source_reference` data source retrieves a list of *available* upstream references for a given Materialize source. These references represent potential tables that can be created based on the source, but they do not necessarily indicate references the source is already ingesting. This allows users to see all upstream data that could be materialized into tables.

## Example Usage

```terraform
data "materialize_source_reference" "source_references" {
  source_id = materialize_source_mysql.test.id
}

output "source_references" {
  value = data.materialize_source_reference.my_source_references.references
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `source_id` (String) The ID of the source to get references for

### Optional

- `region` (String) The region in which the resource is located.

### Read-Only

- `id` (String) The ID of this resource.
- `references` (List of Object) The source references (see [below for nested schema](#nestedatt--references))

<a id="nestedatt--references"></a>
### Nested Schema for `references`

Read-Only:

- `columns` (List of String)
- `name` (String)
- `namespace` (String)
- `source_database_name` (String)
- `source_name` (String)
- `source_schema_name` (String)
- `source_type` (String)
- `updated_at` (String)