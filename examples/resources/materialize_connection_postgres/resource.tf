# Create a Postgres Connection
resource "materialize_connection_postgres" "example_postgres_connection" {
  name     = "example_postgres_connection"
  host     = "instance.foo000.us-west-1.rds.amazonaws.com"
  port     = 5432
  user     = "example"
  password = "example"
  database = "example"
}

# CREATE CONNECTION example_postgres_connection TO POSTGRES (
#     HOST 'instance.foo000.us-west-1.rds.amazonaws.com',
#     PORT 5432,
#     USER 'example',
#     PASSWORD SECRET example,
#     DATABASE 'example'
# );
