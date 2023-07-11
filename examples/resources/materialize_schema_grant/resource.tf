# Grant USAGE to role example_role to schema example_database.example_schema
resource "materialize_schema_grant" "schema_grant_usage" {
  role_name     = "example_role"
  privilege     = "USAGE"
  database_name = "example_database"
  schema_name   = "example_schema"
}
