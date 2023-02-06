resource "materialize_secret" "example_sink_kafka" {
  name                       = "sink_kafka"
  schema_name                = "schema"
  size                       = "3xsmall"
  item_name                  = "schema.table"
  kafka_connection           = "kafka_connection"
  topic                      = "test_avro_topic"
  format                     = "AVRO"
  schema_registry_connection = "csr_connection"
  envelope                   = "UPSERT"
}

# CREATE SINK schema.sink_kafka
#   FROM schema.table
#   INTO KAFKA CONNECTION kafka_connection (TOPIC 'test_avro_topic')
#   FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_connection
#   ENVELOPE UPSERT
#   WITH (SIZE = '3xsmall');