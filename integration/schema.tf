resource "materialize_schema" "schema" {
  name          = "example"
  database_name = materialize_database.database.name
}