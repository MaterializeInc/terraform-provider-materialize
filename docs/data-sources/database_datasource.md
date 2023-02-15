---
page_title: "materialize_database Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
    A database data source.
---

# materialize_database (Data Source)

### Example Usage

```terraform
data "materialize_database" "all" {}
```

### Attributes Reference

- `databases` - (List) A list of databases.

### Nested Schema for `databases`

- `id` - (String) The ID of the database.
- `name` - (String) The name of the database.
