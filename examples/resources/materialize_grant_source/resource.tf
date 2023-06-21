# Grant SELECT to role example_role to source example_database.example_schema.example_source
resource "materialize_grant_source" "source_grant_select" {
  role_name     = "example_role"
  privilege     = "SELECT"
  source_name   = "example_source"
  schema_name   = "example_schema"
  database_name = "example_database"
}
