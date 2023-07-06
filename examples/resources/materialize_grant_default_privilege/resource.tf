# Grant role the privilege UPDATE for tables in the schema database.schema
resource "materialize_grant_default_privilege" "test_schema_database" {
  grantee_name     = "grantee"
  object_type      = "TABLE"
  privilege        = "UPDATE"
  target_role_name = "target_role"
  schema_name      = "schema"
  database_name    = "database"
}
