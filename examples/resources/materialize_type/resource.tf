resource "materialize_type" "list_type" {
  name          = "int4_list"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  list_properties {
    element_type = "int4"
  }
}

resource "materialize_type" "map_type" {
  name          = "int4_map"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  map_properties {
    key_type   = "text"
    value_type = "int4"
  }
}
