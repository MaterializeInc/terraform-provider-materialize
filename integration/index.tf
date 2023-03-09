resource "materialize_index" "index" {
  name         = "index"
  obj_name     = materialize_source_load_generator.load_generator_cluster.qualified_name
  cluster_name = materialize_cluster.cluster.name
}