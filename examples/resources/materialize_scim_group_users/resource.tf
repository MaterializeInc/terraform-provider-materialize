# Create a SCIM group users resource
resource "materialize_user" "example_users1" {
  email = "example-user1@example.com"
  roles = ["Member"]
}

resource "materialize_user" "example_users2" {
  email = "example-user2@example.com"
  roles = ["Member"]
}

resource "materialize_scim_group_users" "example_scim_group_users" {
  group_id = materialize_scim_group.example_scim_group.id
  users = [
    materialize_user.example_users1.id,
    materialize_user.example_users2.id,
  ]
}
