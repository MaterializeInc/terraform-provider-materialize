---
page_title: "materialize_cluster_replica Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
    A cluster replica data source.
---

# materialize_cluster_replica (Data Source)

## Example Usage

```terraform
data "materialize_cluster_replica" "all" {}
```

## Schema

### Attributes Reference

- `cluster_replicas` - (List) A list of cluster replicas.

### Nested Schema for `cluster_replicas`

- `availability_zone` - (String) The availability zone of the cluster replica.
- `id` - (String) The ID of the cluster replica.
- `cluster` - (String) The name of the cluster the replica belongs to.
- `name` - (String) The name of the cluster replica.
- `size` - (String) The size of the cluster replica.
