# Grant USAGE to role example_role to connection example_database.example_schema.example_connection
resource "materialize_grant_connection" "connection_grant_usage" {
  role_name       = "example_role"
  privilege       = "USAGE"
  connection_name = "example_connection"
  schema_name     = "example_schema"
  database_name   = "example_database"
}
