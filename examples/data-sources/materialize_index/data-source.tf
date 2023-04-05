data "materialize_index" "all" {}

data "materialize_index" "materialize" {
  database_name = "materialize"
}

data "materialize_index" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}