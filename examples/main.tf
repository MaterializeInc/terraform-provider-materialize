terraform {
  required_providers {
    materialize = {
      version = "0.1.0"
      # Local reference of provider binary
      source = "materialize.com/devex/materialize"
    }
  }
}

provider "materialize" {
  host     = local.host
  username = local.username
  password = local.password
  port     = local.port
  database = local.database
}

data "materialize_secret" "all" {}

resource "materialize_schema" "example_schema" {
  name          = "example"
  database_name = "materialize"
}

# Create a cluster and attach two 2xsmall cluster replicas
resource "materialize_cluster" "example_cluster" {
  name = "example"
}

resource "materialize_cluster_replica" "example_1_cluster_replica" {
  name         = "example_1"
  cluster_name = materialize_cluster.example_cluster.name
  size         = "2xsmall"
}

resource "materialize_cluster_replica" "example_2_cluster_replica" {
  name         = "example_2"
  cluster_name = materialize_cluster.example_cluster.name
  size         = "2xsmall"
}

# Create a load generator source
resource "materialize_source" "example_source_load_generator" {
  name                = "load_gen_example"
  schema_name         = materialize_schema.example_schema.name
  size                = "3xsmall"
  connection_type     = "LOAD GENERATOR"
  load_generator_type = "COUNTER"
  tick_interval       = "500ms"
  scale_factor        = 0.01
}

# Create a secret
resource "materialize_secret" "example_secret" {
  name        = "example"
  schema_name = materialize_schema.example_schema.name
  value       = "some-secret-value"
}


# Create SSH Connection
resource "materialize_connection" "example_ssh_connection" {
  name            = "ssh_example_connection"
  schema_name     = "public"
  connection_type = "SSH TUNNEL"
  ssh_host        = "example.com"
  ssh_port        = 22
  ssh_user        = "example"
}

# # Create a AWS Private Connection
# Note: you need the max_aws_privatelink_connections increased for this to work:
# show max_aws_privatelink_connections;
# resource "materialize_connection" "example_privatelink_connection" {
#   name            = "example_privatelink_connection"
#   schema_name     = "public"
#   connection_type = "AWS PRIVATELINK"
#   aws_privatelink_service_name = "com.amazonaws.us-east-1.materialize.example"
#   aws_privatelink_availability_zones = ["use1-az2", "use1-az6"]
# }

# Create a Postgres Connection
resource "materialize_connection" "example_postgres_connection" {
  name              = "example_postgres_connection"
  connection_type   = "POSTGRES"
  postgres_host     = "instance.foo000.us-west-1.rds.amazonaws.com"
  postgres_port     = 5432
  postgres_user     = "example"
  postgres_password = "example"
  postgres_database = "example"
}

# Create a Kafka Connection
resource "materialize_connection" "example_kafka_connection" {
  name            = "example_kafka_connection"
  connection_type = "KAFKA"
  # kafka_broker    = "example.com:9092"
  kafka_brokers         = ["example.com:9092", "example.com:9093"]
  kafka_sasl_username   = "example"
  kafka_sasl_password   = "kafka_password"
  kafka_sasl_mechanisms = "SCRAM-SHA-256"
  kafka_progress_topic  = "example"
}

# Create a Confluent Schema Registry Connection
resource "materialize_connection" "example_confluent_schema_registry_connection" {
  name                               = "example_confluent_schema_registry_connection"
  connection_type                    = "CONFLUENT SCHEMA REGISTRY"
  confluent_schema_registry_url      = "https://example.com"
  confluent_schema_registry_password = "example"
  confluent_schema_registry_username = "example"
}
