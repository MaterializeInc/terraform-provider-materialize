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
