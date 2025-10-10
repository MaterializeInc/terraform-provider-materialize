resource "materialize_secret" "sqlserver_password" {
  name    = "sqlserver_password"
  value   = "some-secret-value"
  comment = "secret comment"
}

# Basic SQL Server connection
resource "materialize_connection_sqlserver" "basic" {
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
  validate = true
}

# SQL Server connection with SSH tunnel
resource "materialize_connection_ssh_tunnel" "example_ssh_connection" {
  name = "ssh_connection"
  host = "bastion-host.example.com"
  port = 22
  user = "ubuntu"
}

resource "materialize_connection_sqlserver" "ssh_example" {
  name = "sqlserver_ssh_connection"
  host = "private-sql-server.example.com"
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

  ssh_tunnel {
    name = materialize_connection_ssh_tunnel.example_ssh_connection.name
  }

  validate = false
}

# SQL Server connection with AWS PrivateLink
resource "materialize_connection_aws_privatelink" "sqlserver_privatelink" {
  name               = "sqlserver_privatelink"
  service_name       = "com.amazonaws.vpce.us-east-1.vpce-svc-0e123abc123198abc"
  availability_zones = ["use1-az1", "use1-az4"]
}

resource "materialize_connection_sqlserver" "privatelink_example" {
  name = "sqlserver_privatelink_connection"
  host = "sqlserver.example.com"
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

  aws_privatelink {
    name = materialize_connection_aws_privatelink.sqlserver_privatelink.name
  }

  validate = false
}
