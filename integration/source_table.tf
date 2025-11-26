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
