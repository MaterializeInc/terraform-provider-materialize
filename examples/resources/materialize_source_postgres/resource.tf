resource "materialize_source_postgres" "example_source_postgres" {
  name         = "source_postgres"
  schema_name  = "schema"
  cluster_name = "quickstart"
  publication  = "mz_source"

  postgres_connection {
    name = "pg_connection"
    # Optional parameters
    # database_name = "postgres"
    # schema_name = "public"
  }

  table {
    name  = "schema1.table_1"
    alias = "s1_table_1"
  }

  table {
    name  = "schema2.table_1"
    alias = "s2_table_1"
  }
}

# CREATE SOURCE schema.source_postgres
#   FROM POSTGRES CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source')
#   FOR TABLES (schema1.table_1 AS s1_table_1, schema2_table_1 AS s2_table_1)
#   WITH (SIZE = '3xsmall');
