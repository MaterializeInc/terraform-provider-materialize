resource "materialize_materialized_view" "simple_materialized_view" {
  name          = "simple_materialized_view"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  select_stmt = <<SQL
SELECT
    1 AS id
SQL
}
