# Grant role the privilege CREATEDB
resource "materialize_grant_system_privilege" "role_createdb" {
  role_name = "role"
  privilege = "CREATEDB"
}

# Grant role the privilege CREATENETWORKPOLICY to allow the role to create network policies
resource "materialize_grant_system_privilege" "role_createnetworkpolicy" {
  role_name = "role"
  privilege = "CREATENETWORKPOLICY"
}
