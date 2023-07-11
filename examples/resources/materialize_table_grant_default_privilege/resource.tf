# Grant role the privilege USAGE for objects in the schema database.schema
resource "materialize_table_grant_default_privilege" "example" {
  grantee_name     = "grantee"
  privilege        = "USAGE"
  target_role_name = "target_role"
  schema_name      = "schema"
  database_name    = "database"
}
