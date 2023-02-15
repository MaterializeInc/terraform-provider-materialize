resource "materialize_secret" "password" {
  name  = "password"
  value = "decode('c2VjcmV0Cg==', 'base64')"
}

resource "materialize_secret" "postgres_password" {
  name          = "pg_pass"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "decode('c2VjcmV0Cg==', 'base64')"
}

resource "materialize_secret" "kafka_password" {
  name          = "kafka_pass"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  value         = "decode('c2VjcmV0Cg==', 'base64')"
}