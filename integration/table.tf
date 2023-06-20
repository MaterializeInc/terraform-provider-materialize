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

resource "materialize_grant_table" "table_grant_select" {
  role_name     = materialize_role.role_1.name
  privilege     = "SELECT"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

resource "materialize_grant_table" "table_grant_insert" {
  role_name     = materialize_role.role_1.name
  privilege     = "INSERT"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

resource "materialize_grant_table" "table_grant_update" {
  role_name     = materialize_role.role_2.name
  privilege     = "UPDATE"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

resource "materialize_grant_table" "table_grant_delete" {
  role_name     = materialize_role.role_2.name
  privilege     = "DELETE"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

output "qualified_table" {
  value = materialize_table.simple_table.qualified_sql_name
}

data "materialize_table" "all" {}
