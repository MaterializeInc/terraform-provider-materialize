# Create a SCIM group role
resource "materialize_scim_group" "scim_group_example" {
  name        = "scim_group_example"
  description = "scim_group_example"
}

resource "materialize_scim_group_roles" "scim_group_roles_example" {
  group_id = materialize_scim_group.scim_group_example.id
  roles    = ["Admin", "Member"]
}
