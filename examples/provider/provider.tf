# Configuration-based authentication
provider "materialize" {
  password       = var.materialize_password # optionally use MZ_PASSWORD env var
  default_region = "aws/us-east-1"          # optionally use MZ_REGION env var
}

# Self-hosted Materialize authentication
provider "materialize" {
  host     = "materialized" # optionally use MZ_HOST env var
  port     = 6877           # optionally use MZ_PORT env var
  username = "mz_system"    # optionally use MZ_USER env var
  database = "materialize"  # optionally use MZ_DATABASE env var
  password = ""             # optionally use MZ_PASSWORD env var
  sslmode  = "disable"      # optionally use MZ_SSLMODE env var
}
