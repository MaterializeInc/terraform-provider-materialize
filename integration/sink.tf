resource "materialize_sink_kafka" "sink_kafka" {
  name                       = "sink_kafka"
  schema_name                = materialize_schema.schema.name
  database_name              = materialize_database.database.name
  size                       = "1"
  item_name                  = "example.example.load_gen"
  kafka_connection           = materialize_connection_kafka.kafka_connection.name
  topic                      = "topic1"
  format                     = "AVRO"
  schema_registry_connection = materialize_connection.schema_registry.name
  envelope                   = "DEBEZIUM"

  depends_on = [
    materialize_source_load_generator.load_generator
  ]
}

output "qualified_sink_kafka" {
  value = materialize_sink_kafka.sink_kafka.qualified_name
}
