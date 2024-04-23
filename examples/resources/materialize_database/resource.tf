# Create a Materialize database without a public schema
resource "materialize_database" "example" {
  name = "example"
}

# By default, Materialize creates a public schema in each database
# The Terraform provider on the other hand does not create a public schema by default
# Optionally you can create a public schema in the database using the materialize_schema resource
resource "materialize_schema" "public" {
  name     = "public"
  database_name = materialize_database.example.name
}

# Grant USAGE to the PUBLIC pseudo-role for the public schema
# This matches the default behavior of Materialize
resource "materialize_schema_grant" "schema_grant_usage" {
  role_name     = "PUBLIC"
  privilege     = "USAGE"
  database_name = materialize_database.example.name
  schema_name   = materialize_schema.public.name
}
