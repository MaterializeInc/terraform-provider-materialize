# Grant USAGE to role example_role to table example_database.example_schema.example_table
resource "materialize_table_grant" "table_grant_usage" {
  role_name     = "example_role"
  privilege     = "USAGE"
  database_name = "example_database"
  schema_name   = "example_schema"
  table_name    = "example_table"
}
