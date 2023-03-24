resource "materialize_table" "simple_table" {
  name          = "simple_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  column {
    name = "column_1"
    type = "text"
  }
  column {
    name = "column_2"
    type = "int"
  }
  column {
    name     = "column_3"
    type     = "text"
    nullable = true
  }

}

output "qualified_table" {
  value = materialize_table.simple_table.qualified_sql_name
}
