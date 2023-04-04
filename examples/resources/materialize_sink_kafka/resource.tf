resource "materialize_sink_kafka" "example_sink_kafka" {
  name        = "sink_kafka"
  schema_name = "schema"
  size        = "3xsmall"
  from {
    name = "table"
  }
  topic = "test_avro_topic"
  format {
    avro {
      schema_registry_connection {
        name          = "csr_connection"
        database_name = "database"
        schema_name   = "schema"
      }
    }
  }
  kafka_connection {
    name = "kafka_connection"
  }
  envelope {
    upsert = true
  }
}

# CREATE SINK schema.sink_kafka
#   FROM schema.table
#   INTO KAFKA CONNECTION "kafka_connection" (TOPIC 'test_avro_topic')
#   FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_connection"
#   ENVELOPE UPSERT
#   WITH (SIZE = '3xsmall');
