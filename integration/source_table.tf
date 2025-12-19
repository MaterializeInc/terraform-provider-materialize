# Kafka Source Table
resource "materialize_source_table_kafka" "kafka_table" {
  name          = "kafka_source_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  source {
    name          = materialize_source_kafka.example_source_kafka_format_text.name
    schema_name   = materialize_source_kafka.example_source_kafka_format_text.schema_name
    database_name = materialize_source_kafka.example_source_kafka_format_text.database_name
  }

  topic = "topic1"

  key_format {
    text = true
  }

  value_format {
    json = true
  }

  envelope {
    upsert = true
  }

  include_key             = true
  include_key_alias       = "message_key"
  include_headers         = true
  include_headers_alias   = "message_headers"
  include_partition       = true
  include_partition_alias = "message_partition"
  include_offset          = true
  include_offset_alias    = "message_offset"
  include_timestamp       = true
  include_timestamp_alias = "message_timestamp"

  comment = "Kafka source table integration test"
}

# Postgres Source Table
resource "materialize_source_table_postgres" "postgres_table" {
  name          = "postgres_source_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  source {
    name          = materialize_source_postgres.example_source_postgres.name
    schema_name   = materialize_source_postgres.example_source_postgres.schema_name
    database_name = materialize_source_postgres.example_source_postgres.database_name
  }

  upstream_name        = "table1"
  upstream_schema_name = "public"

  text_columns = ["id"]
  comment      = "Postgres source table integration test"
}

# MySQL Source Table
resource "materialize_source_table_mysql" "mysql_table" {
  name          = "mysql_source_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  source {
    name          = materialize_source_mysql.test.name
    schema_name   = materialize_source_mysql.test.schema_name
    database_name = materialize_source_mysql.test.database_name
  }

  upstream_name        = "mysql_table1"
  upstream_schema_name = "shop"

  exclude_columns = ["banned"]
  comment         = "MySQL source table integration test"
}

# SQL Server Source Table
resource "materialize_source_table_sqlserver" "sqlserver_table_integration" {
  name          = "sqlserver_source_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  source {
    name          = materialize_source_sqlserver.sqlserver_source.name
    schema_name   = materialize_source_sqlserver.sqlserver_source.schema_name
    database_name = materialize_source_sqlserver.sqlserver_source.database_name
  }

  upstream_name        = "table1"
  upstream_schema_name = "dbo"

  exclude_columns = ["about"]
  comment         = "SQL Server source table integration test"
}

# Webhook Source Table
resource "materialize_source_table_webhook" "webhook_table_integration" {
  name          = "webhook_source_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  body_format = "json"

  include_headers {
    all = true
  }

  check_options {
    field {
      body = true
    }
    alias = "bytes"
  }

  comment = "Webhook source table integration test"
}
