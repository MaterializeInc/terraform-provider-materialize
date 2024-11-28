resource "materialize_view" "simple_view" {
  name          = "simple_view"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  comment       = "view comment"

  statement = <<SQL
SELECT
    1 AS id
SQL

}

resource "materialize_view_grant" "database_view_select" {
  role_name     = materialize_role.role_1.name
  privilege     = "SELECT"
  view_name     = materialize_view.simple_view.name
  schema_name   = materialize_view.simple_view.schema_name
  database_name = materialize_view.simple_view.database_name
}

output "qualified_view" {
  value = materialize_view.simple_view.qualified_sql_name
}

data "materialize_view" "all" {}
