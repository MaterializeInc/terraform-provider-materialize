data "materialize_secret" "all" {}

data "materialize_secret" "materialize" {
  database_name = "materialize"
}

data "materialize_secret" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}