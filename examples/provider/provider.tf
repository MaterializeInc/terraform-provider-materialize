# Configuration-based authentication
provider "materialize" {
  host     = var.materialize_hostname # optionally use MZ_HOST env var
  username = var.materialize_username # optionally use MZ_USER env var
  password = var.materialize_password # optionally use MZ_PW env var
  port     = var.materialize_port     # optionally use MZ_PORT env var
  database = var.materialize_database # optionally use MZ_DATABASE env var
}
