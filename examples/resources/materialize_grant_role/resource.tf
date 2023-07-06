# Grant role to user
resource "materialize_grant_role" "role_grant_user" {
  role_name   = "role"
  member_name = "user"
}
