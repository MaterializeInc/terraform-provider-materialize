resource "materialize_schema" "public" {
  name          = "public"
  database_name = materialize_database.database.name
  comment       = "public schema comment"
}

# Create in separate region
resource "materialize_schema" "public_us_west" {
  name          = "public"
  database_name = materialize_database.database.name
  comment       = "public schema comment"
  region        = "aws/us-west-2"
}

resource "materialize_schema" "db1_public" {
  name          = "public"
  database_name = materialize_database.db1.name
  comment       = "public schema comment"
}

# Create in separate region
resource "materialize_schema" "db1_public_us_west" {
  name          = "public"
  database_name = materialize_database.db1.name
  comment       = "public schema comment"
  region        = "aws/us-west-2"
}

resource "materialize_schema" "db2_public" {
  name          = "public"
  database_name = materialize_database.db2.name
  comment       = "public schema comment"
}

# Create in separate region
resource "materialize_schema" "db2_public_us_west" {
  name          = "public"
  database_name = materialize_database.db2.name
  comment       = "public schema comment"
  region        = "aws/us-west-2"
}

resource "materialize_schema" "schema" {
  name          = "example_schema"
  database_name = materialize_database.database.name
  comment       = "schema comment"
}

# Create in separate region
resource "materialize_schema" "schema_us_west" {
  name          = "example_schema"
  database_name = materialize_database.database.name
  comment       = "schema comment"
  region        = "aws/us-west-2"
}

resource "materialize_schema_grant" "schema_grant_usage" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  database_name = materialize_schema.schema.database_name
  schema_name   = materialize_schema.schema.name
}

# Create in separate region
resource "materialize_schema_grant" "schema_grant_usage_us_west" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  database_name = materialize_schema.schema.database_name
  schema_name   = materialize_schema.schema.name
  region        = "aws/us-west-2"
}

resource "materialize_schema_grant" "schema_grant_create" {
  role_name     = materialize_role.role_2.name
  privilege     = "CREATE"
  database_name = materialize_schema.schema.database_name
  schema_name   = materialize_schema.schema.name
}

# Create in separate region
resource "materialize_schema_grant" "schema_grant_create_us_west" {
  role_name     = materialize_role.role_2.name
  privilege     = "CREATE"
  database_name = materialize_schema.schema.database_name
  schema_name   = materialize_schema.schema.name
  region        = "aws/us-west-2"
}

resource "materialize_schema_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
}

# Create in separate region
resource "materialize_schema_grant_default_privilege" "example_us_west" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
  database_name    = materialize_database.database.name
  region           = "aws/us-west-2"
}

output "qualified_schema" {
  value = materialize_schema.schema.qualified_sql_name
}

data "materialize_schema" "all" {}
