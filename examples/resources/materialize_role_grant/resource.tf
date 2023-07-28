# Grant role to user
resource "materialize_role_grant" "role_grant_user" {
  role_name   = "role"
  member_name = "user"
}
