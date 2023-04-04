resource "materialize_connection_kafka" "kafka_connection" {
  name = "kafka_connection"
  kafka_broker {
    broker = "redpanda:9092"
  }
}

resource "materialize_connection_confluent_schema_registry" "schema_registry" {
  name = "schema_registry_connection"
  url  = "http://redpanda:8081"
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
    broker = "redpanda:9092"
  }
  kafka_broker {
    broker = "redpanda:9092"
  }
  sasl_username {
    text = "sasl_user"
  }
  sasl_password {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }
  sasl_mechanisms = "SCRAM-SHA-256"
  progress_topic  = "progress_topic"
}

resource "materialize_connection_postgres" "postgres_connection" {
  name = "postgres_connection"
  host = "postgres"
  port = 5432
  user {
    text = "postgres"
  }
  password {
    name          = materialize_secret.postgres_password.name
    database_name = materialize_secret.postgres_password.database_name
    schema_name   = materialize_secret.postgres_password.schema_name
  }
  database = "postgres"
}

resource "materialize_connection_postgres" "postgres_connection_with_secret" {
  name = "postgres_connection_with_secret"
  host = "postgres"
  port = 5432
  user {
    secret {
      name          = materialize_secret.postgres_password.name
      database_name = materialize_secret.postgres_password.database_name
      schema_name   = materialize_secret.postgres_password.schema_name
    }
  }
  password {
    name          = materialize_secret.postgres_password.name
    database_name = materialize_secret.postgres_password.database_name
    schema_name   = materialize_secret.postgres_password.schema_name
  }
  database = "postgres"
}

output "qualified_ssh_connection" {
  value = materialize_connection_ssh_tunnel.ssh_connection.qualified_sql_name
}
