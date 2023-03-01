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
