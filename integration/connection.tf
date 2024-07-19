resource "materialize_connection_kafka" "kafka_connection" {
  name              = "kafka_connection"
  comment           = "connection kafka comment"
  security_protocol = "PLAINTEXT"

  kafka_broker {
    broker = "redpanda:9092"
  }
  validate = true
}

# Create in separate region
resource "materialize_connection_kafka" "kafka_connection_us_west" {
  name              = "kafka_connection"
  comment           = "connection kafka comment"
  security_protocol = "PLAINTEXT"

  kafka_broker {
    broker = "redpanda:9092"
  }
  region   = "aws/us-west-2"
  validate = true
}

resource "materialize_connection_kafka" "kafka_conn_ssl_auth" {
  name              = "kafka_conn_ssl_auth"
  security_protocol = "SSL"

  kafka_broker {
    broker = "redpanda:9092"
  }

  ssl_certificate {
    text = "certificate-content"
  }

  ssl_key {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  ssl_certificate_authority {
    text = "ca-content"
  }

  validate = false
}

# Create in separate region
resource "materialize_connection_kafka" "kafka_conn_ssl_auth_us_west" {
  name              = "kafka_conn_ssl_auth"
  security_protocol = "SSL"

  kafka_broker {
    broker = "redpanda:9092"
  }

  ssl_certificate {
    text = "certificate-content"
  }

  ssl_key {
    name          = materialize_secret.kafka_password_us_west.name
    database_name = materialize_secret.kafka_password_us_west.database_name
    schema_name   = materialize_secret.kafka_password_us_west.schema_name
  }

  ssl_certificate_authority {
    text = "ca-content"
  }

  region   = "aws/us-west-2"
  validate = false
}

resource "materialize_connection_kafka" "kafka_ssh_tunnel_connection" {
  name              = "kafka_ssh_tunnel_connection"
  security_protocol = "PLAINTEXT"

  kafka_broker {
    broker = "redpanda:9092"
  }

  ssh_tunnel {
    name = materialize_connection_ssh_tunnel.ssh_connection.name
  }

  validate = false
}

resource "materialize_connection_kafka" "kafka_sasl_ssl" {
  name              = "kafka_sasl_ssl"
  security_protocol = "SASL_SSL"

  kafka_broker {
    broker = "redpanda:9092"
  }

  sasl_mechanisms = "SCRAM-SHA-256"

  sasl_username {
    text = "sasl_username"
  }

  sasl_password {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  ssl_certificate {
    text = "ssl_certificate_content"
  }

  ssl_key {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  ssl_certificate_authority {
    text = "ssl_ca_content"
  }

  validate = false
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
  security_protocol                 = "SASL_PLAINTEXT"
  sasl_mechanisms                   = "SCRAM-SHA-256"
  progress_topic                    = "progress_topic"
  progress_topic_replication_factor = 1
  validate                          = false
}

resource "materialize_connection_kafka" "kafka_top_level_privatelink" {
  name = "kafka_top_level_privatelink"
  # The Privatelink connection is created during the docker-compose setup
  # As if you were to drop the privatelink connection, the container would crash
  aws_privatelink {
    privatelink_connection {
      name          = "privatelink_conn"
      database_name = "materialize"
      schema_name   = "public"
    }
    privatelink_connection_port = 9092
  }

  security_protocol = "SASL_SSL"
  sasl_mechanisms   = "SCRAM-SHA-256"

  sasl_username {
    text = "sasl_username"
  }

  sasl_password {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  ssl_certificate {
    text = "ssl_certificate_content"
  }

  ssl_key {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  ssl_certificate_authority {
    text = "ssl_ca_content"
  }

  validate = false
}

resource "materialize_connection_confluent_schema_registry" "schema_registry" {
  name    = "schema_registry_connection"
  comment = "connection schema registry comment"

  url = "http://redpanda:8081"
}

resource "materialize_connection_confluent_schema_registry" "csr_with_basic_auth" {
  name = "csr_with_basic_auth"
  url  = "http://redpanda:8081"

  username {
    text = "username"
  }

  password {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  validate = false
}

resource "materialize_connection_confluent_schema_registry" "csr_with_aws_privatelink" {
  name = "csr_with_aws_privatelink"
  url  = "http://redpanda:8081"

  username {
    text = "username"
  }

  password {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  aws_privatelink {
    name          = "privatelink_conn"
    database_name = "materialize"
    schema_name   = "public"
  }

  validate = false
}

resource "materialize_connection_confluent_schema_registry" "schema_registry_basic_auth_ssl" {
  name = "schema_registry_basic_auth_ssl"
  url  = "http://redpanda:8081"

  username {
    text = "schema_registry_user"
  }

  password {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  ssl_certificate {
    text = "ssl_certificate_content"
  }

  ssl_key {
    name          = materialize_secret.kafka_password.name
    database_name = materialize_secret.kafka_password.database_name
    schema_name   = materialize_secret.kafka_password.schema_name
  }

  ssl_certificate_authority {
    text = "ssl_ca_content" #
  }

  validate = false
}

resource "materialize_connection_confluent_schema_registry" "schema_registry_ssh_tunnel" {
  name = "schema_registry_ssh_tunnel"
  url  = "http://redpanda:8081"

  ssh_tunnel {
    name = materialize_connection_ssh_tunnel.ssh_connection.name
  }

  validate = false
}

resource "materialize_connection_ssh_tunnel" "ssh_connection" {
  name        = "ssh_connection"
  schema_name = "public"
  comment     = "connection ssh tunnel comment"

  host = "ssh_host"
  user = "ssh_user"
  port = 22

  validate = false
}

resource "materialize_connection_kafka" "kafka_conn_ssh_default" {
  name = "kafka_conn_ssh_default"
  kafka_broker {
    broker = "redpanda:9092"
  }
  ssh_tunnel {
    name          = materialize_connection_ssh_tunnel.ssh_connection.name
    database_name = materialize_connection_ssh_tunnel.ssh_connection.database_name
    schema_name   = materialize_connection_ssh_tunnel.ssh_connection.schema_name
  }
  validate = false
}

resource "materialize_connection_kafka" "kafka_conn_ssh_broker" {
  name = "kafka_conn_ssh_broker"
  kafka_broker {
    broker = "redpanda:9092"
    ssh_tunnel {
      name          = materialize_connection_ssh_tunnel.ssh_connection.name
      database_name = materialize_connection_ssh_tunnel.ssh_connection.database_name
      schema_name   = materialize_connection_ssh_tunnel.ssh_connection.schema_name
    }
  }
  kafka_broker {
    broker = "redpanda:9092"
  }
  validate = false
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

resource "materialize_connection_postgres" "postgres_ssh_tunnel_connection" {
  name     = "postgres_ssh_tunnel_connection"
  host     = "postgres"
  port     = 5432
  database = "postgres"
  user {
    text = "postgres"
  }
  password {
    name          = materialize_secret.postgres_password.name
    database_name = materialize_secret.postgres_password.database_name
    schema_name   = materialize_secret.postgres_password.schema_name
  }
  ssh_tunnel {
    name = materialize_connection_ssh_tunnel.ssh_connection.name
  }
  validate = false
}

resource "materialize_connection_postgres" "postgres_ssl_connection" {
  name     = "postgres_ssl_connection"
  host     = "postgres"
  port     = 5432
  database = "postgres"
  user {
    text = "postgres"
  }
  password {
    name          = materialize_secret.postgres_password.name
    database_name = materialize_secret.postgres_password.database_name
    schema_name   = materialize_secret.postgres_password.schema_name
  }
  ssl_mode = "require"
  ssl_certificate {
    text = "client_certificate_content"
  }
  ssl_key {
    name          = materialize_secret.postgres_password.name
    database_name = materialize_secret.postgres_password.database_name
    schema_name   = materialize_secret.postgres_password.schema_name
  }
  ssl_certificate_authority {
    text = "ca_certificate_content"
  }
  validate = false
}

resource "materialize_connection_mysql" "mysql_connection" {
  name    = "mysql_connection"
  comment = "connection mysql comment"

  host = "mysql"
  port = 3306
  user {
    text = "repluser"
  }
  password {
    name          = materialize_secret.mysql_password.name
    database_name = materialize_secret.mysql_password.database_name
    schema_name   = materialize_secret.mysql_password.schema_name
  }
}

resource "materialize_connection_mysql" "mysql_connection_aws_pl" {
  name    = "mysql_connection_aws_pl"
  comment = "connection mysql comment with aws privatelink"

  host = "mysql"
  port = 3306
  user {
    text = "repluser"
  }
  password {
    name          = materialize_secret.mysql_password.name
    database_name = materialize_secret.mysql_password.database_name
    schema_name   = materialize_secret.mysql_password.schema_name
  }

  aws_privatelink {
    name          = "privatelink_conn"
    database_name = "materialize"
    schema_name   = "public"
  }
  validate = false
}

resource "materialize_connection_mysql" "mysql_connection_with_secret" {
  name = "mysql-connection-with-secret"
  host = "mysql"
  port = 3306
  user {
    secret {
      name          = materialize_secret.mysql_user.name
      database_name = materialize_secret.mysql_user.database_name
      schema_name   = materialize_secret.mysql_user.schema_name
    }
  }
  password {
    name          = materialize_secret.mysql_password.name
    database_name = materialize_secret.mysql_password.database_name
    schema_name   = materialize_secret.mysql_password.schema_name
  }
  validate = false
}

resource "materialize_connection_mysql" "mysql_ssh_tunnel_connection" {
  name = "mysql_ssh_tunnel_connection"
  host = "mysql"
  port = 3306
  user {
    text = "repluser"
  }
  password {
    name          = materialize_secret.mysql_password.name
    database_name = materialize_secret.mysql_password.database_name
    schema_name   = materialize_secret.mysql_password.schema_name
  }
  ssh_tunnel {
    name = materialize_connection_ssh_tunnel.ssh_connection.name
  }
  validate = false
}

resource "materialize_connection_mysql" "mysql_ssl_connection" {
  name = "mysql_ssl_connection"
  host = "mysql"
  port = 3306
  user {
    text = "repluser"
  }
  password {
    name          = materialize_secret.mysql_password.name
    database_name = materialize_secret.mysql_password.database_name
    schema_name   = materialize_secret.mysql_password.schema_name
  }
  ssl_mode = "verify-identity"
  ssl_certificate {
    text = "client_certificate_content"
  }
  ssl_key {
    name          = materialize_secret.mysql_password.name
    database_name = materialize_secret.mysql_password.database_name
    schema_name   = materialize_secret.mysql_password.schema_name
  }
  ssl_certificate_authority {
    text = "ca_certificate_content"
  }
  validate = false
}

resource "materialize_connection_aws" "aws_conn" {
  name     = "aws_conn"
  endpoint = "http://localhost:4566"
  access_key_id {
    text = "test"
  }
  secret_access_key {
    name          = materialize_secret.aws_password.name
    database_name = materialize_secret.aws_password.database_name
    schema_name   = materialize_secret.aws_password.schema_name
  }
  validate = false
}

resource "materialize_connection_grant" "connection_grant_usage" {
  role_name       = materialize_role.role_1.name
  privilege       = "USAGE"
  connection_name = materialize_connection_postgres.postgres_connection.name
  schema_name     = materialize_connection_postgres.postgres_connection.schema_name
  database_name   = materialize_connection_postgres.postgres_connection.database_name
}

resource "materialize_connection_grant" "mysql_connection_grant_usage" {
  role_name       = materialize_role.role_1.name
  privilege       = "USAGE"
  connection_name = materialize_connection_mysql.mysql_connection.name
  schema_name     = materialize_connection_mysql.mysql_connection.schema_name
  database_name   = materialize_connection_mysql.mysql_connection.database_name
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
