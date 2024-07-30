resource "materialize_secret" "password" {
  name    = "password"
  value   = "c2VjcmV0Cg=="
  comment = "secret comment"
}

# Create in separate region
resource "materialize_secret" "password_us_west" {
  name    = "password"
  value   = "c2VjcmV0Cg=="
  comment = "secret comment"
  region  = "aws/us-west-2"
}

resource "materialize_secret" "postgres_password" {
  name          = "pg_pass"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "c2VjcmV0Cg=="
}

# Create in separate region
resource "materialize_secret" "postgres_password_us_west" {
  name          = "pg_pass"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  value         = "c2VjcmV0Cg=="
  region        = "aws/us-west-2"
}

resource "materialize_secret" "mysql_password" {
  name          = "mysql__pass"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "c2VjcmV0Cg=="
}

# Create in separate region
resource "materialize_secret" "mysql_password_us_west" {
  name          = "mysql__pass"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  value         = "c2VjcmV0Cg=="
  region        = "aws/us-west-2"
}

resource "materialize_secret" "mysql_user" {
  name          = "mysql_user"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "repluser"
}

# Create in separate region
resource "materialize_secret" "mysql_user_us_west" {
  name          = "mysql_user"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  value         = "repluser"
  region        = "aws/us-west-2"
}

resource "materialize_secret" "kafka_password" {
  name          = "kafka_pass"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "c2VjcmV0Cg=="
}

# Create in separate region
resource "materialize_secret" "kafka_password_us_west" {
  name          = "kafka_pass"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  value         = "c2VjcmV0Cg=="
  region        = "aws/us-west-2"
}

resource "materialize_secret" "aws_password" {
  name          = "aws_password"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "test"
}

# Create in separate region
resource "materialize_secret" "aws_password_us_west" {
  name          = "aws_password"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  value         = "test"
  region        = "aws/us-west-2"
}

resource "materialize_secret_grant" "secret_grant_usage" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  secret_name   = materialize_secret.password.name
  schema_name   = materialize_secret.password.schema_name
  database_name = materialize_secret.password.database_name
}

# Create in separate region
resource "materialize_secret_grant" "secret_grant_usage_us_west" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  secret_name   = materialize_secret.password_us_west.name
  schema_name   = materialize_secret.password_us_west.schema_name
  database_name = materialize_secret.password_us_west.database_name
  region        = "aws/us-west-2"
}

resource "materialize_secret_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
}

# Create in separate region
resource "materialize_secret_grant_default_privilege" "example_us_west" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target_us_west.name
  schema_name      = materialize_schema.schema_us_west.name
  database_name    = materialize_database.database_us_west.name
  region           = "aws/us-west-2"
}

output "qualified_sql_password" {
  value = materialize_secret.password.qualified_sql_name
}

output "qualified_kafka_password" {
  value = materialize_secret.kafka_password.qualified_sql_name
}

data "materialize_secret" "all" {}
