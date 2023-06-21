# Grant SELECT to role example_role to connection example_database.example_schema.example_materialized_view
resource "materialize_grant_materialized_view" "materialized_view_grant_select" {
  role_name              = "example_role"
  privilege              = "SELECT"
  materialized_view_name = "example_materialized_view"
  schema_name            = "example_schema"
  database_name          = "example_database"
}
