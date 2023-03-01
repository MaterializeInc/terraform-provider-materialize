# Create a Postgres Connection
resource "materialize_connection" "example_postgres_connection" {
  name              = "example_postgres_connection"
  connection_type   = "POSTGRES"
  postgres_host     = "instance.foo000.us-west-1.rds.amazonaws.com"
  postgres_port     = 5432
  postgres_user     = "example"
  postgres_password = "example"
  postgres_database = "example"
}

# CREATE CONNECTION example_postgres_connection TO POSTGRES (
#     HOST 'instance.foo000.us-west-1.rds.amazonaws.com',
#     PORT 5432,
#     USER 'example',
#     PASSWORD SECRET example,
#     DATABASE 'example'
# );


# Create a Kafka Connection
resource "materialize_connection" "example_kafka_connection" {
  name            = "example_kafka_connection"
  connection_type = "KAFKA"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
  }
  kafka_sasl_username   = "example"
  kafka_sasl_password   = "kafka_password"
  kafka_sasl_mechanisms = "SCRAM-SHA-256"
  kafka_progress_topic  = "example"
}

# CREATE CONNECTION database.schema.kafka_conn TO KAFKA (
#     BROKER 'example:9092'
#     PROGRESS TOPIC 'topic',
#     SASL MECHANISMS 'PLAIN',
#     SASL USERNAME 'user',
#     SASL PASSWORD SECRET password
# );

# Create a Confluent Schema Registry Connection
resource "materialize_connection" "example_confluent_schema_registry_connection" {
  name                               = "example_csr_connection"
  connection_type                    = "CONFLUENT SCHEMA REGISTRY"
  confluent_schema_registry_url      = "https://rp-f00000bar.data.vectorized.cloud:30993"
  confluent_schema_registry_password = "example"
  confluent_schema_registry_username = "example"
}

# CREATE CONNECTION example_csr_connection TO CONFLUENT SCHEMA REGISTRY (
#     URL 'https://rp-f00000bar.data.vectorized.cloud:30993',
#     USERNAME = 'example',
#     PASSWORD = SECRET example
# );

resource "materialize_connection" "example_kafka_connection" {
  name            = "example_kafka_connection"
  connection_type = "KAFKA"
  kafka_broker {
    broker = "b-1.hostname-1:9096"
    target_group_port = "9001"
    availability_zone = "use1-az1"
    privatelink_connection = "privatelink_conn"
  }
  kafka_broker {
    broker = "b-2.hostname-2:9096"
    target_group_port = "9002"
    availability_zone = "use1-az2"
    privatelink_connection = "privatelink_conn"
  }
}

# CREATE CONNECTION materialize.public.example_kafka_connection TO KAFKA (
#     BROKERS (
#        'b-1.hostname-1:9096' USING AWS PRIVATELINK privatelink_conn (PORT 9001, AVAILABILITY ZONE 'use1-az1'),
#        'b-2.hostname-2:9096' USING AWS PRIVATELINK privatelink_conn (PORT 9002, AVAILABILITY ZONE 'use1-az2')
#     )
# );
