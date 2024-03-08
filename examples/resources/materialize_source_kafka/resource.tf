resource "materialize_source_kafka" "example_source_kafka" {
  name         = "source_kafka"
  schema_name  = "schema"
  cluster_name = "quickstart"
  kafka_connection {
    name          = "kafka_connection"
    database_name = "database"
    schema_name   = "schema"
  }
  format {
    avro {
      schema_registry_connection {
        name          = "csr_connection"
        database_name = "database"
        schema_name   = "schema"
      }
    }
  }
  envelope {
    none = true
  }
}

# CREATE SOURCE kafka_metadata
#   FROM KAFKA CONNECTION "database"."schema"."kafka_connection" (TOPIC 'data')
#   FORMAT AVRO USING CONFLUENT SCHEMA REGISTRY CONNECTION "database"."schema"."csr_connection"
#   ENVELOPE NONE;
