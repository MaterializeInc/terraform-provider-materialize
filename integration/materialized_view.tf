resource "materialize_materialized_view" "simple_materialized_view" {
  name          = "simple_materialized_view"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  statement = <<SQL
SELECT
    1 AS id
SQL
}

output "qualified_materialized_view" {
  value = materialize_materialized_view.simple_materialized_view.qualified_sql_name
}

data "materialize_materialized_view" "all" {}