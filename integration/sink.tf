resource "materialize_sink_kafka" "sink_kafka" {
  name          = "sink_kafka"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  size          = "1"
  from {
    name          = "load_gen"
    database_name = "example"
    schema_name   = "example"
  }
  topic  = "topic1"
  format {
    avro {
        schema_registry_connection {
        name          = materialize_connection_confluent_schema_registry.schema_registry.name
        database_name = materialize_connection_confluent_schema_registry.schema_registry.database_name
        schema_name   = materialize_connection_confluent_schema_registry.schema_registry.schema_name
      }
    }
  }
  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    database_name = materialize_connection_kafka.kafka_connection.database_name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
  }
  envelope {
    debezium = true
  }

  depends_on = [
    materialize_source_load_generator.load_generator
  ]
}

output "qualified_sink_kafka" {
  value = materialize_sink_kafka.sink_kafka.qualified_name
}
