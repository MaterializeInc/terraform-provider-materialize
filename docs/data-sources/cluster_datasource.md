---
page_title: "materialize_cluster Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
    A cluster data source.
---

# materialize_cluster (Data Source)

## Example Usage

```terraform
data "materialize_cluster" "all" {}
```

## Schema

### Attributes Reference

- `clusters` - (List) A list of clusters.

### Nested Schema for `clusters`

- `id` - (String) The ID of the cluster.
- `name` - (String) The name of the cluster.
