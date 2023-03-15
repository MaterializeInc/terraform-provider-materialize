resource "materialize_sink_kafka" "example_sink_kafka" {
  name                       = "sink_kafka"
  schema_name                = "schema"
  size                       = "3xsmall"
  item_name                  = "schema.table"
  topic                      = "test_avro_topic"
  format                     = "AVRO"
  kafka_connection {
    name = "kafka_connection"
  }
  schema_registry_connection {
    name = "csr_connection"
  }
  envelope                   = "UPSERT"
}

# CREATE SINK schema.sink_kafka
#   FROM schema.table
#   INTO KAFKA CONNECTION kafka_connection (TOPIC 'test_avro_topic')
#   FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_connection
#   ENVELOPE UPSERT
#   WITH (SIZE = '3xsmall');
