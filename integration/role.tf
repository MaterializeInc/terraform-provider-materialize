resource "materialize_role" "role_1" {
  name = "role_1"
}

resource "materialize_role" "role_2" {
  name = "role_2"
}

resource "materialize_grant_default_privilege" "test_select" {
  grantee_name     = materialize_role.role_1.name
  object_type      = "TABLE"
  privilege        = "SELECT"
  target_role_name = materialize_role.role_2.name
}

resource "materialize_grant_default_privilege" "test_insert" {
  grantee_name     = materialize_role.role_1.name
  object_type      = "TABLE"
  privilege        = "INSERT"
  target_role_name = materialize_role.role_2.name
}

resource "materialize_grant_default_privilege" "test_schema_database" {
  grantee_name     = materialize_role.role_1.name
  object_type      = "TABLE"
  privilege        = "UPDATE"
  target_role_name = materialize_role.role_2.name
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
}

output "qualified_role" {
  value = materialize_role.role_1.qualified_sql_name
}

data "materialize_role" "all" {}
