---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "materialize_network_policy Data Source - terraform-provider-materialize"
subcategory: ""
description: |-
  A network policy data source. This can be used to get information about all network policies in Materialize.
---

# materialize_network_policy (Data Source)

A network policy data source. This can be used to get information about all network policies in Materialize.

## Example Usage

```terraform
data "materialize_network_policy" "all" {}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `region` (String) The region in which the resource is located.

### Read-Only

- `id` (String) The ID of this resource.
- `network_policies` (List of Object) The network policies in the account (see [below for nested schema](#nestedatt--network_policies))

<a id="nestedatt--network_policies"></a>
### Nested Schema for `network_policies`

Read-Only:

- `comment` (String)
- `id` (String)
- `name` (String)
- `rules` (List of Object) (see [below for nested schema](#nestedobjatt--network_policies--rules))

<a id="nestedobjatt--network_policies--rules"></a>
### Nested Schema for `network_policies.rules`

Read-Only:

- `action` (String)
- `address` (String)
- `direction` (String)
- `name` (String)