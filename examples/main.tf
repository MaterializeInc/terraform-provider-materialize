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

resource "materialize_schema" "example_schema" {
  name          = "example"
  database_name = "materialize"
}

# Create a secret
resource "materialize_secret" "example_secret" {
  name        = "example"
  schema_name = materialize_schema.example_schema.name
  value       = "decode('c2VjcmV0Cg==', 'base64')"
}