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
    upstream_name        = "table1"
    upstream_schema_name = "schema1"
    name                 = "s1_table1"
  }

  table {
    upstream_name        = "table2"
    upstream_schema_name = "schema2"
    name                 = "s2_table2"
  }
}

# CREATE SOURCE schema.source_postgres
#   FROM POSTGRES CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source')
#   FOR TABLES (schema1.table1 AS s1_table1, schema2.table2 AS s2_table2);
