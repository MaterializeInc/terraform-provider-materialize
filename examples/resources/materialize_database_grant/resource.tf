# Grant USAGE to role example_role to database example_database
resource "materialize_database_grant" "database_grant_usage" {
  role_name     = "example_role"
  privilege     = "USAGE"
  database_name = "example_database"
}
