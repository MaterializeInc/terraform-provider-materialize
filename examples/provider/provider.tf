# Configuration-based authentication
provider "materialize" {
  mz_host     = var.materialize_hostname # optionally use MZ_HOST env var
  mz_username = var.materialize_username # optionally use MZ_USERNAME env var
  mz_password = var.materialize_password # optionally use MZ_PASSWORD env var
  mz_port     = var.materialize_port     # optionally use MZ_PORT env var
  mz_database = var.materialize_database # optionally use MZ_DATABASE env var
}
