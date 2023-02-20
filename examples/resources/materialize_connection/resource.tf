# Create SSH Connection
resource "materialize_connection" "example_ssh_connection" {
  name            = "ssh_example_connection"
  schema_name     = "public"
  connection_type = "SSH TUNNEL"
  ssh_host        = "example.com"
  ssh_port        = 22
  ssh_user        = "example"
}

# CREATE CONNECTION ssh_example_connection TO SSH TUNNEL (
#    HOST 'example.com',
#    PORT 22,
#    USER 'example'
# );

# # Create a AWS Private Connection
# Note: you need the max_aws_privatelink_connections increased for this to work:
# show max_aws_privatelink_connections;
resource "materialize_connection" "example_privatelink_connection" {
  name                               = "example_privatelink_connection"
  schema_name                        = "public"
  connection_type                    = "AWS PRIVATELINK"
  aws_privatelink_service_name       = "com.amazonaws.us-east-1.materialize.example"
  aws_privatelink_availability_zones = ["use1-az2", "use1-az6"]
}

# CREATE CONNECTION example_privatelink_connection TO AWS PRIVATELINK (
#     SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',
#     AVAILABILITY ZONES ('use1-az2', 'use1-az6')
# );

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
  kafka_broker    = "example.com:9092"
  # kafka_brokers         = [{
  #   "broker": "b-1.hostname-1:9096",
  # },
  # {
  #   "broker": "b-2.hostname-2:9096",
  # }]
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
