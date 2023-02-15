# Configuration-based authentication
provider "materialize" {
  host     = local.host
  username = local.username
  password = local.password
  port     = local.port
  database = local.database
}
