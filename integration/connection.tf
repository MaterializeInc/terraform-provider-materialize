resource "materialize_connection_kafka" "kafka_connection" {
  name            = "kafka_connection"
  kafka_broker {
    broker = "kafka:9092"
  }
}

resource "materialize_connection" "schema_registry" {
  name                          = "schema_registry_connection"
  connection_type               = "CONFLUENT SCHEMA REGISTRY"
  confluent_schema_registry_url = "http://schema-registry:8081"
}

resource "materialize_connection_ssh_tunnel" "example_ssh_connection" {
  name            = "ssh_example_connection"
  schema_name     = "public"
  host        = "ssh_host"
  user        = "ssh_user"
  port        = 22
}

resource "materialize_connection_kafka" "kafka_conn_multiple_brokers" {
  name            = "kafka_conn_multiple_brokers"
  kafka_broker {
    broker = "kafka:9092"
  }
  kafka_broker {
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
  value = materialize_connection_ssh_tunnel.example_ssh_connection.qualified_name
}
