resource "materialize_connection" "ssh_connection" {
  name            = "ssh_connection"
  connection_type = "SSH TUNNEL"
  ssh_host        = "example.com"
  ssh_port        = 22
  ssh_user        = "example"
}

resource "materialize_connection" "kafka_connection" {
  name            = "kafka_connection"
  connection_type = "KAFKA"
  kafka_brokers {
    broker = "kafka:9092"
  }
}

resource "materialize_connection" "schema_registry" {
  name                          = "schema_registry_connection"
  connection_type               = "CONFLUENT SCHEMA REGISTRY"
  confluent_schema_registry_url = "http://schema-registry:8081"
}

resource "materialize_connection" "example_ssh_connection" {
  name            = "ssh_example_connection"
  schema_name     = "public"
  connection_type = "SSH TUNNEL"
  ssh_host        = "ssh_host"
  ssh_user        = "ssh_user"
  ssh_port        = 22
}

resource "materialize_connection" "kafka_conn_multiple_brokers" {
  name            = "kafka_conn_multiple_brokers"
  connection_type = "KAFKA"
  kafka_brokers {
    broker = "kafka:9092"
  }
  kafka_brokers {
    broker = "kafka2:9092"
  }
  kafka_sasl_username   = "sasl_user"
  kafka_sasl_password   = format("%s.%s.%s", materialize_database.database.name, materialize_schema.schema.name, materialize_secret.kafka_password.name)
  kafka_sasl_mechanisms = "SCRAM-SHA-256"
  kafka_progress_topic  = "progress_topic"
}

resource "materialize_connection" "example_postgres_connection" {
  name              = "example_postgres_connection"
  connection_type   = "POSTGRES"
  postgres_host     = "postgres"
  postgres_port     = 5432
  postgres_user     = "materialize"
  postgres_password = format("%s.%s.%s", materialize_database.database.name, materialize_schema.schema.name, materialize_secret.postgres_password.name)
  postgres_database = "postgres"
}

output "qualified_ssh_connection" {
  value = materialize_connection.ssh_connection.qualified_name
}
