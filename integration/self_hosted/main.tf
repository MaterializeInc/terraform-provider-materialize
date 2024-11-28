terraform {
  required_providers {
    materialize = {
      source = "materialize.com/devex/materialize"
    }
  }
}

provider "materialize" {
  host     = "materialized"
  port     = 6877
  database = "materialize"
  username = "mz_system"
  password = ""
  sslmode  = "disable"
}
