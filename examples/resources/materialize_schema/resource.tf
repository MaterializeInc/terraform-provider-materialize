resource "materialize_schema" "example_schema" {
  name          = "schema"
  database_name = "database"
}

resource "materialize_schema" "example_by_name" {
  name             = "schema"
  database_name    = "database"
  identify_by_name = true # Set to true to use the schema name as the resource ID
}