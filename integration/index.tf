resource "materialize_index" "loadgen_index" {
  name         = "loadgen_index"
  cluster_name = materialize_cluster.cluster.name

  obj_name {
    name          = materialize_source_load_generator.load_generator_cluster.name
    schema_name   = materialize_source_load_generator.load_generator_cluster.schema_name
    database_name = materialize_source_load_generator.load_generator_cluster.database_name
  }
}

output "qualified_index" {
  value = materialize_index.loadgen_index.qualified_sql_name
}

data "materialize_index" "all" {}
