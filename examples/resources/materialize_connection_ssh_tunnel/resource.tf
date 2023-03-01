# Create SSH Connection
resource "materialize_connection_ssh_tunnel" "example_ssh_connection" {
  name            = "ssh_example_connection"
  schema_name     = "public"
  host        = "example.com"
  port        = 22
  user        = "example"
}

# CREATE CONNECTION ssh_example_connection TO SSH TUNNEL (
#    HOST 'example.com',
#    PORT 22,
#    USER 'example'
# );
