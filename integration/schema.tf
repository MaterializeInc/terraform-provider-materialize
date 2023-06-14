resource "materialize_schema" "schema" {
  name          = "example_schema"
  database_name = materialize_database.database.name
}

output "qualified_schema" {
  value = materialize_schema.schema.qualified_sql_name
}

data "materialize_schema" "all" {}
