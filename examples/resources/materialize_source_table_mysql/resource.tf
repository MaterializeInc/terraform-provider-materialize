resource "materialize_source_table_mysql" "mysql_table_from_source" {
  name          = "mysql_table_from_source"
  schema_name   = "public"
  database_name = "materialize"

  source {
    name          = materialize_source_mysql.example.name
    schema_name   = materialize_source_mysql.example.schema_name
    database_name = materialize_source_mysql.example.database_name
  }

  upstream_name        = "mysql_table_name" # The name of the table in the MySQL database
  upstream_schema_name = "mysql_db_name"    # The name of the database in the MySQL database

  text_columns = [
    "updated_at"
  ]

  ignore_columns = ["about"]
}
