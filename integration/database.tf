resource "materialize_database" "database" {
  name    = "example_database"
  comment = "database comment"
}

# Create in separate region
resource "materialize_database" "database_us_west" {
  name    = "example_database"
  comment = "database comment"
  region  = "aws/us-west-2"
}

resource "materialize_database_grant" "database_grant_usage" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  database_name = materialize_database.database.name
}

# Create in separate region
resource "materialize_database_grant" "database_grant_usage_us_west" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  database_name = materialize_database.database.name
  region        = "aws/us-west-2"
}

resource "materialize_database_grant" "database_grant_create" {
  role_name     = materialize_role.role_2.name
  privilege     = "CREATE"
  database_name = materialize_database.database.name
}

# Create in separate region
resource "materialize_database_grant" "database_grant_create_us_west" {
  role_name     = materialize_role.role_2_us_west.name
  privilege     = "CREATE"
  database_name = materialize_database.database_us_west.name
  region        = "aws/us-west-2"
}

resource "materialize_database_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
}

# Create in separate region
resource "materialize_database_grant_default_privilege" "example_us_west" {
  grantee_name     = materialize_role.grantee_us_west.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target_us_west.name
  region           = "aws/us-west-2"
}

data "materialize_database" "all" {}

data "materialize_current_database" "default" {}
