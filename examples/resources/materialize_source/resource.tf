resource "materialize_source" "example_source_load_generator" {
  name                = "source_load_generator"
  schema_name         = "schema"
  size                = "3xsmall"
  connection_type     = "LOAD GENERATOR"
  load_generator_type = "COUNTER"
  tick_interval       = "500ms"
  scale_factor        = 0.01
}

# CREATE SOURCE schema.source_load_generator
#   FROM LOAD GENERATOR COUNTER
#   (TICK INTERVAL '500ms' SCALE FACTOR 0.01)
#   WITH (SIZE = '3xsmall');

resource "materialize_source" "example_source_postgres" {
  name                = "source_postgres"
  schema_name         = "schema"
  size                = "3xsmall"
  connection_type     = "POSTGRES"
  postgres_connection = "pg_connection"
  publication         = "mz_source"
  tables = {
    "schema1.table_1" = "s1_table_1"
    "schema2_table_1" = "s2_table_1"
  }
}

# CREATE SOURCE schema.source_postgres
#   FROM POSTGRES CONNECTION pg_connection (PUBLICATION 'mz_source')
#   FOR TABLES (schema1.table_1 AS s1_table_1, schema2_table_1 AS s2_table_1)
#   WITH (SIZE = '3xsmall');

resource "materialize_source" "example_source_kafka" {
  name                       = "source_kafka"
  schema_name                = "schema"
  size                       = "3xsmall"
  connection_type            = "KAFKA"
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