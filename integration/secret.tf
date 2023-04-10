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

output "qualified_sql_password" {
  value = materialize_secret.password.qualified_sql_name
}

output "qualified_kafka_password" {
  value = materialize_secret.kafka_password.qualified_sql_name
}

data "materialize_secret" "all" {}
