resource "materialize_secret" "sqlserver_password" {
  name  = "sqlserver_password"
  value = base64encode("c2VjcmV0Cg==")
}

resource "materialize_connection_sqlserver" "sqlserver_connection" {
  name = "sqlserver_connection"
  host = "sql-server.example.com"
  port = 1433

  user {
    text = "sqluser"
  }

  password {
    name          = materialize_secret.sqlserver_password.name
    schema_name   = materialize_secret.sqlserver_password.schema_name
    database_name = materialize_secret.sqlserver_password.database_name
  }

  database = "testdb"
}

# Basic SQL Server source for specific tables
resource "materialize_source_sqlserver" "example" {
  name         = "sqlserver_source"
  cluster_name = "quickstart"

  sqlserver_connection {
    name          = materialize_connection_sqlserver.sqlserver_connection.name
    schema_name   = materialize_connection_sqlserver.sqlserver_connection.schema_name
    database_name = materialize_connection_sqlserver.sqlserver_connection.database_name
  }

  table {
    upstream_name = "dbo.customers"
    name          = "customers"
  }

  table {
    upstream_name = "dbo.orders"
    name          = "orders"
  }

  table {
    upstream_name = "dbo.products"
    name          = "products"
  }
}

# SQL Server source for all tables
resource "materialize_source_sqlserver" "all_tables" {
  name         = "sqlserver_source_all"
  cluster_name = "quickstart"

  sqlserver_connection {
    name          = materialize_connection_sqlserver.sqlserver_connection.name
    schema_name   = materialize_connection_sqlserver.sqlserver_connection.schema_name
    database_name = materialize_connection_sqlserver.sqlserver_connection.database_name
  }

  # No table blocks specified means all tables will be included
}

# SQL Server source with text columns and excluded columns
resource "materialize_source_sqlserver" "with_options" {
  name         = "sqlserver_source_with_options"
  cluster_name = "quickstart"

  sqlserver_connection {
    name          = materialize_connection_sqlserver.sqlserver_connection.name
    schema_name   = materialize_connection_sqlserver.sqlserver_connection.schema_name
    database_name = materialize_connection_sqlserver.sqlserver_connection.database_name
  }

  table {
    upstream_name = "dbo.users"
    name          = "users"
  }

  table {
    upstream_name = "dbo.posts"
    name          = "posts"
  }

  table {
    upstream_name = "dbo.comments"
    name          = "comments"
  }

  # Convert unsupported data types to text
  text_columns = ["dbo.users.description", "dbo.posts.content", "dbo.comments.metadata"]

  # Exclude problematic columns
  exclude_columns = ["dbo.users.image_data", "dbo.posts.binary_data"]
}
