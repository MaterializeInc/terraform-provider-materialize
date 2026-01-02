data "materialize_table" "all" {}

data "materialize_table" "materialize" {
  database_name = "materialize"
}

data "materialize_table" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}
