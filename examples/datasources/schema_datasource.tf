data "materialize_schema" "all" {}

data "materialize_schema" "materialize" {
  database_name = "materialize"
}

