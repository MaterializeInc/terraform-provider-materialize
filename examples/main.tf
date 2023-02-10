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

resource "materialize_cluster" "example_cluster" {
  name = "terraform"
}

resource "materialize_cluster_replica" "example_cluster_replica" {
  name         = "c1"
  cluster_name = "terraform"
  size         = "2xsmall"
}