# Create a Confluent Schema Registry Connection
resource "materialize_connection_confluent_schema_registry" "example_confluent_schema_registry_connection" {
  name     = "example_csr_connection"
  url      = "https://rp-f00000bar.data.vectorized.cloud:30993"
  password = "example"
  username {
    text = "example"
  }
}

# CREATE CONNECTION example_csr_connection TO CONFLUENT SCHEMA REGISTRY (
#     URL 'https://rp-f00000bar.data.vectorized.cloud:30993',
#     USERNAME = 'example',
#     PASSWORD = SECRET example
# );
