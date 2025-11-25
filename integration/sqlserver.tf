# SQL Server Connection
resource "materialize_secret" "sqlserver_password" {
  name  = "sqlserver_password"
  value = "Password123!"
}

resource "materialize_connection_sqlserver" "sqlserver_connection" {
  name = "sqlserver_connection"
  host = "sqlserver"
  port = 1433

  user {
    text = "sa"
  }

  password {
    name          = materialize_secret.sqlserver_password.name
    schema_name   = materialize_secret.sqlserver_password.schema_name
    database_name = materialize_secret.sqlserver_password.database_name
  }

  database = "testdb"
  validate = false
}

# SQL Server Source for specific tables
resource "materialize_source_sqlserver" "sqlserver_source" {
  name         = "sqlserver_source"
  cluster_name = "quickstart"

  sqlserver_connection {
    name          = materialize_connection_sqlserver.sqlserver_connection.name
    schema_name   = materialize_connection_sqlserver.sqlserver_connection.schema_name
    database_name = materialize_connection_sqlserver.sqlserver_connection.database_name
  }

  table {
    upstream_name        = "table1"
    upstream_schema_name = "dbo"
    name                 = "sqlserver_table1"
  }

  exclude_columns = ["dbo.table1.about"]
}

# SQL Server Source for all tables
resource "materialize_source_sqlserver" "sqlserver_source_all" {
  name         = "sqlserver_source_all"
  cluster_name = "quickstart"

  sqlserver_connection {
    name          = materialize_connection_sqlserver.sqlserver_connection.name
    schema_name   = materialize_connection_sqlserver.sqlserver_connection.schema_name
    database_name = materialize_connection_sqlserver.sqlserver_connection.database_name
  }

  exclude_columns = ["dbo.table3.data", "dbo.table1.about", "dbo.table2.about", "dbo.table5.large_text", "dbo.table5.image_data", "dbo.table5.xml_data", "dbo.table5.json_data", "dbo.table10.text_col", "dbo.table10.nvarchar_max"]
}

# SQL Server Source Table
resource "materialize_source_table_sqlserver" "sqlserver_table" {
  name          = "sqlserver_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  source {
    name          = materialize_source_sqlserver.sqlserver_source.name
    schema_name   = materialize_source_sqlserver.sqlserver_source.schema_name
    database_name = materialize_source_sqlserver.sqlserver_source.database_name
  }

  upstream_name        = "table1"
  upstream_schema_name = "dbo"

  exclude_columns = ["id", "about"]
}
