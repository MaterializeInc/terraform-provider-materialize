resource "materialize_source" "example_source_kafka" {
  name                       = "source_kafka"
  schema_name                = "schema"
  size                       = "3xsmall"
  kafka_connection           = "kafka_connection"
  schema_registry_connection = "csr_connection"
  format                     = "AVRO"
  envelope                   = "data"
}

# CREATE SOURCE kafka_metadata
#   FROM KAFKA CONNECTION kafka_connection (TOPIC 'data')
#   FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION csr_connection
#   ENVELOPE NONE
#   WITH (SIZE = '3xsmall');