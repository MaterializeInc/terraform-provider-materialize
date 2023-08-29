resource "materialize_schema" "schema" {
  name          = "example_schema"
  database_name = materialize_database.database.name
}

resource "materialize_schema_grant" "schema_grant_usage" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  database_name = materialize_schema.schema.database_name
  schema_name   = materialize_schema.schema.name
}

resource "materialize_schema_grant" "schema_grant_create" {
  role_name     = materialize_role.role_2.name
  privilege     = "CREATE"
  database_name = materialize_schema.schema.database_name
  schema_name   = materialize_schema.schema.name
}

resource "materialize_schema_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
}

output "qualified_schema" {
  value = materialize_schema.schema.qualified_sql_name
}

data "materialize_schema" "all" {}
