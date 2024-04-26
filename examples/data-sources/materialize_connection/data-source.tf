data "materialize_connection" "all" {}

data "materialize_connection" "materialize" {
  database_name = "materialize"
}

data "materialize_connection" "materialize_schema" {
  database_name = "materialize"
  schema_name   = "schema"
}

data "materialize_connection" "by_id" {
  connection_id = "u1234" # The ID of the connection
}
