resource "materialize_source_load_generator" "load_generator" {
  name                = "load_gen"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  size                = "1"
  load_generator_type = "COUNTER"
  tick_interval       = "500ms"
  scale_factor        = 0.01
}

resource "materialize_source_load_generator" "load_generator_cluster" {
  name                = "load_gen_cluster"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  cluster_name        = materialize_cluster.cluster_source.name
  load_generator_type = "COUNTER"
  tick_interval       = "500ms"
  scale_factor        = 0.01
}

resource "materialize_source_postgres" "example_source_postgres" {
  name                = "source_postgres"
  size                = "2"
  postgres_connection = materialize_connection.example_postgres_connection.qualified_name
  publication         = "mz_source"
  tables {
    name  = "table1"
    alias = "s1_table1"
  }
  tables {
    name  = "table2"
    alias = "s2_table1"
  }
}

resource "materialize_source_kafka" "example_source_kafka_format_text" {
  name                       = "source_kafka_text"
  size                       = "2"
  kafka_connection           = materialize_connection.kafka_connection.qualified_name
  format                     = "TEXT"
  topic                      = "topic1"
  key_format                 = "TEXT"
}

# resource "materialize_source_kafka" "example_source_kafka_format_avro" {
#   name                       = "source_kafka_avro"
#   size                       = "2"
#   kafka_connection           = materialize_connection.kafka_connection.qualified_name
#   format                     = "AVRO"
#   topic                      = "topic1"
#   schema_registry_connection = materialize_connection.schema_registry.name
#   key_format                 = "TEXT"
# }

output "qualified_load_generator" {
  value = materialize_source_load_generator.load_generator.qualified_name
}
