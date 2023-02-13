terraform {
  required_providers {
    materialize = {
      source = "materialize.com/devex/materialize"
    }
  }
}

provider "materialize" {
  host     = "materialized"
  username = "materialize"
  password = "password"
  port     = 6875
  database = "materialize"
  testing  = true
}