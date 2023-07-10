resource "materialize_database" "database" {
  name = "example_database"
}

resource "materialize_database_grant" "database_grant_usage" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  database_name = materialize_database.database.name
}

resource "materialize_database_grant" "database_grant_create" {
  role_name     = materialize_role.role_2.name
  privilege     = "CREATE"
  database_name = materialize_database.database.name
}

data "materialize_database" "all" {}

data "materialize_current_database" "default" {}
