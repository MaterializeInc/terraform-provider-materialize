resource "materialize_source_table_sqlserver" "sqlserver_table_from_source" {
  name          = "sqlserver_table_from_source"
  schema_name   = "public"
  database_name = "materialize"

  source {
    name          = materialize_source_sqlserver.example.name
    schema_name   = materialize_source_sqlserver.example.schema_name
    database_name = materialize_source_sqlserver.example.database_name
  }

  upstream_name        = "sqlserver_table_name" # The name of the table in the SQL Server database
  upstream_schema_name = "dbo"                  # The schema of the table in the SQL Server database (typically "dbo")

  text_columns    = ["updated_at"]
  exclude_columns = ["id"]
}
