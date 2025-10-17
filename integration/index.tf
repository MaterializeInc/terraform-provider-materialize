resource "materialize_index" "loadgen_index" {
  name    = "loadgen_index"
  comment = "index comment"

  cluster_name = materialize_cluster.cluster.name

  obj_name {
    name          = materialize_source_load_generator.load_generator_cluster.name
    schema_name   = materialize_source_load_generator.load_generator_cluster.schema_name
    database_name = materialize_source_load_generator.load_generator_cluster.database_name
  }

  col_expr {
    field = "key"
  }
}

# Create in separate region
resource "materialize_index" "loadgen_index_us_west" {
  name    = "loadgen_index"
  comment = "index comment"
  region  = "aws/us-west-2"

  cluster_name = materialize_cluster.cluster_source_us_west.name

  obj_name {
    name          = materialize_source_load_generator.load_generator_cluster_us_west.name
    schema_name   = materialize_source_load_generator.load_generator_cluster_us_west.schema_name
    database_name = materialize_source_load_generator.load_generator_cluster_us_west.database_name
  }

  col_expr {
    field = "key"
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

resource "materialize_index" "materialized_view_default_index" {
  cluster_name = "quickstart"

  default = true

  obj_name {
    name          = materialize_materialized_view.simple_materialized_view.name
    schema_name   = materialize_materialized_view.simple_materialized_view.schema_name
    database_name = materialize_materialized_view.simple_materialized_view.database_name
  }

}

output "qualified_index" {
  value = materialize_index.loadgen_index.qualified_sql_name
}

data "materialize_index" "all" {}
