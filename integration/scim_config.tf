data "materialize_scim_groups" "all" {}

data "materialize_scim_configs" "all" {}

resource "materialize_scim_config" "example_scim_config" {
  connection_name = "example_connection"
  source          = "okta"
}

resource "materialize_scim_group" "scim_group_example" {
  name        = "scim_group_example"
  description = "scim_group_example"
}

resource "materialize_scim_group_roles" "scim_group_roles_example" {
  group_id = materialize_scim_group.scim_group_example.id
  roles    = ["Admin"]
}

resource "materialize_user" "example_users" {
  count = 5

  email = "test${count.index + 1}@example.com"
  roles = ["Member"]
}

resource "materialize_scim_group_users" "scim_group_users_example" {
  group_id = materialize_scim_group.scim_group_example.id
  users    = [for user in materialize_user.example_users : user.id]
}
