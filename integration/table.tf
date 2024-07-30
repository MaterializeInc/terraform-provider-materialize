resource "materialize_table" "simple_table" {
  name          = "simple_table"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  comment       = "table comment"

  column {
    name = "column_1"
    type = "text"
  }
  column {
    name    = "column_2"
    type    = "int"
    comment = "column_2 comment"
  }
  column {
    name     = "column_3"
    type     = "text"
    nullable = true
  }
  column {
    name    = "column_4"
    type    = "text"
    default = "NULL"
  }
  column {
    name     = "column_5"
    type     = "text"
    nullable = true
    default  = "NULL"
  }
}

# Create in separate region
resource "materialize_table" "simple_table_us_west" {
  name          = "simple_table"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  comment       = "table comment"
  region        = "aws/us-west-2"

  column {
    name = "column_1"
    type = "text"
  }
  column {
    name    = "column_2"
    type    = "int"
    comment = "column_2 comment"
  }
  column {
    name     = "column_3"
    type     = "text"
    nullable = true
  }
  column {
    name    = "column_4"
    type    = "text"
    default = "NULL"
  }
  column {
    name     = "column_5"
    type     = "text"
    nullable = true
    default  = "NULL"
  }
}

resource "materialize_table" "simple_table_sink" {
  name          = "simple_table_sink"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  comment       = "table sink comment"

  column {
    name = "key_column"
    type = "text"
  }
  column {
    name = "kafka_header"
    type = "map[text => text]"
  }
  lifecycle {
    ignore_changes = [column]
  }
}

# Create in separate region
resource "materialize_table" "simple_table_sink_us_west" {
  name          = "simple_table_sink"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  comment       = "table sink comment"
  region        = "aws/us-west-2"

  column {
    name = "key_column"
    type = "text"
  }
  column {
    name = "kafka_header"
    type = "map[text => text]"
  }
  lifecycle {
    ignore_changes = [column]
  }
}

resource "materialize_table_grant" "table_grant_select" {
  role_name     = materialize_role.role_1.name
  privilege     = "SELECT"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

# Create in separate region
resource "materialize_table_grant" "table_grant_select_us_west" {
  role_name     = materialize_role.role_1_us_west.name
  privilege     = "SELECT"
  database_name = materialize_table.simple_table_us_west.database_name
  schema_name   = materialize_table.simple_table_us_west.schema_name
  table_name    = materialize_table.simple_table_us_west.name
  region        = "aws/us-west-2"
}

resource "materialize_table_grant" "table_grant_insert" {
  role_name     = materialize_role.role_1.name
  privilege     = "INSERT"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

# Create in separate region
resource "materialize_table_grant" "table_grant_insert_us_west" {
  role_name     = materialize_role.role_1_us_west.name
  privilege     = "INSERT"
  database_name = materialize_table.simple_table_us_west.database_name
  schema_name   = materialize_table.simple_table_us_west.schema_name
  table_name    = materialize_table.simple_table_us_west.name
  region        = "aws/us-west-2"
}

resource "materialize_table_grant" "table_grant_update" {
  role_name     = materialize_role.role_2.name
  privilege     = "UPDATE"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

# Create in separate region
resource "materialize_table_grant" "table_grant_update_us_west" {
  role_name     = materialize_role.role_2_us_west.name
  privilege     = "UPDATE"
  database_name = materialize_table.simple_table_us_west.database_name
  schema_name   = materialize_table.simple_table_us_west.schema_name
  table_name    = materialize_table.simple_table_us_west.name
  region        = "aws/us-west-2"
}

resource "materialize_table_grant" "table_grant_delete" {
  role_name     = materialize_role.role_2.name
  privilege     = "DELETE"
  database_name = materialize_table.simple_table.database_name
  schema_name   = materialize_table.simple_table.schema_name
  table_name    = materialize_table.simple_table.name
}

# Create in separate region
resource "materialize_table_grant" "table_grant_delete_us_west" {
  role_name     = materialize_role.role_2_us_west.name
  privilege     = "DELETE"
  database_name = materialize_table.simple_table_us_west.database_name
  schema_name   = materialize_table.simple_table_us_west.schema_name
  table_name    = materialize_table.simple_table_us_west.name
  region        = "aws/us-west-2"
}

resource "materialize_table_grant_default_privilege" "example" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "SELECT"
  target_role_name = materialize_role.target.name
  schema_name      = materialize_schema.schema.name
  database_name    = materialize_database.database.name
}

# Create in separate region
resource "materialize_table_grant_default_privilege" "example_us_west" {
  grantee_name     = materialize_role.grantee_us_west.name
  privilege        = "SELECT"
  target_role_name = materialize_role.target_us_west.name
  schema_name      = materialize_schema.schema_us_west.name
  database_name    = materialize_database.database_us_west.name
  region           = "aws/us-west-2"
}

resource "materialize_table_grant_default_privilege" "example_all" {
  grantee_name     = materialize_role.grantee.name
  privilege        = "SELECT"
  target_role_name = materialize_role.target.name
}

# Create in separate region
resource "materialize_table_grant_default_privilege" "example_all_us_west" {
  grantee_name     = materialize_role.grantee_us_west.name
  privilege        = "SELECT"
  target_role_name = materialize_role.target_us_west.name
  region           = "aws/us-west-2"
}

output "qualified_table" {
  value = materialize_table.simple_table.qualified_sql_name
}

data "materialize_table" "all" {}
