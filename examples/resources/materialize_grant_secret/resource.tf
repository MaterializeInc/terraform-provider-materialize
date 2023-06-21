# Grant USAGE to role example_role to secret example_database.example_schema.example_secret
resource "materialize_grant_secret" "secret_grant_usage" {
  role_name     = "example_role"
  privilege     = "USAGE"
  secret_name   = "example_secret"
  schema_name   = "example_schema"
  database_name = "example_database"
}
