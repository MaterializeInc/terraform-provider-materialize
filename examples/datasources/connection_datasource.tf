data "materialize_connection" "all" {}

data "materialize_connection" "materialize" {
  database_name = "materialize"
}

data "materialize_connection" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}