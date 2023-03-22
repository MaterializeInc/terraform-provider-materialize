resource "materialize_index" "loadgen_index" {
  name         = "loadgen_index"
  cluster_name = materialize_cluster.cluster.name

  obj_name {
    name          = materialize_source_load_generator.load_generator_cluster.name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
    database_name = materialize_connection_kafka.kafka_connection.database_name
  }
}