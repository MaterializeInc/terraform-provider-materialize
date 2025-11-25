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

# PostgreSQL source with text columns and excluded columns
resource "materialize_source_postgres" "with_options" {
  name         = "source_postgres_with_options"
  schema_name  = "schema"
  cluster_name = "quickstart"
  publication  = "mz_source"

  postgres_connection {
    name = "pg_connection"
  }

  table {
    upstream_name        = "users"
    upstream_schema_name = "public"
    name                 = "users"
  }

  table {
    upstream_name        = "posts"
    upstream_schema_name = "public"
    name                 = "posts"
  }

  # Convert unsupported data types to text
  text_columns = ["public.users.description", "public.posts.content"]

  # Exclude problematic columns
  exclude_columns = ["public.users.image_data", "public.posts.binary_data"]
}

# CREATE SOURCE schema.source_postgres_with_options
#   FROM POSTGRES CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source', TEXT COLUMNS (public.users.description, public.posts.content), EXCLUDE COLUMNS (public.users.image_data, public.posts.binary_data))
#   FOR TABLES (public.users AS users, public.posts AS posts);
