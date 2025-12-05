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
