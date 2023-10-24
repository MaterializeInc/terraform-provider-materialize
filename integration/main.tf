terraform {
  required_providers {
    materialize = {
      source = "materialize.com/devex/materialize"
    }
  }
}

provider "materialize" {
  mz_host     = "materialized"
  mz_username = "mz_system"
  mz_password = "password"
  mz_port     = 6877
  mz_database = "materialize"
  mz_sslmode  = false
}
