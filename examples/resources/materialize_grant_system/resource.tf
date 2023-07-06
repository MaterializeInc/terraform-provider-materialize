# Grant role the privilege CREATEDB
resource "materialize_grant_system" "role_createdb" {
  role_name = "role"
  privilege = "CREATEDB"
}
