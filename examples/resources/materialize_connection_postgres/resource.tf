# Create a Postgres Connection
resource "materialize_connection_postgres" "example_postgres_connection" {
  name = "example_postgres_connection"
  host = "instance.foo000.us-west-1.rds.amazonaws.com"
  port = 5432
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
  database = "example"
}

# CREATE CONNECTION example_postgres_connection TO POSTGRES (
#     HOST 'instance.foo000.us-west-1.rds.amazonaws.com',
#     PORT 5432,
#     USER SECRET "database"."schema"."example"
#     PASSWORD SECRET "database"."schema"."example",
#     DATABASE 'example'
# );


# Create a Postgres Connection with SSH tunnel & plain text user
resource "materialize_connection_postgres" "example_postgres_connection" {
  name     = "example_postgres_connection"
  host     = "instance.foo000.us-west-1.rds.amazonaws.com"
  port     = 5432
  database = "example"

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

# CREATE CONNECTION example_postgres_connection TO POSTGRES (
#     HOST 'instance.foo000.us-west-1.rds.amazonaws.com',
#     PORT 5432,
#     USER "my_user",
#     PASSWORD SECRET "database"."schema"."example",
#     DATABASE 'example',
#     SSH TUNNEL "example"
# );
