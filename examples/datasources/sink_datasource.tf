data "materialize_sink" "all" {}

data "materialize_sink" "materialize" {
  database_name = "materialize"
}

data "materialize_sink" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}