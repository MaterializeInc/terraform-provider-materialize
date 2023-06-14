resource "materialize_database" "database" {
  name = "example_database"
}

data "materialize_database" "all" {}

data "materialize_current_database" "default" {}
