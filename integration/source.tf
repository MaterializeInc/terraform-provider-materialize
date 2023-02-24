resource "materialize_source_loadgen" "load_generator" {
  name                = "load_gen"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  size                = "1"
  load_generator_type = "COUNTER"
  tick_interval       = "500ms"
  scale_factor        = 0.01
}

resource "materialize_source_loadgen" "load_generator_cluster" {
  name                = "load_gen_cluster"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  cluster_name        = materialize_cluster.cluster_source.name
  load_generator_type = "COUNTER"
  tick_interval       = "500ms"
  scale_factor        = 0.01
}

output "qualified_load_generator" {
  value = materialize_source_loadgen.load_generator.qualified_name
}