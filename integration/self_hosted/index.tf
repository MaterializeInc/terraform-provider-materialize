resource "materialize_index" "loadgen_index" {
  name    = "loadgen_index"
  comment = "index comment"

  cluster_name = materialize_cluster.cluster.name

  obj_name {
    name          = "accounts"
    schema_name   = materialize_source_load_generator.load_generator_cluster.schema_name
    database_name = materialize_source_load_generator.load_generator_cluster.database_name
  }

  col_expr {
    field = "id"
  }
}

resource "materialize_index" "materialized_view_index" {
  name         = "simple"
  cluster_name = "quickstart"

  obj_name {
    name          = materialize_materialized_view.simple_materialized_view.name
    schema_name   = materialize_materialized_view.simple_materialized_view.schema_name
    database_name = materialize_materialized_view.simple_materialized_view.database_name
  }

  col_expr {
    field = "id"
  }
}

output "qualified_index" {
  value = materialize_index.loadgen_index.qualified_sql_name
}

data "materialize_index" "all" {}
