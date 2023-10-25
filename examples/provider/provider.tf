# Configuration-based authentication
provider "materialize" {
  host     = var.materialize_host     # optionally use MZ_HOST env var
  user     = var.materialize_user     # optionally use MZ_USER env var
  password = var.materialize_password # optionally use MZ_PASSWORD env var
  port     = var.materialize_port     # optionally use MZ_PORT env var
  database = var.materialize_database # optionally use MZ_DATABASE env var
}
