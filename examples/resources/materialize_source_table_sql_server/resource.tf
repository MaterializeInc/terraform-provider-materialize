resource "materialize_connection_sql_server" "sql_server_connection" {
  name = "sql_server_connection"
  host = "sqlserver.example.com"
  port = 1433
  database = "mydb"
  ssl_mode = "require"
  user {
    text = "sa"
  }
  password {
    name = materialize_secret.sql_server_password.name
  }
}

resource "materialize_secret" "sql_server_password" {
  name  = "sql_server_password"
  value = "your-password-here"
}

resource "materialize_source_sql_server" "example_source_sql_server" {
  name         = "source_sql_server"
  schema_name  = "schema"
  cluster_name = "quickstart"

  sql_server_connection {
    name          = materialize_connection_sql_server.sql_server_connection.name
    database_name = materialize_connection_sql_server.sql_server_connection.database_name
    schema_name   = materialize_connection_sql_server.sql_server_connection.schema_name
  }

  table {
    upstream_name        = "table1"
    upstream_schema_name = "dbo"
    name                 = "s1_table1"
  }

  table {
    upstream_name        = "table2"
    upstream_schema_name = "dbo"
    name                 = "s2_table2"
  }
}

resource "materialize_source_table_sql_server" "sql_server_table" {
  name          = "sql_server_table"
  schema_name   = "public"
  database_name = "materialize"

  source {
    name          = materialize_source_sql_server.example_source_sql_server.name
    schema_name   = materialize_source_sql_server.example_source_sql_server.schema_name
    database_name = materialize_source_sql_server.example_source_sql_server.database_name
  }

  upstream_name        = "another_table"
  upstream_schema_name = "dbo"

  text_columns = [
    "text_column"
  ]
}

# CREATE SOURCE schema.source_sql_server
#   IN CLUSTER quickstart
#   FROM SQL SERVER CONNECTION "database"."schema"."sql_server_connection"
#   FOR TABLES (dbo.table1 AS s1_table1, dbo.table2 AS s2_table2);
