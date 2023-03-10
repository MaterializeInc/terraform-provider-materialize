resource "materialize_view" "simple_view" {
  name          = "simple_view"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  select_stmt = <<SQL
SELECT
    *
FROM
    ${materialize_table.simple_table.qualified_name}
SQL

  depends_on = [materialize_table.simple_table]
}

resource "materialize_view" "simple_view" {
  name          = "simple_view"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  select_stmt = "SELECT * FROM materialize.public.simple_table"
}
