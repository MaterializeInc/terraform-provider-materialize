resource "materialize_connection" "ssh_connection" {
  name            = "ssh_connection"
  connection_type = "SSH TUNNEL"
  ssh_host        = "example.com"
  ssh_port        = 22
  ssh_user        = "example"
}

resource "materialize_connection" "kafka_connection" {
  name            = "kafka_connection"
  connection_type = "KAFKA"
  kafka_broker    = "kafka:9092"
}

resource "materialize_connection" "schema_registry" {
  name                          = "schema_registry_connection"
  connection_type               = "CONFLUENT SCHEMA REGISTRY"
  confluent_schema_registry_url = "http://schema-registry:8081"
}

output "qualified_ssh_connection" {
  value = materialize_connection.ssh_connection.qualified_name
}