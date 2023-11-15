resource "materialize_source_load_generator" "load_generator" {
  name          = "load_gen"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  comment       = "source load generator comment"

  size                = "3xsmall"
  load_generator_type = "COUNTER"

  counter_options {
    tick_interval = "500ms"
  }
}

resource "materialize_source_load_generator" "load_generator_cluster" {
  name                = "load_gen_cluster"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  cluster_name        = materialize_cluster.cluster_source.name
  load_generator_type = "COUNTER"

  counter_options {
    tick_interval = "500ms"
  }
}

resource "materialize_source_load_generator" "load_generator_auction" {
  name                = "load_gen_auction"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  cluster_name        = materialize_cluster.cluster_source.name
  load_generator_type = "AUCTION"

  auction_options {
    tick_interval = "500ms"
  }
}

resource "materialize_source_postgres" "example_source_postgres" {
  name    = "source_postgres"
  comment = "source postgres comment"

  size = "3xsmall"
  postgres_connection {
    name          = materialize_connection_postgres.postgres_connection.name
    schema_name   = materialize_connection_postgres.postgres_connection.schema_name
    database_name = materialize_connection_postgres.postgres_connection.database_name
  }
  publication = "mz_source"
  table {
    name  = "table1"
    alias = "s1_table1"
  }
  table {
    name  = "table2"
    alias = "s2_table1"
  }
  text_columns = ["table1.id"]
}

resource "materialize_source_postgres" "example_source_postgres_schema" {
  name = "source_postgres_schema"
  size = "3xsmall"
  postgres_connection {
    name          = materialize_connection_postgres.postgres_connection.name
    schema_name   = materialize_connection_postgres.postgres_connection.schema_name
    database_name = materialize_connection_postgres.postgres_connection.database_name
  }
  publication = "mz_source"
  schema      = ["PUBLIC"]
}

resource "materialize_source_kafka" "example_source_kafka_format_text" {
  name    = "source_kafka_text"
  comment = "source kafka comment"

  size = "3xsmall"
  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
    database_name = materialize_connection_kafka.kafka_connection.database_name
  }
  topic = "topic1"
  key_format {
    text = true
  }
  value_format {
    text = true
  }
}

resource "materialize_source_kafka" "example_source_kafka_format_bytes" {
  name = "source_kafka_bytes"
  size = "2xsmall"
  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
    database_name = materialize_connection_kafka.kafka_connection.database_name
  }
  topic = "topic1"
  format {
    bytes = true
  }
}

resource "materialize_source_kafka" "example_source_kafka_format_avro" {
  name = "source_kafka_avro"
  size = "3xsmall"
  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
    database_name = materialize_connection_kafka.kafka_connection.database_name
  }
  format {
    avro {
      schema_registry_connection {
        name          = materialize_connection_confluent_schema_registry.schema_registry.name
        schema_name   = materialize_connection_confluent_schema_registry.schema_registry.schema_name
        database_name = materialize_connection_confluent_schema_registry.schema_registry.database_name
      }
    }
  }
  envelope {
    none = true
  }
  topic      = "topic1"
  depends_on = [materialize_sink_kafka.sink_kafka]
}

resource "materialize_source_webhook" "example_webhook_source" {
  name             = "example_webhook_source"
  comment          = "source webhook comment"
  cluster_name     = materialize_cluster.cluster_source.name
  body_format      = "json"
  check_expression = "headers->'x-mz-api-key' = secret"

  include_headers {
    not = ["x-mz-api-key"]
  }

  check_options {
    field {
      headers = true
    }
  }

  check_options {
    field {
      secret {
        name          = materialize_secret.postgres_password.name
        database_name = materialize_secret.postgres_password.database_name
        schema_name   = materialize_secret.postgres_password.schema_name
      }
    }
    alias = "secret"
  }
}

resource "materialize_source_grant" "source_grant_select" {
  role_name     = materialize_role.role_1.name
  privilege     = "SELECT"
  source_name   = materialize_source_load_generator.load_generator.name
  schema_name   = materialize_source_load_generator.load_generator.schema_name
  database_name = materialize_source_load_generator.load_generator.database_name
}

output "qualified_load_generator" {
  value = materialize_source_load_generator.load_generator.qualified_sql_name
}

data "materialize_source" "all" {}
