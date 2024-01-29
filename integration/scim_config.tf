data "materialize_scim_groups" "all" {}

data "materialize_scim_configs" "all" {}

resource "materialize_scim_config" "example_scim_config" {
  connection_name = "example_connection"
  source          = "okta"
}
