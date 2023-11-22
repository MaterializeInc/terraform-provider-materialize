resource "materialize_type" "row_type" {
  name          = "int4_row"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  row_properties {
    field_name  = "a"
    field_type = "int4"
  }
  row_properties {
    field_name  = "b"
    field_type = "text"
  }
}

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

resource "materialize_type_grant" "type_grant_usage" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  type_name     = materialize_type.list_type.name
  schema_name   = materialize_type.list_type.schema_name
  database_name = materialize_type.list_type.database_name
}

resource "materialize_type_grant_default_privilege" "type_grant_default_privilege" {
  grantee_name     = materialize_role.role_1.name
  privilege        = "USAGE"
  target_role_name = materialize_role.role_2.name
}

resource "materialize_type_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
}

output "qualified_type" {
  value = materialize_type.list_type.qualified_sql_name
}

data "materialize_type" "all" {}