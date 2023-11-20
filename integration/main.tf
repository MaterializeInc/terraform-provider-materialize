terraform {
  required_providers {
    materialize = {
      source = "materialize.com/devex/materialize"
    }
  }
}

provider "materialize" {
  host     = "materialized"
  user     = "mz_system"
  password = "password"
  port     = 6877
  database = "materialize"
  sslmode  = "disable"
}
