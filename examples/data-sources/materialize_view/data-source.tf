data "materialize_view" "all" {}

data "materialize_view" "materialize" {
  database_name = "materialize"
}

data "materialize_view" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}