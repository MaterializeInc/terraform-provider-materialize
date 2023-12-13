# Configuration-based authentication
provider "materialize" {
  password       = var.materialize_password # optionally use MZ_PASSWORD env var
  default_region = "aws/us-east-1" # optionally use MZ_REGION env var
}
