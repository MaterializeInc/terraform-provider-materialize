resource "materialize_view" "simple_view" {
  name          = "simple_view"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  select_stmt = <<SQL
SELECT
    1 AS id
SQL

}

output "qualified_view" {
  value = materialize_view.simple_view.qualified_sql_name
}