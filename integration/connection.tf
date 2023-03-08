resource "materialize_connection_kafka" "kafka_connection" {
  name = "kafka_connection"
  kafka_broker {
    broker = "kafka:9092"
  }
}

resource "materialize_connection_confluent_schema_registry" "schema_registry" {
  name = "schema_registry_connection"
  url  = "http://schema-registry:8081"
}

resource "materialize_connection_ssh_tunnel" "ssh_connection" {
  name        = "ssh_connection"
  schema_name = "public"
  host        = "ssh_host"
  user        = "ssh_user"
  port        = 22
}

resource "materialize_connection_kafka" "kafka_conn_multiple_brokers" {
  name = "kafka_conn_multiple_brokers"
  kafka_broker {
    broker = "kafka:9092"
  }
  kafka_broker {
    broker = "kafka2:9092"
  }
  sasl_username   = "sasl_user"
  sasl_password   = materialize_secret.kafka_password.qualified_name
  sasl_mechanisms = "SCRAM-SHA-256"
  progress_topic  = "progress_topic"
}

resource "materialize_connection_postgres" "postgres_connection" {
  name     = "postgres_connection"
  host     = "postgres"
  port     = 5432
  user     = "postgres"
  password = materialize_secret.postgres_password.qualified_name
  database = "postgres"
}

output "qualified_ssh_connection" {
  value = materialize_connection_ssh_tunnel.ssh_connection.qualified_name
}
