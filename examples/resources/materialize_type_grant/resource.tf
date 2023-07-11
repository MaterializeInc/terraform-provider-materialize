# Grant USAGE to role example_role to type example_database.example_schema.example_type
resource "materialize_type_grant" "type_grant_usage" {
  role_name     = "example_role"
  privilege     = "USAGE"
  database_name = "example_database"
  schema_name   = "example_schema"
  type_name     = "example_type"
}
