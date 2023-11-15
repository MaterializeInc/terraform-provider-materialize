resource "materialize_connection_kafka" "kafka_connection" {
  name    = "kafka_connection"
  comment = "connection kafka comment"

  kafka_broker {
    broker = "redpanda:9092"
  }
}

resource "materialize_connection_confluent_schema_registry" "schema_registry" {
  name    = "schema_registry_connection"
  comment = "connection schema registry comment"

  url = "http://redpanda:8081"
}

resource "materialize_connection_ssh_tunnel" "ssh_connection" {
  name        = "ssh_connection"
  schema_name = "public"
  comment     = "connection ssh tunnel comment"

  host = "ssh_host"
  user = "ssh_user"
  port = 22
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
    text = "sasl-user"
  }
  sasl_password {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }
  security_protocol = "SASL_PLAINTEXT"
  sasl_mechanisms   = "SCRAM-SHA-256"
  progress_topic    = "progress_topic"
  validate          = false
}

resource "materialize_connection_postgres" "postgres_connection" {
  name    = "postgres_connection"
  comment = "connection postgres comment"

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
  name = "postgres-connection-with-secret"
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
  validate = false
}

resource "materialize_connection_grant" "connection_grant_usage" {
  role_name       = materialize_role.role_1.name
  privilege       = "USAGE"
  connection_name = materialize_connection_postgres.postgres_connection.name
  schema_name     = materialize_connection_postgres.postgres_connection.schema_name
  database_name   = materialize_connection_postgres.postgres_connection.database_name
}

resource "materialize_connection_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
}

output "qualified_ssh_connection" {
  value = materialize_connection_ssh_tunnel.ssh_connection.qualified_sql_name
}

data "materialize_connection" "all" {}
