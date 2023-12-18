# Create a AWS Connection
resource "materialize_connection_aws" "example_connection" {
  name              = "example_connection"
  schema_name       = "public"
  access_key_id     = "foo"
  secret_access_key = "bar"
}

# CREATE CONNECTION example_connection TO AWS WITH (
#     ACCESS_KEY_ID = 'foo',
#     SECRET_ACCESS_KEY = 'bar'
# );
