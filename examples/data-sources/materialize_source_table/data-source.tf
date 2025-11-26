data "materialize_source_table" "all" {}

data "materialize_source_table" "materialize" {
  database_name = "materialize"
}

data "materialize_source_table" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}
