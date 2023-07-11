# Grant role the privilege CREATEDB
resource "materialize_grant_system_privilege" "role_createdb" {
  role_name = "role"
  privilege = "CREATEDB"
}
