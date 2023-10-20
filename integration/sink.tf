resource "materialize_sink_kafka" "sink_kafka" {
  name          = "sink_kafka"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  comment       = "sink comment"
  size          = "3xsmall"
  from {
    name          = materialize_source_load_generator.load_generator.name
    database_name = materialize_source_load_generator.load_generator.database_name
    schema_name   = materialize_source_load_generator.load_generator.schema_name
  }
  topic = "topic1"
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
}

output "qualified_sink_kafka" {
  value = materialize_sink_kafka.sink_kafka.qualified_sql_name
}

data "materialize_sink" "all" {}
