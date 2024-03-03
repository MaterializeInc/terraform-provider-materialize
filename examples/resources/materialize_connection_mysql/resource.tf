# Create a MySQL Connection
resource "materialize_connection_mysql" "example_mysql_connection" {
  name = "example_mysql_connection"
  host = "instance.foo000.us-west-1.rds.amazonaws.com"
  port = 3306
  user {
    secret {
      name          = "example"
      database_name = "database"
      schema_name   = "schema"
    }
  }
  password {
    name          = "example"
    database_name = "database"
    schema_name   = "schema"
  }
}

# CREATE CONNECTION example_mysql_connection TO MYSQL (
#     HOST 'instance.foo000.us-west-1.rds.amazonaws.com',
#     PORT 3306,
#     USER SECRET "database"."schema"."example"
#     PASSWORD SECRET "database"."schema"."example",
# );


# Create a MySQL Connection with SSH tunnel & plain text user
resource "materialize_connection_mysql" "example_mysql_connection" {
  name = "example_mysql_connection"
  host = "instance.foo000.us-west-1.rds.amazonaws.com"
  port = 3306

  user {
    text = "my_user"
  }
  password {
    name          = "example"
    database_name = "database"
    schema_name   = "schema"
  }
  ssh_tunnel {
    name = "example"
  }
}

# CREATE CONNECTION example_mysql_connection TO POSTGRES (
#     HOST 'instance.foo000.us-west-1.rds.amazonaws.com',
#     PORT 3306,
#     USER "my_user",
#     PASSWORD SECRET "database"."schema"."example",
#     SSH TUNNEL "example"
# );
