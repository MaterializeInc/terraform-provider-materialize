terraform {
  required_providers {
    materialize = {
      source = "materialize.com/devex/materialize"
    }
  }
}

provider "materialize" {
  endpoint       = "http://frontegg:3000"
  cloud_endpoint = "http://cloud:3001"
  password       = "mzp_1b2a3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b"
  database       = "materialize"
  sslmode        = "disable"
}
