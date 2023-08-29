resource "materialize_secret" "password" {
  name  = "password"
  value = "c2VjcmV0Cg=="
}

resource "materialize_secret" "postgres_password" {
  name          = "pg_pass"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "c2VjcmV0Cg=="
}

resource "materialize_secret" "kafka_password" {
  name          = "kafka_pass"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "c2VjcmV0Cg=="
}

resource "materialize_secret_grant" "secret_grant_usage" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  secret_name   = materialize_secret.password.name
  schema_name   = materialize_secret.password.schema_name
  database_name = materialize_secret.password.database_name
}

resource "materialize_secret_grant_default_privilege" "example" {
  grantee_name     = materialize_role.role_1.name
  privilege        = "USAGE"
  target_role_name = materialize_role.role_2.name
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
}

output "qualified_sql_password" {
  value = materialize_secret.password.qualified_sql_name
}

output "qualified_kafka_password" {
  value = materialize_secret.kafka_password.qualified_sql_name
}

data "materialize_secret" "all" {}
