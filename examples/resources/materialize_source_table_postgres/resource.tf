resource "materialize_source_table_postgres" "postgres_table_from_source" {
  name          = "postgres_table_from_source"
  schema_name   = "public"
  database_name = "materialize"

  source {
    name          = materialize_source_postgres.example.name
    schema_name   = materialize_source_postgres.example.schema_name
    database_name = materialize_source_postgres.example.database_name
  }

  upstream_name        = "postgres_table_name"  # The name of the table in the postgres database
  upstream_schema_name = "postgres_schema_name" # The name of the database in the postgres database

  text_columns = [
    "updated_at"
  ]

}
