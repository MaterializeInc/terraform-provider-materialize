resource "materialize_type" "row_type" {
  name          = "row_type"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  row_properties {
    field_name = "a"
    field_type = "int4"
  }
  row_properties {
    field_name = "b"
    field_type = "text"
  }
}

# Create in separate region
resource "materialize_type" "row_type_us_west" {
  name          = "row_type"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  region        = "aws/us-west-2"

  row_properties {
    field_name = "a"
    field_type = "int4"
  }
  row_properties {
    field_name = "b"
    field_type = "text"
  }
}

resource "materialize_type" "row_nested_type" {
  name          = "nested_row_type"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  row_properties {
    field_name = "a"
    field_type = materialize_type.row_type.qualified_sql_name
  }
  row_properties {
    field_name = "b"
    field_type = "float8"
  }
}

# Create in separate region
resource "materialize_type" "row_nested_type_us_west" {
  name          = "nested_row_type"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  region        = "aws/us-west-2"

  row_properties {
    field_name = "a"
    field_type = materialize_type.row_type_us_west.qualified_sql_name
  }
  row_properties {
    field_name = "b"
    field_type = "float8"
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

# Create in separate region
resource "materialize_type" "list_type_us_west" {
  name          = "int4_list"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  region        = "aws/us-west-2"

  list_properties {
    element_type = "int4"
  }
}

resource "materialize_type" "list_nested_type" {
  name          = "int4_nested_list"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  list_properties {
    element_type = materialize_type.list_type.qualified_sql_name
  }
}

# Create in separate region
resource "materialize_type" "list_nested_type_us_west" {
  name          = "int4_nested_list"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  region        = "aws/us-west-2"

  list_properties {
    element_type = materialize_type.list_type_us_west.qualified_sql_name
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

# Create in separate region
resource "materialize_type" "map_type_us_west" {
  name          = "int4_map"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  region        = "aws/us-west-2"

  map_properties {
    key_type   = "text"
    value_type = "int4"
  }
}

resource "materialize_type" "map_nested_type" {
  name          = "int4_nested_map"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name

  map_properties {
    key_type   = "text"
    value_type = materialize_type.map_type.qualified_sql_name
  }
}

# Create in separate region
resource "materialize_type" "map_nested_type_us_west" {
  name          = "int4_nested_map"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  region        = "aws/us-west-2"

  map_properties {
    key_type   = "text"
    value_type = materialize_type.map_type_us_west.qualified_sql_name
  }
}

resource "materialize_type_grant" "type_grant_usage" {
  role_name     = materialize_role.role_1.name
  privilege     = "USAGE"
  type_name     = materialize_type.list_type.name
  schema_name   = materialize_type.list_type.schema_name
  database_name = materialize_type.list_type.database_name
}

# Create in separate region
resource "materialize_type_grant" "type_grant_usage_us_west" {
  role_name     = materialize_role.role_1_us_west.name
  privilege     = "USAGE"
  type_name     = materialize_type.list_type_us_west.name
  schema_name   = materialize_type.list_type_us_west.schema_name
  database_name = materialize_type.list_type_us_west.database_name
  region        = "aws/us-west-2"
}

resource "materialize_type_grant_default_privilege" "type_grant_default_privilege" {
  grantee_name     = materialize_role.role_1.name
  privilege        = "USAGE"
  target_role_name = materialize_role.role_2.name
}

# Create in separate region
resource "materialize_type_grant_default_privilege" "type_grant_default_privilege_us_west" {
  grantee_name     = materialize_role.role_1_us_west.name
  privilege        = "USAGE"
  target_role_name = materialize_role.role_2_us_west.name
  region           = "aws/us-west-2"
}

resource "materialize_type_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target.name
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
}

# Create in separate region
resource "materialize_type_grant_default_privilege" "example_us_west" {
  grantee_name     = materialize_role.grantee_us_west.name
  privilege        = "USAGE"
  target_role_name = materialize_role.target_us_west.name
  schema_name      = materialize_schema.schema_us_west.name
  database_name    = materialize_database.database_us_west.name
  region           = "aws/us-west-2"
}

output "qualified_type" {
  value = materialize_type.list_type.qualified_sql_name
}

data "materialize_type" "all" {}
