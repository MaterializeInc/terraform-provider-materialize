# Grant role the privilege USAGE for objects in the database database
resource "materialize_database_grant_default_privilege" "example" {
  grantee_name     = "grantee"
  privilege        = "USAGE"
  target_role_name = "target_role"
}
