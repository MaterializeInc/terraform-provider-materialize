# Grant SELECT to role example_role to view example_database.example_schema.example_view
resource "materialize_grant_view" "view_grant_select" {
  role_name     = "example_role"
  privilege     = "SELECT"
  database_name = "example_database"
  schema_name   = "example_schema"
  view_name     = "example_view"
}
