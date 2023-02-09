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
  name                = "example"
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
  value       = "decode('c2VjcmV0Cg==', 'base64')"
}