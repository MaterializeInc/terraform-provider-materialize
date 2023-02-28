resource "materialize_source" "example_source_postgres" {
  name                = "source_postgres"
  schema_name         = "schema"
  size                = "3xsmall"
  postgres_connection = "pg_connection"
  publication         = "mz_source"
  tables = {
    "schema1.table_1" = "s1_table_1"
    "schema2_table_1" = "s2_table_1"
  }
}

# CREATE SOURCE schema.source_postgres
#   FROM POSTGRES CONNECTION pg_connection (PUBLICATION 'mz_source')
#   FOR TABLES (schema1.table_1 AS s1_table_1, schema2_table_1 AS s2_table_1)
#   WITH (SIZE = '3xsmall');