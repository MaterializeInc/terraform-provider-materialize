resource "materialize_source_load_generator" "load_generator" {
  name                = "load_gen"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  comment             = "source load generator comment"
  cluster_name        = materialize_cluster.cluster_source.name
  load_generator_type = "COUNTER"

  counter_options {
    tick_interval = "500ms"
  }
  expose_progress {
    name = "expose_load_gen"
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

resource "materialize_source_load_generator" "load_generator_marketing" {
  name                = "load_gen_marketing"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  cluster_name        = materialize_cluster.cluster_source.name
  load_generator_type = "MARKETING"

  marketing_options {
    tick_interval = "500ms"
  }
}

resource "materialize_source_load_generator" "load_generator_tpch" {
  name                = "load_gen_tpch"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  cluster_name        = materialize_cluster.cluster_source.name
  load_generator_type = "TPCH"

  tpch_options {
    tick_interval = "500ms"
  }
}

resource "materialize_source_load_generator" "load_generator_key_value" {
  name                = "load_gen_key_value"
  schema_name         = materialize_schema.schema.name
  database_name       = materialize_database.database.name
  cluster_name        = materialize_cluster.cluster_source.name
  load_generator_type = "KEY VALUE"

  key_value_options {
    keys                   = 100
    snapshot_rounds        = 5
    transactional_snapshot = true
    value_size             = 256
    tick_interval          = "2s"
    seed                   = 11
    partitions             = 10
    batch_size             = 10
  }

  expose_progress {
    name = "expose_load_gen_key_value"
  }
}

resource "materialize_source_postgres" "example_source_postgres" {
  name         = "source_postgres"
  comment      = "source postgres comment"
  cluster_name = materialize_cluster.cluster_source.name
  text_columns = ["table1.id"]

  postgres_connection {
    name          = materialize_connection_postgres.postgres_connection.name
    schema_name   = materialize_connection_postgres.postgres_connection.schema_name
    database_name = materialize_connection_postgres.postgres_connection.database_name
  }
  publication = "mz_source"
  table {
    upstream_name        = "table1"
    upstream_schema_name = "public"
    name                 = "s1_table1"
  }
  table {
    upstream_name        = "table2"
    upstream_schema_name = "public"
    name                 = "s2_table1"
  }
  expose_progress {
    name = "expose_postgres"
  }
}

resource "materialize_source_kafka" "example_source_kafka_format_text" {
  name         = "source_kafka_text"
  comment      = "source kafka comment"
  cluster_name = materialize_cluster.cluster_source.name
  topic        = "topic1"

  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
    database_name = materialize_connection_kafka.kafka_connection.database_name
  }
  key_format {
    text = true
  }
  value_format {
    text = true
  }
  expose_progress {
    name = "expose_kafka"
  }

  depends_on = [materialize_sink_kafka.sink_kafka]
}

resource "materialize_source_kafka" "example_source_kafka_format_bytes" {
  name         = "source_kafka_bytes"
  cluster_name = materialize_cluster.cluster_source.name
  topic        = "topic1"

  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
    database_name = materialize_connection_kafka.kafka_connection.database_name
  }
  format {
    bytes = true
  }

  depends_on = [materialize_sink_kafka.sink_kafka]
}

resource "materialize_source_kafka" "example_source_kafka_format_avro" {
  name         = "source_kafka_avro"
  cluster_name = materialize_cluster.cluster_source.name
  topic        = "topic1"

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

resource "materialize_source_mysql" "test" {
  name         = "source_mysql"
  cluster_name = materialize_cluster.cluster_source.name

  mysql_connection {
    name = materialize_connection_mysql.mysql_connection.name
  }

  ignore_columns = ["shop.mysql_table2.id"]
  text_columns   = ["shop.mysql_table4.status"]

  table {
    upstream_name        = "mysql_table1"
    upstream_schema_name = "shop"
    name                 = "mysql_table1_local"
  }
  table {
    upstream_name        = "mysql_table2"
    upstream_schema_name = "shop"
    name                 = "mysql_table2_local"
  }
  table {
    upstream_name        = "mysql_table3"
    upstream_schema_name = "shop"
    name                 = "mysql_table3_local"
  }
  table {
    upstream_name        = "mysql_table4"
    upstream_schema_name = "shop"
    name                 = "mysql_table4_local"
  }
}

resource "materialize_source_grant" "source_grant_select" {
  role_name     = materialize_role.role_1.name
  privilege     = "SELECT"
  source_name   = materialize_source_load_generator.load_generator.name
  schema_name   = materialize_source_load_generator.load_generator.schema_name
  database_name = materialize_source_load_generator.load_generator.database_name
}

resource "materialize_source_kafka" "kafka_upsert_options_source" {
  name = "kafka_upsert_options_source"
  kafka_connection {
    name = materialize_connection_kafka.kafka_connection.name
  }

  # depends on sink_kafka_cluster to ensure that the topic exists
  depends_on = [materialize_sink_kafka.sink_kafka_cluster]

  cluster_name = materialize_cluster.cluster_source.name
  topic        = "topic1"
  key_format {
    text = true
  }
  value_format {
    text = true
  }
  envelope {
    upsert = true
    upsert_options {
      value_decoding_errors {
        inline {
          enabled = true
          alias   = "my_error_col"
        }
      }
    }
  }

  start_offset            = [0]
  include_timestamp_alias = "timestamp_alias"
  include_offset          = true
  include_offset_alias    = "offset_alias"
  include_partition       = true
  include_partition_alias = "partition_alias"
  include_key_alias       = "key_alias"
}

output "qualified_load_generator" {
  value = materialize_source_load_generator.load_generator.qualified_sql_name
}

data "materialize_source" "all" {}
