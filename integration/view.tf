resource "materialize_view" "simple_view" {
  name          = "simple_view"
  schema_name   = materialize_schema.schema.name
  database_name = materialize_database.database.name
  comment       = "view comment"

  statement = <<SQL
SELECT
    1 AS id
SQL

}

# Create in separate region
resource "materialize_view" "simple_view_us_west" {
  name          = "simple_view"
  schema_name   = materialize_schema.schema_us_west.name
  database_name = materialize_database.database_us_west.name
  comment       = "view comment"
  region        = "aws/us-west-2"

  statement = <<SQL
SELECT
    1 AS id
SQL

}

resource "materialize_view_grant" "database_view_select" {
  role_name     = materialize_role.role_1.name
  privilege     = "SELECT"
  view_name     = materialize_view.simple_view.name
  schema_name   = materialize_view.simple_view.schema_name
  database_name = materialize_view.simple_view.database_name
}

# Create in separate region
resource "materialize_view_grant" "database_view_select_us_west" {
  role_name     = materialize_role.role_1_us_west.name
  privilege     = "SELECT"
  view_name     = materialize_view.simple_view_us_west.name
  schema_name   = materialize_view.simple_view_us_west.schema_name
  database_name = materialize_view.simple_view_us_west.database_name
  region        = "aws/us-west-2"
}

output "qualified_view" {
  value = materialize_view.simple_view.qualified_sql_name
}

data "materialize_view" "all" {}
