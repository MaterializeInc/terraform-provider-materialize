resource "materialize_source" "load_generator" {
  name                = "load_gen"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  size                = "1"
  connection_type     = "LOAD GENERATOR"
  load_generator_type = "COUNTER"
  tick_interval       = "500ms"
  scale_factor        = 0.01
}

resource "materialize_source" "load_generator_cluster" {
  name                = "load_gen_cluster"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  cluster_name        = materialize_cluster.cluster_source.name
  connection_type     = "LOAD GENERATOR"
  load_generator_type = "COUNTER"
  tick_interval       = "500ms"
  scale_factor        = 0.01
}

# resource "materialize_source" "example_source_kafka" {
#   name             = "source_kafka"
#   schema_name      = materialize_schema.schema.name
#   database_name    = materialize_database.database.name
#   cluster_name     = materialize_cluster.cluster_source.name
#   connection_type  = "KAFKA"
#   kafka_connection = materialize_connection.kafka_connection.name
#   topic            = "TOPIC"
#   format           = "AVRO"
#   envelope         = "UPSERT"
# }

output "qualified_load_generator" {
  value = materialize_source.load_generator.qualified_name
}