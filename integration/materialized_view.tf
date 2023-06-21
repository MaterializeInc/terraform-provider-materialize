resource "materialize_materialized_view" "simple_materialized_view" {
  name          = "simple_materialized_view"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  cluster_name  = "default"

  statement = <<SQL
SELECT
    1 AS id
SQL
}

resource "materialize_grant_materialized_view" "materialized_view_grant_select" {
  role_name              = materialize_role.role_1.name
  privilege              = "SELECT"
  materialized_view_name = materialize_materialized_view.simple_materialized_view.name
  schema_name            = materialize_materialized_view.simple_materialized_view.schema_name
  database_name          = materialize_materialized_view.simple_materialized_view.database_name
}

output "qualified_materialized_view" {
  value = materialize_materialized_view.simple_materialized_view.qualified_sql_name
}

data "materialize_materialized_view" "all" {}