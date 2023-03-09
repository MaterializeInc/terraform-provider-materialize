resource "materialize_table" "simple_table" {
  name          = "simple_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  columns {
    col_name = "column_1"
    col_type = "text"
  }
  columns {
    col_name = "column_2"
    col_type = "int"
  }
  columns {
    col_name = "column_3"
    col_type = "text"
    not_null = true
  }

}