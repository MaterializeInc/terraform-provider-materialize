data "materialize_source" "all" {}

data "materialize_source" "materialize" {
  database_name = "materialize"
}

data "materialize_source" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}